/*
	Butly API ( Version 3.0 )
*/

package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/streadway/amqp"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
	"github.com/satori/go.uuid"
)

// MongoDB Config
var mongodb_server = os.Getenv("MONGODB_SERVER")
var mongodb_user = os.Getenv("MONGODB_USER")
var mongodb_password = os.Getenv("MONGODB_PASSWORD")
var mongodb_database = "cmpe281"
var mongodb_collection = "shortlinks"

// RabbitMQ Config
var rabbitmq_server = os.Getenv("RABBITMQ_SERVER")
var rabbitmq_port = "5672"
var rabbitmq_exchange = "message_bus"
var rabbitmq_user = os.Getenv("RABBITMQ_USER")
var rabbitmq_pass = os.Getenv("RABBITMQ_PASSWORD")

// UUID
var this_id = uuid.NewV4()

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

// Check Connections
func init() {

	// check mongo-cp
	session, err := mgo.Dial(mongodb_user+":"+mongodb_password+"@"+mongodb_server)
  failOnError(err, "Error connecting to mongodb")
	defer session.Close()

	// check database and table
	var short shortlink
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mongodb_database).C(mongodb_collection)
	_ = c.Find(bson.M{}).One(&short)
	fmt.Println(short)

	// check rabbitmq
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
}

// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/link_save", butlyShortenUrlHandler(formatter)).Methods("POST")
}

// API Ping Handler
func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"CP Server: " + this_id.String() + " - API version 3.0 alive!"})
	}
}

// API Shorten a URL
func butlyShortenUrlHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var reqBody shortlinkReq

		err := json.NewDecoder(req.Body).Decode(&reqBody)
		if err != nil {
			warnOnError(err, "Erroring processing request body")
			log.Println("Request body: ", req.Body)
			formatter.JSON(w, http.StatusBadRequest, nil)
			return
		}

		fmt.Printf("Short Request: %+v \n", reqBody)

		if reqBody.OrigUrl == "" {
			log.Println("Request Body is empty")
			formatter.JSON(w, http.StatusBadRequest, nil)
			return
		}
		url, err := url.Parse(reqBody.OrigUrl)
		warnOnError(err, "Error parsing url")
		if ( url.Opaque != "" || url.Host == "" ) && url.Scheme != "" && url.Scheme != "http" && url.Scheme == "https" {
			log.Println("Invalid URL")
			formatter.JSON(w, http.StatusBadRequest, nil)
			return
		}

		// connect to mongo
		session, err := mgo.Dial(mongodb_user+":"+mongodb_password+"@"+mongodb_server)
		if err != nil {
			warnOnError(err, "Error connection to " + mongodb_server)
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodb_database).C(mongodb_collection)

		// insert into mongo and remove shortlink
		var short_url shortlink
		err = c.Find(bson.M{}).One(&short_url)
		if err != nil {
			warnOnError(err, "Ran out of shortlinks")
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}
		_ = c.RemoveId(short_url.Id)
		msg := shortlinkMsg {
			OrigUrl: reqBody.OrigUrl,
			ShortUrl: short_url.Id,
		}

		msgJson, _ := json.Marshal(msg)
		queue_send( string(msgJson) )

		result := shortlinkResp {
			ShortUrl: short_url.Id,
		}
		fmt.Println("Shortened Url: ", result)
		formatter.JSON(w, http.StatusOK, result)
	}
}

// Send Order to Queue for Processing
func queue_send(message string) {
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	warnOnError(err, "Error connecting to rabbitmq: ")
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	_ = ch.ExchangeDeclare(
		rabbitmq_exchange,  // name
		"topic", 					// type
	   true,     					// durable
	   false,   					// auto-deleted
	   false,  					  // internal
	   false,   					// no-wait
	   nil,     					// arguments
	)

	_ = ch.Publish(
		rabbitmq_exchange,  // exchange
		"cp.shortlink.create", // routing key
		false,  					  // mandatory
		false,  					  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		})
	log.Printf(" [x] Sent %s", message)
}

// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func warnOnError(err error, msg string) {
	if err != nil {
		log.Println("%s: %s", msg, err)
	}
}
