package main

import "net/http"
import "mime"
import "errors"
import "fmt"
import "log"

// Meta describes HTTP metainfo
type Meta struct {
	MediaType string
	Boundary  string
	Range     *Range
	FileName  string
	Property  *Property
}

// Range describes HTTP range
type Range struct {
	Start int64
	End   int64
	Size  int64
}

// Property describes user-define file info
type Property struct {
	Name    string
	Hash    string
	Creator string
	SysID   string
}

// ParseMeta parses request information and makes Meta
func ParseMeta(req *http.Request) (*Meta, error) {
	meta := &Meta{}

	if err := meta.parseContentType(req.Header.Get("Content-Type")); err != nil {
		return nil, err
	}

	if err := meta.parseContentRange(req.Header.Get("Content-Range")); err != nil {
		return nil, err
	}

	if err := meta.parseContentDisposition(req.Header.Get("Content-Disposition")); err != nil {
		return nil, err
	}

	// Parse user defined HTTP headers

	meta.parseName(req.Header.Get("name"))
	meta.parseCreator(req.Header.Get("creator"))
	meta.parseHash(req.Header.Get("hash"))
	meta.parseSysID(req.Header.Get("sysId"))

	return meta, nil
}

func (meta *Meta) parseContentType(ct string) error {
	if ct == "" {
		meta.MediaType = "application/octet-stream"
		return nil
	}

	mediatype, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}

	if mediatype == "multipart/form-data" {
		boundary, ok := params["boundary"]
		if !ok {
			return errors.New("Meta: boundary not defined")
		}

		meta.MediaType = mediatype
		meta.Boundary = boundary
	} else {
		meta.MediaType = "application/octet-stream"
	}

	return nil
}

func (meta *Meta) parseContentRange(cr string) error {
	if cr == "" {
		return nil
	}

	var start, end, size int64

	_, err := fmt.Sscanf(cr, "bytes %d-%d/%d", &start, &end, &size)
	if err != nil {
		return err
	}

	meta.Range = &Range{Start: start, End: end, Size: size}

	return nil
}

func (meta *Meta) parseContentDisposition(cd string) error {
	if cd == "" {
		return nil
	}

	_, params, err := mime.ParseMediaType(cd)
	if err != nil {
		return err
	}

	filename, ok := params["filename"]
	if !ok {
		return errors.New("Meta: file in Content-Disposition not defined")
	}

	meta.FileName = filename

	return nil
}

// Parse user defined headers

func (meta *Meta) parseName(name string) error {
	if name == "" {
		log.Println("Request with empty Name header")
		return nil
	}

	meta.Property.Name = name

	return nil
}

const sha256Size = 32

func (meta *Meta) parseHash(h string) error {
	if h == "" {
		log.Println("Request with empty Hash header")
		return nil
	}

	if len(h) != sha256Size {
		log.Println("Hash size not equal to SHA-2 size")
		return nil
	}

	meta.Property.Hash = h

	return nil
}

func (meta *Meta) parseCreator(creator string) error {
	if creator == "" {
		log.Println("Request with empty Creator header")
		return nil
	}

	meta.Property.Creator = creator

	return nil
}

func (meta *Meta) parseSysID(sysID string) error {
	if sysID == "" {
		log.Println("Request with empty sysId header")
		return nil
	}

	meta.Property.SysID = sysID

	return nil
}
