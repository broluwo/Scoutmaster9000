package main

import (
	"encoding/json"
	"log"
	"net/http"

	sts "github.com/broluwo/Scoutmaster9000/structs"
	"github.com/gorilla/mux"
)

func rootHandler(w http.ResponseWriter, req *http.Request) {}

func specTeamHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":

		w.Write([]byte("Hello " + s.dummyRead("scoutServer", "team")))
		break

	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break
	}
}
func genTeamHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
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
	case "PUT":
		break
	case "PATCH":
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
	case "PUT":
		break
	case "PATCH":
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
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}

//This should respond to get requests as well
func genRegionalHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		//501 error
		break
	case "POST":
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		var r sts.Regional
		err := decoder.Decode(&r)
		if err != nil {
			panic(err)
		}
		log.Println(r)
		http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)
		break
	case "PUT":
		break
	case "PATCH":
		break
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}
