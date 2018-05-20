package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var storage = Storage{}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func testingEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	meta, err := ParseMeta(r)
	check(err)

	fmt.Println(meta.FileName)

	file, err := createTempFile("temp.jpg")
	check(err)
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		log.Fatal(err)
	}
}

func createTempFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0600)
	check(err)

	return file, nil
}

func getFileEndpoint(w http.ResponseWriter, r *http.Request) {
	file, err := storage.OpenFile("hello")
	if err != nil {
		respondWithError(w, 404, "grid file not found")
		return
	}
	defer file.Close()
	if _, err := io.Copy(w, file); err != nil {
		respondWithError(w,
			http.StatusInternalServerError,
			"error while reading grid file")
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w,
		code,
		map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	storage.Connect()

	router := mux.NewRouter()
	router.HandleFunc("/testing", testingEndpoint).Methods("PUT")
	router.HandleFunc("/testing", getFileEndpoint).Methods("GET")

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}
}