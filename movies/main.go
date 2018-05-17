package main

import "net/http"
import "github.com/gorilla/mux"
import "fmt"
import "log"

// AllMoviesEndPoint show all movies
func AllMoviesEndPoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "not implemented yet")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/movies", AllMoviesEndPoint).Methods("GET")

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}
}
