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
	"fmt"
	"log"
	"net/http"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

// MongoDB Config
var mongodb_server = "mongo"
var mongodb_database = "cmpe281"
var mongodb_collection = "url_lookup"

// RabbitMQ Config
var rabbitmq_server = "rabbitmq"
var rabbitmq_port = "5672"
var rabbitmq_queue = "new_shortlink"
var rabbitmq_user = "user"
var rabbitmq_pass = "password"

// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	n := negroni.Classic()
	mx := mux.NewRouter()
	initRoutes(mx, formatter)
	n.UseHandler(mx)
	return n
}

// Init Mongo DB Connection
func init() {

	session, err := mgo.Dial(mongodb_server)
  failOnError(err, "Error connecting to mongodb")
	defer session.Close()

}


// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/r/{key}", butlyRedirectUrlHandler(formatter)).Methods("GET")
}

// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// API Ping Handler
func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"LR Server - API version 2.0 alive!"})
	}
}

// API Redirect a URL
func butlyRedirectUrlHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		params := mux.Vars(req)
		var shortUrl string = params["key"]
		fmt.Println( "Short Url Key: ", shortUrl )

		if shortUrl == ""  {
			formatter.JSON(w, http.StatusBadRequest, "")
		} else {
			session, err := mgo.Dial(mongodb_server)
			if err != nil {
				log.Fatal(err)
			}
			defer session.Close()
			session.SetMode(mgo.Monotonic, true)
			c := session.DB(mongodb_database).C(mongodb_collection)
			doc := &shortlinkDoc{}
			err = c.Find(bson.M{"shorturl": shortUrl}).One(&doc)
			failOnError(err, "Error finding shorturl in mongodb")
			fmt.Println( doc )
			query := bson.M{"_id": doc.Id}
			change := bson.M{"$inc": bson.M{"visits" : 1}}
			err = c.Update( query, change )
			failOnError(err, "Error updating visits of accessed url")
			result := redirectUrlResp{ doc.OrigUrl }
			formatter.JSON(w, http.StatusOK, result)
		}
	}
}
