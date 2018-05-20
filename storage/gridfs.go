package main

import "gopkg.in/mgo.v2"
import "errors"
import "log"
import "fmt"
import "io"
import "os"

type Storage struct {
	Server     string
	Database   string
	Collection string
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

// StoreFromDisk stores disk file in GridFS
// Is local file's sha-256 is not equal to sha-256 value in FileMeta,
// error is returned
func (s *Storage) StoreFromDisk(file *LocalFile, meta *FileMeta) error {
	if file.Sha256() == meta.Hash {
		gridFile, err := s.CreateGridFile(meta.Name)
		if err != nil {
			log.Println("In StoreFromDisk: ", err)
			return err
		}
		defer gridFile.Close()

		localFile, err := os.Open(file.Path)
		if err != nil {
			log.Println("While opening local file in StoreFromDisk: ", err)
			return err
		}
		defer localFile.Close()

		bytesWritten, err := io.Copy(gridFile, localFile)
		if err != nil {
			log.Println("While copying local file to GridFS: ", err)
			return err
		}
		log.Printf("Copied %d bytes to GridFS.", bytesWritten)
	}

	return errors.New("file.sha256 != meta.sha256")
}

func (s *Storage) UploadGridFile() {
	file, err := s.CreateGridFile("hello")
	if err != nil {
		log.Fatal(err)
	}
	file.Write([]byte("hello world"))
	file.Close()
}

func (s *Storage) OpenFile(name string) (io.ReadCloser, error) {
	file, err := db.GridFS("fs").Open(name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return file, nil
}

func (s *Storage) SaveFileToDisk(name string) {
	// opens file in mongo GridFS
	file, err := db.GridFS("fs").Open(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	dest, err := os.OpenFile("tmp",
		os.O_CREATE|os.O_WRONLY,
		0644)
	defer dest.Close()

	// Copies from grid file to disk file
	if _, err := io.Copy(dest, file); err != nil {
		fmt.Println(err)
	}
}

func (s *Storage) WriteToGridFile() {

}
