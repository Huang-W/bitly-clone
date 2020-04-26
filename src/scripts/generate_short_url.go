package main

import (
  "fmt"
  "time"
  "math/rand"
  "log"
  "math/big"
  "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

// MongoDB Config
var mongodb_server = "35.225.69.179"
var mongodb_user = "cmpe281"
var mongodb_password = "mymongocppassword"
var mongodb_database = "cmpe281"
var mongodb_collection = "shortlinks"

func main() {

  // check mongo-cp
	session, err := mgo.Dial(mongodb_user+":"+mongodb_password+"@"+mongodb_server)
  failOnError(err, "Error connecting to mongodb")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mongodb_database).C(mongodb_collection)
  b := c.Bulk()
  vals := []string{}
  for i := int64(62); i < 3844; i++ {
    short_url := big.NewInt(i).Text(62)
    vals = append( vals, short_url )
  }

  Shuffle( vals )

  mgoVals := []interface{}{}
  for _, d := range vals {
    mgoVals = append( mgoVals, bson.M{ "_id": d } )
  }

  // insert into mongo
  b.Insert(mgoVals...)
  res, err := b.Run()
  failOnError(err, "Failed bulk insert")
  fmt.Println(res)
}

// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// https://www.calhoun.io/how-to-shuffle-arrays-and-slices-in-go/
func Shuffle(vals []string) {
  r := rand.New(rand.NewSource(time.Now().Unix()))
  for len(vals) > 0 {
    n := len(vals)
    randIndex := r.Intn(n)
    vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
    vals = vals[:n-1]
  }
}

/*

	-- Create Database Schema (DB User: root, DB Pass: cmpe281)

		Database Schema: cmpe281

	-- Create Database Table

		CREATE TABLE short_links (
		id bigint(20) NOT NULL AUTO_INCREMENT, short_url varchar(45) BINARY NOT NULL, claimed tinyint(1) DEFAULT false, PRIMARY KEY (id), UNIQUE KEY short_url (short_url) ) ;

	-- Create Procedure

*/
