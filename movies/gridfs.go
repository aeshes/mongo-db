package main

import "gopkg.in/mgo.v2"
import "errors"
import "log"
import "fmt"

type Storage struct {
	session *mgo.Session
}

var db *mgo.Database

// Connect connects to the default server and opens default DB
func (s *Storage) Connect() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB("binary")
}

// CreateGridFile creates file in GridFS
func (s *Storage) CreateGridFile(name string) (*mgo.GridFile, error) {
	file, err := db.GridFS("fs").Create(name)
	if err != nil {
		return nil, errors.New("Can not create Grid file")
	}
	return file, nil
}

func (s *Storage) UploadGridFile() {
	file, err := s.CreateGridFile("hello")
	if err != nil {
		log.Fatal(err)
	}
	file.Write([]byte("hello world"))
	file.Close()
}

func (s *Storage) ShowGridFile(name string) {
	file, err := db.GridFS("fs").Open(name)
	if err != nil {
		log.Fatal(err)
	}
	buffer := make([]byte, 256)
	file.Read(buffer)
	fmt.Println(string(buffer))
	file.Close()
}

func (s *Storage) WriteToGridFile() {

}
