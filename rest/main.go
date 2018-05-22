package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"gopkg.in/mgo.v2/bson"
)

type FileResource struct {
}

func (f *FileResource) RestService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/commonfs").
		Consumes(restful.MIME_JSON, restful.MIME_OCTET).
		Produces(restful.MIME_JSON, restful.MIME_OCTET)

	ws.Route(ws.HEAD("/{file-id}").To(findByID).
		Doc("Find metainfo about file by file id").
		Param(ws.PathParameter("user-id", "identifier of the file").DataType("string")).
		Writes(FileMeta{}))

	ws.Route(ws.HEAD("/search").To(searchFile).
		Doc("Returns fileid if file exists").
		Writes(FileMeta{}).
		Returns(200, "OK", FileMeta{}).
		Returns(400, "Error", nil))

	ws.Route(ws.PUT("/createAtomically").To(createAtomically).
		Doc("Create atomically small file"))

	ws.Route(ws.PUT("/createWithBuilder/start").To(builderStart).
		Doc("Create empty temporary file"))

	return ws
}

var datastore = DataStorage{}

func main() {
	datastore.Connect()

	r := FileResource{}
	restful.DefaultContainer.Add(r.RestService())
	http.ListenAndServe(":3000", nil)
}

func getHTTPFileMeta(request *restful.Request) *FileMeta {
	meta := FileMeta{}
	meta.Name = request.HeaderParameter("name")
	meta.Creator = request.HeaderParameter("creator")
	meta.Hash = request.HeaderParameter("hash")
	meta.SysID = request.HeaderParameter("sysId")

	return &meta
}

/*HEAD

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
none*/

func findByID(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("file-id")
	meta := datastore.QueryMeta(id)
	if meta != nil {
		response.AddHeader("name", meta.Name)
		response.AddHeader("creator", meta.Creator)
		response.AddHeader("hash", meta.Hash)
		response.AddHeader("sysId", meta.SysID)
		response.WriteHeader(http.StatusOK)

		return
	}
	response.WriteError(http.StatusNotFound, errors.New("absents"))
}

/*2. HEAD
/commonfs/search

Request headers:
<creator> : creator property
<sysId>: sysId property

Response status:
200 - presents, 404 - absents

Response headers:
<fileId> : fileId property*/
func searchFile(request *restful.Request, response *restful.Response) {
	creator := request.HeaderParameter("creator")
	sysid := request.HeaderParameter("sysId")

	fileid, err := datastore.QueryFileId(bson.M{"creator": creator, "sysId": sysid})
	if err != nil {
		log.Println("In SearchEndpoint: ", err)
		response.WriteError(http.StatusNotFound, errors.New("absents"))
		return
	}

	response.AddHeader("fileId", fileid)
	response.WriteHeader(http.StatusOK)
}

/*4. PUT
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
none*/
func createAtomically(request *restful.Request, response *restful.Response) {
	name := request.HeaderParameter("name")
	creator := request.HeaderParameter("creator")
	wantedHash := request.HeaderParameter("hash")
	sysID := request.HeaderParameter("sysId")

	if name != "" {
		if file, err := datastore.CreateTempFile(name); err == nil {
			defer file.Close()
			io.Copy(file, request.Request.Body)
			meta := FileMeta{Name: name, Creator: creator, Hash: wantedHash, SysID: sysID}
			id, err := datastore.StoreFromDisk(name, &meta)
			if err == nil {
				fmt.Println(id)
				response.AddHeader("fileId", id)
				response.WriteHeader(http.StatusOK)
			}
		} else {
			response.WriteHeader(400)
		}
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
func builderStart(request *restful.Request, response *restful.Response) {
	meta := getHTTPFileMeta(request)
	datastore.CreateTempFile(meta.Name)
	response.WriteHeader(http.StatusOK)
}
