package main

import (
	"log"
	"net/http"
	"os"
	"time"

	sts "github.com/broluwo/Scoutmaster9000/structs" // Renaming structs to sts for convenience
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

const (
	mongoDefaultURI = "127.0.0.1"
	//NotFound is the constant for a 404 page, used in the NotFoundHandler
	NotFound = iota + 404
	//NotSupported == 405
	NotSupported
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
	routes = sts.Routes{
		{"/user/{name:[a-z]+}", userHandler, []string{"GET", "POST"}},
		{"/team/{teamNum:[0-9]+}", teamHandler, []string{"GET", "POST"}},
	}

	s = Server{}
)

func main() {
	serverInit()
	defer s.Session.Close()
	http.Handle("/", s.initHandlers())
	log.Println("Listening...")
	http.ListenAndServe(":9000", nil)

}

func serverInit() {
	s.setupDB()
	s.Routes = routes
	s.NotThere = NotFoundHandler{NotFound}
	s.dummyWrite("scoutServer", "team")
}

func (s *Server) setupDB() {
	s.DBURI = mongoDefaultURI
	var err error
	di := &mgo.DialInfo{
		Addrs:    []string{s.DBURI},
		Direct:   true,
		Timeout:  time.Duration(30 * time.Second),
		FailFast: true,
	}
	s.Session, err = mgo.DialWithInfo(di)
	if err != nil {
		log.Printf("Can't find Mongodb.\n Ensure that it is running and you have the correct address., %v\n", err)
		os.Exit(3)
	}
	// Ensure that any query that changes data is processed without error
	//Set to nil for faster throughput but no error checking
	s.Session.SetSafe(&mgo.Safe{})
	s.Session.SetMode(mgo.Monotonic, true)
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
		log.Printf("Can't assert index, %v\n", err)
		os.Exit(3)
	}
	err = collection.Insert(document)
	if err != nil {
		log.Printf("Can't insert document, %v\n", err)
		os.Exit(3)
	}
}

//Write writes data to the MongoDB instance
//Consider using bulk api
//http://blog.mongodb.org/post/84922794768/mongodbs-new-bulk-api
func (s *Server) Write(collection *mgo.Collection, data ...interface{}) {}

func (s *Server) initHandlers() *mux.Router {
	router := mux.NewRouter()
	for _, value := range s.Routes {
		router.HandleFunc(value.Route, value.Handler).Methods(value.Methods...)
	}
	router.NotFoundHandler = NotFoundHandler{}
	return router
}

func (p NotFoundHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch p.Method {
	case NotSupported:
		w.Write([]byte("405: Method Not Supported, man"))
		p.Method = NotFound
		break
	default:
		w.Write([]byte("404 page not found, man"))
		break
	}
}

func rootHandler(w http.ResponseWriter, req *http.Request) {}

func teamHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		// params := mux.Vars(req)
		// name := params["teamNum"]
		w.Write([]byte("Hello " + s.dummyRead("scoutServer", "team")))
		break
	case "POST":
		break
	default:
		s.NotThere.Method = NotSupported
		s.NotThere.ServeHTTP(w, req)
		break
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
		log.Printf("Can't assert index, %v\n", err)
		os.Exit(3)
	}
	result := sts.Team{}
	err = collection.Find(nil).One(&result)
	if err != nil {
		log.Printf("Can't read document, %v\n", err)
		os.Exit(3)
	}
	log.Println(result)
	return result.Name
}

func userHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		params := mux.Vars(req)
		name := params["name"]
		w.Write([]byte("Hello " + name))
		break
	case "POST":
		break
	default:
		s.NotThere.Method = NotSupported
		s.NotThere.ServeHTTP(w, req)
		break

	}
}
