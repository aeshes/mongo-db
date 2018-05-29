package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const MongoDefaultDB = "main"
const TmpDir = "./tmp/"

type RestHandler struct {
	session *mgo.Session
	gridFS  *mgo.GridFS
}

//
// NewMongoDB creates new mongo connection handler
//
func NewRestHandler() *RestHandler {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		return nil
	}
	handler := new(RestHandler)
	handler.session = session
	handler.gridFS = session.DB(MongoDefaultDB).GridFS("fs")
	return handler
}

func ParseMeta(r *http.Request) Meta {
	return Meta{
		Name:    r.Header.Get("name"),
		Creator: r.Header.Get("creator"),
		SysID:   r.Header.Get("sysId"),
		Hash:    r.Header.Get("hash"),
	}
}

/*
5. PUT
/commonfs/createWithBuilder/start

Request headers:
<creator> : creator property
<sysId>: sysId property
<name> : file name


Request body:
none

Response status:
200 - success, 400 - error

Response headers:
<fileTempId> : fileTempId property
*/
func (handler RestHandler) CreateTemp(w http.ResponseWriter, r *http.Request) {
	meta := ParseMeta(r)

	gridTempID := handler.createTempGridFile(meta)
	if gridTempID == "" {
		w.WriteHeader(400)
		return
	}

	if err := handler.createTempFile(gridTempID); err != nil {
		w.WriteHeader(400)
		return
	}

	w.Header().Add("fileTempId", gridTempID)
	w.WriteHeader(200)
}

func (handler RestHandler) createTempGridFile(meta Meta) string {
	fs := handler.session.DB(MongoDefaultDB).GridFS("fs")

	gridFile, err := fs.Create(meta.Name)
	if err != nil {
		return ""
	}
	defer gridFile.Close()

	objMeta := ObjectMeta{}
	objMeta.Filename = meta.Name
	objMeta.Metadata = meta

	gridFile.SetName(objMeta.Filename)
	gridFile.SetMeta(objMeta.Metadata)

	return gridFile.Id().(bson.ObjectId).Hex()
}

func (handler RestHandler) createTempFile(name string) error {
	file, err := os.Create(TmpDir + name)
	file.Close()
	return err
}

/*
/*
6. POST
/commonfs/createWithBuilder/appendBytes

Request headers:
<fileTempId> : fileTempId property


Request body:
file body

Response status:
200 - success, 400 - error

Response headers:
none
*/
func (handler RestHandler) AppendBytes(w http.ResponseWriter, r *http.Request) {
	fileTempID := r.Header.Get("fileTempId")
	if fileTempID == "" {
		w.WriteHeader(400)
		return
	}

	file, err := os.OpenFile(TmpDir+fileTempID,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0600)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
}

/*
7. GET
/commonfs/createWithBuilder/commit

Request headers:
<fileTempId> : fileTempId property
<hash> : wanted hash property

Request body:
none

Response status:
200 - success, 400 - error

Response headers:
<fileId> : fileId property
*/
func (handler RestHandler) CommitTemp(w http.ResponseWriter, r *http.Request) {
	fileTempID := r.Header.Get("fileTempId")
	hash := r.Header.Get("hash")

	// Check hashes equal
	if !handler.checkPolicy(fileTempID, hash) {
		w.WriteHeader(400)
		return
	}

	// Check is fileTempID a valid object ID
	if !bson.IsObjectIdHex(fileTempID) {
		w.WriteHeader(400)
		return
	}

	fs := handler.session.DB(MongoDefaultDB).GridFS("fs")
	objID := bson.ObjectIdHex(fileTempID)

	// Get old metadata by ID and save
	oldMeta, _ := handler.getOldMetaByID(objID)
	oldMeta.Hash = hash
	objMeta := ObjectMeta{}
	objMeta.Filename = oldMeta.Name
	objMeta.Metadata = oldMeta

	gridFile, err := fs.Create(objMeta.Filename)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	defer gridFile.Close()

	file, err := os.Open(TmpDir + fileTempID)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	defer file.Close()

	if _, err := io.Copy(gridFile, file); err != nil {
		w.WriteHeader(400)
		return
	}

	gridFile.SetName(objMeta.Filename)
	gridFile.SetMeta(oldMeta)

	fileID := gridFile.Id().(bson.ObjectId).Hex()
	w.Header().Add("fileId", fileID)
	w.WriteHeader(200)

	// remove temp entry
	handler.removeById(objID)
}

func (handler RestHandler) getOldMetaByID(id bson.ObjectId) (Meta, error) {
	meta := Meta{}
	fs := handler.session.DB(MongoDefaultDB).GridFS("fs")
	gridFile, err := fs.OpenId(id)
	if err != nil {
		return meta, err
	}
	defer gridFile.Close()

	gridFile.GetMeta(&meta)
	return meta, nil
}

