package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	sts "github.com/broluwo/Scoutmaster9000/structs" // Renaming structs to sts for convenience
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

const (
	mongoDefaultURI = "127.0.0.1:27017"
)

var (
	routes = []sts.Route{
		sts.Route{"/user/{name:[a-z]+}", userHandler, []string{"GET", "POST"}},
		sts.Route{"/team/{teamNum:[0-9]+}", teamHandler, []string{"GET", "POST"}},
	}
	session *mgo.Session
)

//mongo default on 27017
func main() {
	var err error
	session, err = mgo.Dial(mongoDefaultURI)
	if err != nil {
		log.Printf("Can't find mongodb, %v\n", err)
		os.Exit(3)
	}
	defer session.Close()
	// Ensure that any query that changes data is processed without error
	//Set to nil for faster throughput but no error checking
	session.SetSafe(&mgo.Safe{})

	dummyWrite("scoutServer", "team")
	http.Handle("/", initHandlers())
	log.Println("Listening...")
	http.ListenAndServe(":9000", nil)

}
func setupDB() {
	var err error
	session, err = mgo.Dial(mongoDefaultURI)
	if err != nil {
		log.Printf("Can't find Mongodb.\n Ensure that it is running and you have the correct address., %v\n", err)
		os.Exit(3)
	}
	defer session.Close()
	// Ensure that any query that changes data is processed without error
	//Set to nil for faster throughput but no error checking
	session.SetSafe(&mgo.Safe{})

}
func dummyWrite(dbName string, collectionName string) {
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
func Write(collection *mgo.Collection, data ...interface{}) {

}

func initHandlers() *mux.Router {
	router := mux.NewRouter()
	for _, value := range routes {
		router.HandleFunc(value.Route, value.Handler).Methods(value.Methods...)
	}
	return router
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
}

func teamHandler(w http.ResponseWriter, req *http.Request) {
	// params := mux.Vars(req)
	// name := params["teamNum"]

	w.Write([]byte("Hello " + dummyRead("scoutServer", "team")))
}

func dummyRead(dbName string, collectionName string) string {
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
	fmt.Println(result)
	return result.Name
}

func userHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	name := params["name"]
	w.Write([]byte("Hello " + name))
}
