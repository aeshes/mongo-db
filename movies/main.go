package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"movies/dao"
	"movies/models"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

var d = dao.MoviesDAO{}

func TestingEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Println(r.Header["Content-Range"])
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	file, err := os.OpenFile("tmp.jpg", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
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
