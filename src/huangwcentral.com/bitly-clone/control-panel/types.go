/*
	Butly API ( Version 3.0 )
*/

package main

//import "gopkg.in/mgo.v2/bson"

type shortlink struct {
	Id string `json:"id" bson:"_id"`
}

type shortlinkMsg struct {
	OrigUrl  string
	ShortUrl string
}

type shortlinkReq struct {
	OrigUrl string
}

type shortlinkResp struct {
	ShortUrl string
}
