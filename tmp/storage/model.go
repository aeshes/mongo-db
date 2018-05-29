package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

//
// Meta is metadata for persistent and temporary file in gridfs
//
type Meta struct {
	Name    string `bson:"name" json:"name"`
	Creator string `bson:"creator" json:"creator"`
	Hash    string `bson:"hash" json:"hash"`
	SysID   string `bson:"sysId" json:"sysId"`
}

//
// ObjectMeta represents stored object meta information
//
type ObjectMeta struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Filename    string        `bson:"filename" json:"filename"`
	ContentType string        `bson:"contentType" json:"content_type"`
	Size        int64         `bson:"length" json:"size"`
	ChunkSize   int64         `bson:"chunkSize" json:"chunk_size"`
	CheckSum    string        `bson:"md5" json:"md5"`
	CreatedOn   time.Time     `bson:"uploadDate" json:"created_on"`
	Metadata    Meta          `bson:"metadata,omitempty" json:"extra,omitempty"`
}
