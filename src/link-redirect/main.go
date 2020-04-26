/*
	Butly API ( Version 3.0 )
*/

package main

import (
	// set the PORT ENV variable
	"os"
	"fmt"
	"log"
	"bytes"
	"io/ioutil"
	"net/http"
	"github.com/streadway/amqp"
	"encoding/json"
	"database/sql"
_ "github.com/go-sql-driver/mysql"
)

// NoSQL Lookup Cache
var nosql_host = os.Getenv("NOSQL_HOST")
var nosql_port = "9090"
var nosql_api = "api"

// RabbitMQ Config
var rabbitmq_server = os.Getenv("RABBITMQ_SERVER")
var rabbitmq_port = "5672"
var rabbitmq_exchange = "message_bus"
var rabbitmq_queue = "lr_queue"
var rabbitmq_user = os.Getenv("RABBITMQ_USER")
var rabbitmq_pass = os.Getenv("RABBITMQ_PASSWORD")

// MySQL Config
var mysql_server = os.Getenv("MYSQL_SERVER")
var mysql_user = os.Getenv("MYSQL_USER")
var mysql_password = os.Getenv("MYSQL_PASSWORD")
var mysql_connect = mysql_user + ":" + mysql_password + "@tcp(" + mysql_server + ")/cmpe281"

func main() {

	// check rabbitmq
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// check nosql
	resp, err := http.Get("http://"+nosql_host+":"+nosql_port+"/"+nosql_api)
  failOnError(err, "Error communicating with NoSQL project")
	fmt.Println("NoSQL Response Status: ", resp.Status)
	resp.Body.Close()

	// check mysql
	db, err := sql.Open("mysql", mysql_connect)
	failOnError(err, "Error connecting to mysql")
	defer db.Close()
	var (
		id int
		short_url string
		orig_url string
	)
	row := db.QueryRow("select id, short_url, orig_url claimed from tiny_urls where ? limit 1", 1)
	err = row.Scan(&id, &short_url, &orig_url)
	warnOnError(err, "Error processing query in init")
	log.Println(id, short_url, orig_url)

	ch, _ := conn.Channel()
	defer ch.Close()

	_ = ch.ExchangeDeclare(
		rabbitmq_exchange,  // name
		"topic",  					// type
	   true,     					// durable
	   false,   					// auto-deleted
	   false,  					  // internal
	   false,   					// no-wait
	   nil,     					// arguments
	)

	q, _ := ch.QueueDeclare(
		rabbitmq_queue,			 // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	_ = ch.QueueBind(
		q.Name, 					     // queue name
	  "cp.shortlink.create", // routing key
	  rabbitmq_exchange,     // exchange
	  false,
	  nil,
	)

	msgs, _ := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	client := &http.Client{}

	// run consumer in background
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var msg shortlinkMsg
			_ = json.Unmarshal(d.Body, &msg)
			reader := bytes.NewReader(d.Body)

			req, _ := http.NewRequest(http.MethodPost, "http://"+nosql_host+":"+nosql_port+"/"+nosql_api+"/"+msg.ShortUrl, reader)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Type", "text/plain")

			// insert into Cache
			resp, err := client.Do(req)
			warnOnError(err, "Error creating entry in NoSQL project")
			body, _ := ioutil.ReadAll(resp.Body)
			var respBody shortlinkMsg
			_ = json.Unmarshal(body, &respBody)
			warnOnError(err, "Error unmarshaling NoSQL response to create document")
			fmt.Println("Inserted Document into NoSQL K/V: ", respBody)
			resp.Body.Close()
		}
	}()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3001"
	}

	// run http server
	server := NewServer()
	server.Run(":" + port)

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
