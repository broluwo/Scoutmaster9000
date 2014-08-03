package main

import (
	"log"
	"net/http"

	sts "github.com/broluwo/Scoutmaster9000/structs" // Renaming structs to sts for faster typing
	"github.com/gorilla/mux"
)

var (
	routes = []sts.Route{
		sts.Route{"/user/{name:[a-z]+}", userHandler, []string{"GET", "POST"}},
		sts.Route{"/team/{teamNum:[0-9]+}", teamHandler, []string{"GET", "POST"}},
	}
)

func main() {
	http.Handle("/", initHandlers())
	log.Println("Listening...")
	http.ListenAndServe(":9000", nil)
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
	log.Println("We out here.")
	params := mux.Vars(req)
	name := params["teamNum"]
	w.Write([]byte("Hello " + name))
}
func userHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	name := params["name"]
	w.Write([]byte("Hello " + name))
}
