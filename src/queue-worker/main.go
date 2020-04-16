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
	"github.com/streadway/amqp"
	"encoding/json"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
)

//var mysql_connect = "root:cmpe281@tcp(localhost:3306)/cmpe281"
var mysql_connect = "root:cmpe281@tcp(mysql:3306)/cmpe281"

// MongoDB Config
var mongodb_server = "mongo"
var mongodb_database = "cmpe281"
var mongodb_collection = "url_lookup"

// RabbitMQ Config
var rabbitmq_server = "rabbitmq"
var rabbitmq_port = "5672"
var rabbitmq_queue = "create_queue"
var rabbitmq_user = "user"
var rabbitmq_pass = "password"

func main() {
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		rabbitmq_queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	db, err := sql.Open("mysql", mysql_connect)
	defer db.Close()
	failOnError(err, "Error opening mysql connection")

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var msg shortlinkMsg
			err := json.Unmarshal(d.Body, &msg)
			failOnError(err, "Error decoding message from create_queue")

			// insert into mysql
			res, err := db.Exec("insert into tiny_urls ( orig_url, short_url, visits ) values ( ?, ?, ? ) ;", msg.OrigUrl, msg.ShortUrl, 0)
			failOnError(err, "Error inserting new entry into database")
			lastInsId, _ := res.LastInsertId()
			rowsInsed, _ := res.RowsAffected()
			fmt.Println( "Last Inserted: ", lastInsId, " Rows Affected: ", rowsInsed )

			// insert into mongo
			session, err := mgo.Dial(mongodb_server)
			failOnError(err, "Error connecting to mongodb")
			defer session.Close()
			session.SetMode(mgo.Monotonic, true)
			c := session.DB(mongodb_database).C(mongodb_collection)
			err = c.Insert(&shortlinkDoc{msg.OrigUrl, msg.ShortUrl, 0})
			failOnError(err, "Error inserting document into mongo")
			result := shortlinkDoc{}
			err =  c.Find(bson.M{"shorturl": msg.ShortUrl}).One(&result)
			failOnError(err, "Error finding inserted mongo document")
			fmt.Println(result)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
