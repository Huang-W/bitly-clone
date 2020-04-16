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
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/streadway/amqp"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
    "database/sql"
	_ "github.com/go-sql-driver/mysql"
)

/*
	Go's SQL Package:
		Tutorial: http://go-database-sql.org/index.html
		Reference: https://golang.org/pkg/database/sql/
*/

//var mysql_connect = "root:cmpe281@tcp(localhost:3306)/cmpe281"
var mysql_connect = "root:cmpe281@tcp(mysql:3307)/cmpe281"

// RabbitMQ Config
var rabbitmq_server = "rabbitmq"
var rabbitmq_port = "5672"
var create_queue = "create_queue"
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

// Init MySQL DB Connection
func init() {

		db, err := sql.Open("mysql", mysql_connect)
		if err != nil {
			log.Fatal(err)
		} else {
			var (
				id int
				short_url string
				claimed bool
			)
			rows, err := db.Query("select id, short_url, claimed from short_links where ? limit 1", 1)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			for rows.Next() {
				err := rows.Scan(&id, &short_url, &claimed)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(id, short_url, claimed)
			}
			err = rows.Err()
			if err != nil {
				log.Fatal(err)
			}
		}
		defer db.Close()

}

// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/link_save", butlyShortenUrlHandler(formatter)).Methods("POST")
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
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"CP Server - API version 2.0 alive!"})
	}
}

// API Shorten a URL
func butlyShortenUrlHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var reqBody shortlinkReq

		err := json.NewDecoder(req.Body).Decode(&reqBody)
		failOnError(err, "Error parsing request body")

		fmt.Printf("Short Request: %+v \n", reqBody)

		var (
			id int
			orig_url string
			short_url string
		)

		orig_url = reqBody.OrigUrl

		db, err := sql.Open("mysql", mysql_connect)
		defer db.Close()
		if err != nil {
			log.Fatal(err)
		} else {
			rows, err := db.Query("select id, short_url from short_links where claimed = false limit 1 ;" )
			failOnError(err, "Error on selecting new short_link from mysql db")
			defer rows.Close()
			for rows.Next() {
				err := rows.Scan(&id, &short_url)
				failOnError(err, "Error on scanning row")
				log.Println(id, short_url)
			}
		}
		res, err := db.Exec("update short_links set claimed = true where id = ?", id)
		failOnError( err, "Error on claiming short_url")
		log.Println(res)

		msg := shortlinkMsg {
			OrigUrl: orig_url,
			ShortUrl: short_url,
		}

		msgJson, err := json.Marshal(msg)
		failOnError(err, "Error marshalling json from shortlink")
		queue_send( string(msgJson) )

		result := shortlinkResp {
			ShortUrl: short_url,
		}
		fmt.Println("Shortened Url: ", result)
		formatter.JSON(w, http.StatusOK, result)
	}
}

// Send Order to Queue for Processing
func queue_send(message string) {
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		create_queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := message
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	log.Printf(" [x] Sent %s", body)
	failOnError(err, "Failed to publish a message")
}

/*

	-- Create Database Schema (DB User: root, DB Pass: cmpe281)

		Database Schema: cmpe281

	-- Create Database Table

		CREATE TABLE tiny_urls (
		id bigint(20) NOT NULL AUTO_INCREMENT, orig_url varchar(512) NOT NULL, short_url varchar(45) NOT NULL, visits int(11) NOT NULL, created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (id), UNIQUE KEY short_url (short_url) ) ;

		CREATE TABLE short_links (
		id bigint(20) NOT NULL AUTO_INCREMENT, short_url varchar(45) BINARY NOT NULL, claimed tinyint(1) DEFAULT false, PRIMARY KEY (id), UNIQUE KEY short_url (short_url) ) ;

	-- Create Procedure

*/
