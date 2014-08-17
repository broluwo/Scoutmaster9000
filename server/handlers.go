package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	sts "github.com/broluwo/Scoutmaster9000/structs"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

//A lot more copy pasta than I'd like, need to generalize some of these patterns

//No need to sanitize inputted strings being stored in mongo as it's being
//stored in a binary format (BSON) anyway.

//Gen handlers will respond to: GET and POST; possibly HEAD, OPTIONS
//Spec will respond to: GET,PUT,and PATCH; possibly DELETE, OPTIONS

/*
Upon using an UPDATE method there are a series of steps that need to occur:
	0. Check whether or not they are allowed to do so.
	1. Check that the object exists
	2. Update the object with the provided data ensuring it doesn't invalidate
		 what was already there.
	3. Return the appropriate Response Code
*/
//TODO:What should this do? Perhaps announce that the api server is actually on
//this port? Perhaps return a StatusNoContent(204) or a StatusFound(300)?
//TODO: 405 errors need to write the ALLOW header
func rootHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "The Scoutmaster9000 API Server resides on this port.",
		http.StatusNoContent)
}

func specTeamHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	//Relatively easy to support multiple queries,
	//Would have to change the path regex, and then when we get the teamnumber
	// we would split on commas. send the search as go funcs, have a select
	//waiting for responses, append to an array and serve the array
	case "GET":
		teamNum, err := parseTeamNum(req)
		if err != nil {
			//This error should never occur as the regex condition should catch it.
			//We may want to change that behaviour.
			http.Error(w, "That's not a parseable int. Can't find the team.",
				http.StatusBadRequest)
			return
		}
		log.Println(teamNum)
		teams, e := SearchByTeamNum(teamNum, 0, 1)
		//Doesn't return error if no results are found as the search didn't fail,
		//it just found nothing.
		if e != nil || len(teams) == 0 {
			http.Error(w, "Team Couldn't Be Found.",
				http.StatusBadRequest)
			return
		}
		log.Println(teams)
		ServeJSON(w, teams)
		break
	case "PUT", "PATCH":
		teamNum, err := parseTeamNum(req)
		if err != nil {
			http.Error(w, "That's not a parseable int. Can't find the team.",
				http.StatusBadRequest)
			return
		}
		log.Println(teamNum)
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
		teams, err := SearchTeam(nil, 0, -1)
		if err != nil {
			log.Printf("Couldn't fetch documents, %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		log.Println(teams)
		ServeJSON(w, teams)
		break
	case "POST":
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		var t sts.Team
		err := decoder.Decode(&t)
		if err != nil {
			//TODO: Actually handle this correctly. There's absolutely no reason to
			//panic in almost any situation
			panic(err)
		}
		// I''m here trying to implement the actual storing of team data. This error
		// should stop me from being able to move on without addressing this.

		//Check Perms here...
		session := s.getSession()
		collection := session.DB(s.dbName).C("team")
		if _, err := collection.UpsertId(bson.M{"Number": t.Number}, &t); err != nil {
			log.Printf("Can't insert/update document, %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
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
		log.Println(t.Name)
		http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)
		break
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break
	}
}

func parseTeamNum(req *http.Request) (teamNum int, err error) {
	params := mux.Vars(req)
	teamNum, err = strconv.Atoi(params["teamNum"])
	return
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
		year, err := strconv.Atoi(params["year"])

		if err != nil {
			//This error should never occur as the regex condition should catch it.
			//We may want to change that behaviour.
			http.Error(w, "That's not a parseable int. Can't find the regional.",
				http.StatusBadRequest)
			return
		}

		regional := params["regionalName"]
		log.Println(regional)
		log.Println(year)
		//if the regional var is empty, return all regionals for the year
		break
	case "PUT", "PATCH":
		//This update will probably be about a match, might be helpful
		break
	default:
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}

func genRegionalHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {

	case "GET": //Returns all stored regionals
		regionals, err := SearchRegional(nil, 0, -1)
		if err != nil {
			log.Printf("Couldn't fetch documents, %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
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
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
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

	default: //TODO:Don't need to use custom 404 handler. can just serve a 405 error from here
		s.NotThere.Method = http.StatusMethodNotAllowed
		s.NotThere.ServeHTTP(w, req)
		break

	}
}
