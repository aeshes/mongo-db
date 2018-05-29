package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Connect to DB, create repository
	rest := NewRestHandler()

	// Setup routing
	router := mux.NewRouter()

	router.HandleFunc("/commonfs/createWithBuilder/start", func(w http.ResponseWriter, r *http.Request) {
		rest.CreateTemp(w, r)
	}).Methods("PUT")

	router.HandleFunc("/commonfs/createWithBuilder/appendBytes", func(w http.ResponseWriter, r *http.Request) {
		rest.AppendBytes(w, r)
	}).Methods("POST")

	router.HandleFunc("/commonfs/createWithBuilder/commit", func(w http.ResponseWriter, r *http.Request) {
		rest.CommitTemp(w, r)
	}).Methods("GET")

	router.HandleFunc("/commonfs/search", func(w http.ResponseWriter, r *http.Request) {
		rest.FindByForeignKey(w, r)
	}).Methods("HEAD")

	router.HandleFunc("/commonfs/{fileid}", func(w http.ResponseWriter, r *http.Request) {
		rest.GetFileMetadata(w, r)
	}).Methods("HEAD")

	router.HandleFunc("/commonfs/{fileid}", func(w http.ResponseWriter, r *http.Request) {
		rest.Download(w, r)
	}).Methods("GET")

	router.HandleFunc("/commonfs/createAtomically", func(w http.ResponseWriter, r *http.Request) {
		rest.CreateAtomically(w, r)
	}).Methods("PUT")

	http.ListenAndServe(":3000", router)
}
