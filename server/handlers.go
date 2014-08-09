package main

import (
	"encoding/json"
	"log"
	"net/http"

	sts "github.com/broluwo/Scoutmaster9000/structs"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

//Gen handlers will respond to: GET and POST; possibly HEAD, OPTIONS
//Spec will respond to: GET,PUT,and PATCH; possibly DELETE, OPTIONS

func rootHandler(w http.ResponseWriter, req *http.Request) {}

func specTeamHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Write([]byte("Hello " + s.dummyRead("scoutServer", "team")))
		break
	case "PUT", "PATCH":
		break

	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break
	}
}
func genTeamHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		break
	case "POST":
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		var t sts.Team
		err := decoder.Decode(&t)
		if err != nil {
			panic(err)
		}
		log.Println(t.Name)
		http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)
		break
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break
	}
}

func specUserHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		params := mux.Vars(req)
		name := params["name"]
		w.Write([]byte("Hello " + name))
		break
	case "PUT", "PATCH":
		break
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}

//This should respond to get requests as well
func genUserHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		break
	case "POST":
		break
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}
func specRegionalHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		params := mux.Vars(req)
		name := params["name"]
		w.Write([]byte("Hello " + name))
		break
	case "PUT", "PATCH":
		//This update will probably be about the match
		break
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}

//This should respond to get requests as well
func (s *Server) genRegionalHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {

	case "GET": //Returns all stored regionals
		regionals, err := SearchRegional(nil, 0, -1)
		if err != nil {
			log.Printf("Couldn't fetch documents, %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Println(regionals)
		ServeJSON(w, regionals)
		break

	case "POST":
		var r sts.Regional
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		// var t sts.Team
		err := decoder.Decode(&r)
		if err != nil {
			panic(err)
		}
		//	ReadJSON(req, r)

		//Check Perms here...
		session := s.getSession()
		collection := session.DB(s.dbName).C("regional")
		if _, err := collection.UpsertId(bson.M{"Location": r.Location, "Year": r.Year}, &r); err != nil {
			log.Printf("Can't insert/update document, %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			session.Close()
			return
		}
		session.Close()
		// log.Printf("%v", changeInfo)
		// index := mgo.Index{
		// 	Key:        []string{"year"},
		// 	Unique:     true,
		// 	DropDups:   true,
		// 	Background: true,
		// 	Sparse:     true,
		// }
		// err = collection.EnsureIndex(index)
		// if err != nil {
		// 	log.Fatalf("Can't assert index, %v\n", err)
		// }
		// err = collection.Insert(r)
		// if err != nil {
		// 	log.Fatalf("Can't insert document, %v\n", err)
		// }
		http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)

		break

	default: //Don't need to use custom 404 handler. can just serve a 405 error from here
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}
