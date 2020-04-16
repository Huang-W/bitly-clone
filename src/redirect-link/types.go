/*
	Butly API ( Version 2.0 )
	CP -> create_queue
	CP -> CP:MySQL
	LR -> MongoDB
	QW <- create_queue
	QW -> Main:MySQL
	QW -> MongoDB
*/

package main

import (
  "gopkg.in/mgo.v2/bson"
)

type shortlinkDoc struct {
	Id bson.ObjectId `json:"id" bson:"_id,omitempty"`
	OrigUrl string
	ShortUrl string
	Visits int
}

type redirectUrlResp struct {
	OrigUrl string
}
