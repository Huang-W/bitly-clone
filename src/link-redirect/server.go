/*
	Butly API ( Version 3.0 )
*/

package main

import (
	"fmt"
	"log"
	"strings"
	"bytes"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/streadway/amqp"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/satori/go.uuid"
	"database/sql"
_ "github.com/go-sql-driver/mysql"
)

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

// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/r/{key}", butlyRedirectUrlHandler(formatter)).Methods("GET")
}

// API Ping Handler
func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"LR Server: " + this_id.String() + " - API version 3.0 alive!"})
	}
}

// Test Redirect
func testHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "http://www.golang.org", http.StatusMovedPermanently)
	}
}

// API Redirect a URL
func butlyRedirectUrlHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		params := mux.Vars(req)
		var short_url string = params["key"]
		fmt.Println( "Short Url Key: ", short_url )

		if short_url == ""  {
			formatter.JSON(w, http.StatusBadRequest, nil)
			return
		}

		// look up short url in cache
		resp, err := http.Get("http://"+nosql_host+":"+nosql_port+"/"+nosql_api+"/"+short_url)
		if ( err != nil ) {
			warnOnError(err, "Error getting doc from NoSQL")
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}
		defer resp.Body.Close()

		var	(
			httpResponse originalUrl
			queueMsg shortlinkMsg
			queueMsgJson []byte
		)

		if resp.StatusCode == 200 {

			// cache hit
			respBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("Response Body: ", respBody)
			_ = json.Unmarshal(respBody, &queueMsg)
			httpResponse = originalUrl{ queueMsg.OrigUrl }
			queueMsg.ShortUrl = short_url

		} else {

			// cache miss - query mysql
			db, err := sql.Open("mysql", mysql_connect)
			if err != nil {
				warnOnError(err, "Error connecting to mysql")
				formatter.JSON(w, http.StatusInternalServerError, nil)
				return
			} else {
				row := db.QueryRow("select orig_url from tiny_urls where short_url = ?", short_url)
				err := row.Scan(&httpResponse.OrigUrl)
				if err != nil {
					warnOnError(err, "Short URL not found")
					formatter.JSON(w, http.StatusNotFound, nil)
					return
				}
			}
			defer db.Close()

			queueMsg.OrigUrl = httpResponse.OrigUrl
			queueMsg.ShortUrl = short_url

			// create entry in cache
			client := &http.Client{}
			log.Println("Original URL from MySQL: ", httpResponse)
			queueMsgJson, _ = json.Marshal(queueMsg)
			reader := bytes.NewReader( []byte( queueMsgJson ) )
			req, err := http.NewRequest(http.MethodPost, "http://"+nosql_host+":"+nosql_port+"/"+nosql_api+"/"+short_url, reader)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Type", "text/plain")
			resp, err := client.Do(req)
			warnOnError(err, "Error creating entry in NoSQL project")

			// verify insertion
			body, _ := ioutil.ReadAll(resp.Body)
			var respBody shortlinkMsg
			_ = json.Unmarshal(body, &respBody)
			fmt.Println("Inserted Document into NoSQL K/V: ", respBody)
		}
		queueMsgJson, _ = json.Marshal(queueMsg)
		warnOnError(err, "Error marshaling json from shortlinkMsg")
		queue_send( queueMsgJson )
		if (strings.HasPrefix( httpResponse.OrigUrl, "http://" ) || strings.HasPrefix( httpResponse.OrigUrl, "https://" )) == false {
			httpResponse.OrigUrl = "http://" + httpResponse.OrigUrl;
		}
		http.Redirect(w, req, httpResponse.OrigUrl, http.StatusMovedPermanently)
	}
}

// update visits
func queue_send(message []byte) {
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		rabbitmq_exchange,  // name
		"topic", 					  // type
	   true,     					// durable
	   false,   					// auto-deleted
	   false,  					  // internal
	   false,   					// no-wait
	   nil,     					// arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = ch.Publish(
		rabbitmq_exchange,     // exchange
		"lr.shortlink.update", // routing key
		false,  				    	 // mandatory
		false,  					     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         message,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", message)
}
