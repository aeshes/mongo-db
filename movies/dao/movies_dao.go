package dao

import "gopkg.in/mgo.v2"
import "log"

type MoviesDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const Collection = "movies"

func (m *MoviesDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

func (m *MoviesDAO) Insert(movie Movie) error {
	err := db.C(Collection).Insert(&movie)
	return err
}
