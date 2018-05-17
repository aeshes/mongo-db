package dao

import "gopkg.in/mgo.v2"
import "gopkg.in/mgo.v2/bson"
import "log"
import "movies/models"

type MoviesDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

// Collection is the default collection to store movies in
const Collection = "movies"

// Connect is used to establish connection with database server
func (m *MoviesDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

//FindAll returns all movies from model and an error if occured
func (m *MoviesDAO) FindAll() ([]models.Movie, error) {
	var movies []models.Movie
	err := db.C(Collection).Find(bson.M{}).All(&movies)
	return movies, err
}

// Insert inserts new Movie into Database
func (m *MoviesDAO) Insert(movie models.Movie) error {
	err := db.C(Collection).Insert(&movie)
	return err
}
