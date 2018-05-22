package main

// FileMeta describes user-define file info
type FileMeta struct {
	Name    string `bson:"name" json:"name"`
	Hash    string `bson:"hash" json:"hash"`
	Creator string `bson:"creator" json:"creator"`
	SysID   string `bson:"sysId" json:"sysId"`
}