func (handler RestHandler) checkPolicy(tempID, wantedHash string) bool {
	path := TmpDir + tempID
	return strings.EqualFold(CalcSha256(path), wantedHash)
}

// CalcSha256 Calculates sha-256 checksum of this file
func CalcSha256(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return ""
	}

	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Println(err)
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%x", hasher.Sum(nil))
	return b.String()
}

func (handler RestHandler) removeById(id bson.ObjectId) error {
	return handler.session.DB(MongoDefaultDB).GridFS("fs").RemoveId(id)
}

/*
3. GET

/commonfs/<fileid>
Request headers:
none

Response status:
200 - presents, 404 - absents

Response headers:
<name> : file name
<hash> : file hash
<creator> : creator property
<sysId>: sysId property

Response body:
file body
*/
func (handler RestHandler) Download(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["fileid"]
	if !bson.IsObjectIdHex(fileID) {
		w.WriteHeader(404)
		return
	}

	objID := bson.ObjectIdHex(fileID)
	fs := handler.session.DB(MongoDefaultDB).GridFS("fs")

	// Open GridFile by id
	file, err := fs.OpenId(objID)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	defer file.Close()

	meta := handler.getGridFileMetadata(file)

	w.Header().Add("name", meta.Name)
	w.Header().Add("creator", meta.Creator)
	w.Header().Add("hash", meta.Hash)
	w.Header().Add("sysId", meta.SysID)

	if file.ContentType() != "" {
		w.Header().Set("Content-Type", file.ContentType())
	}

	if file.Name() != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", file.Name()))
	}

	if _, err := io.Copy(w, file); err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
}

// For opened grid file handler returns its metadata
func (handler RestHandler) getGridFileMetadata(gridFile *mgo.GridFile) Meta {
	meta := Meta{}
	gridFile.GetMeta(&meta)
	return meta
}

/*
1. HEAD

/commonfs/<fileid>
Request headers:
none

Response status:
200 - presents, 404 - absents

Response headers:
<name> : file name
<hash> : file hash
<creator> : creator property
<sysId>: sysId property

Response body:
none
*/
func (handler RestHandler) GetFileMetadata(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["fileid"]
	gridFile, err := handler.openGridFileByID(fileID)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	defer gridFile.Close()

	meta := handler.getGridFileMetadata(gridFile)
	w.Header().Add("name", meta.Name)
	w.Header().Add("hash", meta.Hash)
	w.Header().Add("creator", meta.Creator)
	w.Header().Add("sysId", meta.SysID)
	w.WriteHeader(200)
}

func (handler RestHandler) openGridFileByID(fileID string) (*mgo.GridFile, error) {
	if !bson.IsObjectIdHex(fileID) {
		return nil, errors.New("Incorrect fileID")
	}

	objID := bson.ObjectIdHex(fileID)
	return handler.session.DB(MongoDefaultDB).GridFS("fs").OpenId(objID)
}

/*
2. HEAD
/commonfs/search

Request headers:
<creator> : creator property
<sysId>: sysId property

Response status:
200 - presents, 404 - absents

Response headers:
<fileId> : fileId property
*/
func (handler RestHandler) FindByForeignKey(w http.ResponseWriter, r *http.Request) {
	creator := r.Header.Get("creator")
	sysID := r.Header.Get("sysId")
	fileID := handler.findByKey(creator, sysID)
	if fileID == "" {
		w.WriteHeader(404)
		return
	}
	w.Header().Add("fileId", fileID)
	w.WriteHeader(200)
}

func (handler RestHandler) findByKey(creator, sysID string) string {
	result := ObjectMeta{}
	fs := handler.session.DB(MongoDefaultDB).GridFS("fs")
	fs.Find(bson.M{"metadata.creator": creator, "metadata.sysId": sysID}).One(&result)
	return result.ID.Hex()
}

/*
4. PUT
/commonfs/createAtomically

Request headers:
<creator> : creator property
<sysId>: sysId property
<name> : file name
<hash> : wanted hash property

Request body:
file body

Response status:
200 - success, 400 - error

Response headers:
<fileId> : fileId property

Response body:
none
*/
func (handler RestHandler) CreateAtomically(w http.ResponseWriter, r *http.Request) {
	meta := ParseMeta(r)

	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	defer r.Body.Close()

	// Check SHA-256
	checksum := digest(body)
	if !strings.EqualFold(checksum, meta.Hash) {
		w.WriteHeader(400)
		return
	}

	// Create Grid File
	file, err := handler.gridFS.Create(meta.Name)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	defer file.Close()

	if _, err := file.Write(body); err != nil {
		w.WriteHeader(400)
		return
	}

	file.SetName(meta.Name)
	file.SetMeta(meta)

	w.Header().Add("fileId", file.Id().(bson.ObjectId).Hex())
	w.WriteHeader(200)
}

func digest(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
