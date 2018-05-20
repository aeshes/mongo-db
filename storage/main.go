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
var local = LocalStorage{}

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

// HEAD request
func handleFileID(w http.ResponseWriter, r *http.Request) {
	meta, err := ParseMeta(r)
	if err != nil {
		log.Println(err)
		return
	}

	if meta.Property.isValid() {
		fmt.Printf("Name = %s\n", meta.Property.Name)
		fmt.Printf("Creator = %s\n", meta.Property.Creator)
		fmt.Printf("Hash = %s\n", meta.Property.Hash)
		fmt.Printf("sysId = %s\n", meta.Property.SysID)
	} else {
		respondWithError(w,
			404,
			"empty user defined headers")
	}
}

// PUT request
//Atomically creates "small" file which can be POSTed in one request
func handleCreateAtomically(w http.ResponseWriter, r *http.Request) {
	meta, err := ParseMeta(r)
	if err != nil {
		log.Println("In handleCreateAtomically: ", err)
		respondWithError(w, 400, "cant parse meta info")
		return
	}

	if meta.Property.isValid() {
		file, err := local.CreateTempFile(meta.Property.Name)
		if err != nil {
			log.Println("When CreateAtomically: create temporary, ", err)
			respondWithError(w, 400, "cant create local file")
			return
		}

		defer file.Close()

		if _, err := io.Copy(file, r.Body); err != nil {
			log.Println("When CreateAtomically, write file: ", err)
			respondWithError(w, 400, "cant write local file")
		}
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
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
	router.HandleFunc("/commonfs/{fileid}", handleFileID).Methods("HEAD")
	router.HandleFunc("/commonfs/createAtomically", handleCreateAtomically).Methods("PUT")

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}
}
