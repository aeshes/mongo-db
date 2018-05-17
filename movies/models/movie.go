package models

import "gopkg.in/mgo.v2/bson"

type Movie struct {
	ID          bson.ObjectId
	Name        string
	CoverImage  string
	Description string
}
