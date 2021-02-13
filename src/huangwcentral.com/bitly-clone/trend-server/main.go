/*
	Butly API ( Version 3.0 )
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
)

// MongoDB Config
var mongodb_server = os.Getenv("MONGODB_SERVER")
var mongodb_user = os.Getenv("MONGODB_USER")
var mongodb_password = os.Getenv("MONGODB_PASSWORD")
var mongodb_database = "cmpe281"
var mongodb_collection = "visits"

// RabbitMQ Config
var rabbitmq_server = os.Getenv("RABBITMQ_SERVER")
var rabbitmq_port = "5672"
var rabbitmq_exchange = "message_bus"
var rabbitmq_queue = "ts_queue"
var rabbitmq_user = os.Getenv("RABBITMQ_USER")
var rabbitmq_pass = os.Getenv("RABBITMQ_PASSWORD")

func main() {

	// check mongo
	session, err := mgo.Dial(mongodb_user + ":" + mongodb_password + "@" + mongodb_server)
	failOnError(err, "Error connecting to mongodb")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mongodb_database).C(mongodb_collection)

	// check for database / table?

	// check rabbitmq
	conn, err := amqp.Dial("amqp://" + rabbitmq_user + ":" + rabbitmq_pass + "@" + rabbitmq_server + ":" + rabbitmq_port + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	warnOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		rabbitmq_exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	warnOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		rabbitmq_queue, // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	warnOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,                // queue name
		"lr.shortlink.update", // routing key
		rabbitmq_exchange,     // exchange
		false,
		nil,
	)
	warnOnError(err, "Failed to bind queue to exchange")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	warnOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var msg shortlinkMsg
			err := json.Unmarshal(d.Body, &msg)
			warnOnError(err, "Error decoding message from create_queue")

			// check if document exists
			count, err := c.Find(bson.M{"_id": msg.ShortUrl}).Count()
			warnOnError(err, "Error finding inserted mongo document")
			if count == 0 {
				// insert into mongo
				err = c.Insert(bson.M{"_id": msg.ShortUrl,
					"origurl": msg.OrigUrl,
					"visits":  1})
				warnOnError(err, "Error inserting document into mongo")
			} else {
				// update visits
				query := bson.M{"_id": msg.ShortUrl}
				change := bson.M{"$inc": bson.M{"visits": 1}}
				err = c.Update(query, change)
				warnOnError(err, "Error updating visits of accessed url")
			}
		}
	}()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3002"
	}

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
