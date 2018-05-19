package main

import (
	"encoding/json"
	"fmt"
	"log"
	"movies/dao"
	"movies/models"
	"movies/upload"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

var d = dao.MoviesDAO{}
var storage = Storage{}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func TestingEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	meta, err := upload.ParseMeta(r)
	check(err)

	fmt.Println(meta.MediaType)
}

func fileUploadEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

}

// AllMoviesEndPoint show all movies
func AllMoviesEndPoint(w http.ResponseWriter, r *http.Request) {
	movies, err := d.FindAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, movies)
}

// CreateMoviesEndPoint inserts new movie into BD
func CreateMoviesEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var movie models.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	movie.ID = bson.NewObjectId()
	if err := d.Insert(movie); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, movie)
}

func UpdateMovieEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var movie models.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := d.Update(movie); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func init() {
	d.Server = "127.0.0.1"
	d.Database = "movies"
}

func main() {
	storage.Connect()
	storage.UploadGridFile()
	storage.ShowGridFile("hello")

	d.Connect()
	router := mux.NewRouter()
	router.HandleFunc("/movies", AllMoviesEndPoint).Methods("GET")
	router.HandleFunc("/movies", CreateMoviesEndPoint).Methods("POST")
	router.HandleFunc("/movies", UpdateMovieEndPoint).Methods("PUT")
	router.HandleFunc("/testing", TestingEndpoint).Methods("PUT")

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}
}
