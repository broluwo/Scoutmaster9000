package main

import (
	"log"
	"net/http"

	sts "github.com/broluwo/Scoutmaster9000/structs" // Renaming structs to sts for convenience
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

const (
	mongoDefaultURI = "127.0.0.1"
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
}

var (
	//If a method is not included in the string slice which holds the recognized
	//methods, it will auto 404 it instead of passing in a 405 error. While this
	//still occurs we will put every method in the subrouter Methods call and use
	// a switch to filter out the unnecessary/unsupported methods
	routes = sts.Routes{
		// we have to escape the \w because go tries to inteerpret it as a string
		//literal. \\w means match any word character including letters, numbers,
		//and underscores
		{"/user", "/{name:[\\w]+}", genUserHandler, specUserHandler},
		{"/teams", "/{teamNum:[0-9]+}", genTeamHandler, specTeamHandler},
		//TODO: Remove s from regionals and teams in a bit
		{"/regionals", "/{regionals:[a-zA-z]+}", genRegionalHandler, specRegionalHandler},
	}

	s = Server{}
	//RestMethods that could be used
	RestMethods = []string{"POST", "PUT", "PATCH", "GET", "HEAD", "DELETE", "OPTIONS"}
)

func main() {
	serverInit()
	//Don't close session till end of main block, which doesn't occur
	//until the server itself is killed
	defer s.Session.Close()
	http.Handle("/", s.initHandlers())
	log.Println("Listening...")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func serverInit() {
	s.initDB()
	s.Routes = routes
	s.NotThere = NotFoundHandler{}
	s.dummyWrite("scoutServer", "team")
}

func (s *Server) initDB() {
	s.DBURI = mongoDefaultURI
	var err error
	di := &mgo.DialInfo{
		Addrs:    []string{s.DBURI},
		Direct:   true,
		FailFast: true,
	}
	s.Session, err = mgo.DialWithInfo(di)
	if err != nil {
		log.Fatalf("Can't find Mongodb.\n Ensure that it is running and you have the correct address., %v\n", err)
	}
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

func (s *Server) dummyWrite(dbName string, collectionName string) {
	session := s.Session.Copy()
	collection := session.DB(dbName).C(collectionName)
	document := sts.Team{
		//Team is the struct that represents a team
		Force:  false,
		Number: 449,
		Name:   "The Blair Robot Project",
	}
	index := mgo.Index{
		Key:        []string{"Number"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := collection.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Can't assert index, %v\n", err)
	}
	err = collection.Insert(document)
	if err != nil {
		log.Fatalf("Can't insert document, %v\n", err)
	}
}

func (s *Server) dummyRead(dbName string, collectionName string) string {
	session := s.Session.Copy()
	collection := session.DB(dbName).C(collectionName)
	index := mgo.Index{
		Key:        []string{"Number"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := collection.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Can't assert index, %v\n", err)
	}
	result := sts.Team{}
	err = collection.Find(nil).One(&result)
	if err != nil {
		log.Fatalf("Can't read document, %v\n", err)
	}
	log.Println(result)
	return result.Name
}
