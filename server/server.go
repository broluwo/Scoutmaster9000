package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	sts "github.com/broluwo/Scoutmaster9000/structs" // Renaming structs to sts for convenience
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//We could store all the essential variables in the path variable
const (
	mongoDefaultURI = "127.0.0.1"
	dbName          = "scoutServer"
)

//NotFoundHandler for 404 and 405 errors depending on the switching of the Method value
type NotFoundHandler struct {
	Method int
}

//Server ...
type Server struct {
	Session  *mgo.Session
	DBURI    string
	Routes   sts.Routes
	NotThere NotFoundHandler
	dbName   string
}

//If a method is not included in the string slice which holds the recognized
//methods, it will auto 404 it instead of passing in a 405 error. While this
//still occurs we will put every method in the subrouter Methods call and use
// a switch to filter out the unnecessary/unsupported methods
var (
	routes = sts.Routes{
		// we have to escape the \w because go tries to interperet it as a string
		//literal. \w means match any word character including letters, numbers,
		//and underscores
		{"/user", "/{name:[\\w]+}", genUserHandler, specUserHandler},
		{"/team", "/{teamNum:[0-9]+}", genTeamHandler, specTeamHandler},
		//TODO:Add a route for just the year
		{"/regional", "/{year:[0-9]+}/{regionalName:[a-zA-z]+}", genRegionalHandler, specRegionalHandler},
	}

	s = Server{}

	//RestMethods that could be used
	RestMethods = []string{"POST", "PUT", "PATCH", "GET", "HEAD", "DELETE", "OPTIONS"}
)

func main() {
	initServer()
	//Don't close session till end of main block, which doesn't occur
	//until the server itself is killed
	defer s.Session.Close()
	http.Handle("/", s.initHandlers())
	log.Println("Listening...")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func initServer() {
	s.initDB()
	s.Routes = routes
	s.NotThere = NotFoundHandler{}
}

func (s *Server) initDB() {
	s.DBURI = mongoDefaultURI
	s.dbName = dbName
	s.getSession()
	// Ensure that any query that changes data is processed without error
	//Set to nil for faster throughput but no error checking
	s.Session.SetSafe(&mgo.Safe{})
	s.Session.SetMode(mgo.Monotonic, true)
}

//Write writes data to the MongoDB instance
//Consider using bulk api
//http://blog.mongodb.org/post/84922794768/mongodbs-new-bulk-api
func (s *Server) Write(collection *mgo.Collection, data ...interface{}) {}

//Query queries data from DB
func (s *Server) Query(collection *mgo.Collection) {}

func (s *Server) initHandlers() *mux.Router {
	r := mux.NewRouter()
	//Forces the router to recognize /path and /path/ as the same.
	//Commented out because it returns a 301 Perm Redirect, and i haven't found a
	//good way(non hackish) to handle it.
	// r.StrictSlash(true)
	for _, value := range s.Routes {
		router := r.PathPrefix(value.PrefixRoute).Subrouter()
		router.HandleFunc("/", value.PrefixHandler).Methods(RestMethods...).Name(value.PrefixRoute)
		router.HandleFunc(value.PostfixRoute, value.PostfixHandler).Methods(RestMethods...).Name(value.PostfixRoute)
	}
	r.NotFoundHandler = s.NotThere
	return r
}

func (p NotFoundHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch p.Method {
	case http.StatusMethodNotAllowed:
		http.Error(w, "That Method isn't allowed on this resource", http.StatusMethodNotAllowed)
		p.Method = http.StatusNotFound
		break
	default: //Defaulted because on a true 404, mux returns an empty string.
		http.NotFound(w, req)
		break
	}
}

// ServeJSON replies to the request with a JSON
// representation of resource v.
func ServeJSON(w http.ResponseWriter, v interface{}) {
	// avoid json vulnerabilities, always wrap v in an object literal
	//	doc := map[string]interface{}{"d": v}
	if data, err := json.Marshal(v); err != nil {
		log.Printf("Error marshalling json: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

// ReadJSON decodes JSON data into a provided struct
//Could do it the json decoder way if i defined a unmarshaller for each type
//Use this then do a tyoe assertion
func ReadJSON(req *http.Request, v interface{}) error {
	//	var k json.RawMessage
	defer req.Body.Close()
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&v)
	return err
}

//Use this method to debug things
func logRequest(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var s = time.Now()
		handler(w, r)
		log.Printf("%s %s %6.3fms", r.Method, r.RequestURI, (time.Since(s).Seconds() * 1000))
	}
}
func (s *Server) getSession() *mgo.Session {
	if s.Session == nil {
		var err error
		di := &mgo.DialInfo{
			Addrs:    []string{s.DBURI},
			Direct:   true,
			FailFast: true, //You may want to turn this off if you're expecting latency
		}
		s.Session, err = mgo.DialWithInfo(di)
		if err != nil {
			log.Fatalf("Can't find Mongodb.\n Ensure that it is running and you have the correct address., %v\n", err)
		}
	}
	//If you also want to reuse the socket, use clone instead
	return s.Session.Copy()
}

//WithCollection takes the name of the collection, along with a function
//that expects the connection object to that collection,
//and can execute access functions on it.
func withCollection(collection string, fn func(*mgo.Collection) error) error {
	session := s.getSession()
	defer session.Close()
	c := session.DB(s.dbName).C(collection)
	return fn(c)
}

//SearchTeam is a generic form for searching for a Team
//Set skip to zero is you want all the results,
//Set limit to < 0  if you want all the results
//Naming the results allows us to not have to return them
func SearchTeam(q interface{}, skip int, limit int) (searchResults []sts.Team, err error) {
	searchResults = []sts.Team{}
	query := func(c *mgo.Collection) error {
		fn := c.Find(q).Skip(skip).Limit(limit).All(&searchResults)
		if limit < 0 {
			fn = c.Find(q).Skip(skip).All(&searchResults)
		}
		return fn
	}
	search := func() error {
		return withCollection("team", query)
	}
	err = search()
	return
}

//SearchByTeamNum is a wrapper for
//SearchTeam(bson.M{"Number": teamNum, skip, limit})
func SearchByTeamNum(teamNum int, skip int, limit int) (searchResults []sts.Team, err error) {
	searchResults, err = SearchTeam(bson.M{"Number": teamNum}, skip, limit)
}

//SearchRegional is a generic form for searching for a Regional
//Set skip to zero is you want all the results, set limit to < 0  if you want all the results
//Naming the results allows us to not have to return them
func SearchRegional(q interface{}, skip int, limit int) (searchResults []sts.Regional, err error) {
	searchResults = []sts.Regional{}
	query := func(c *mgo.Collection) error {
		fn := c.Find(q).Skip(skip).Limit(limit).All(&searchResults)
		if limit < 0 {
			fn = c.Find(q).Skip(skip).All(&searchResults)
		}
		return fn
	}

	search := func() error {
		return withCollection("regional", query)
	}
	err = search()
	return
}
