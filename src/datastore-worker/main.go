/*
	Butly API ( Version 3.0 )
*/

package main

import (
	"os"
	"fmt"
	"log"
	"encoding/json"
	"database/sql"
	"github.com/streadway/amqp"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
)

// MongoDB Config
var mongodb_server = os.Getenv("MONGODB_SERVER")
var mongodb_user = "cmpe281"
var mongodb_password = "mymongoeventpassword"
var mongodb_database = "cmpe281"
var mongodb_collection = "eventlogs"

// MySQL Config
var mysql_server = os.Getenv("MYSQL_SERVER")
var mysql_user = "root"
var mysql_password = "i2FO95C0OqHyhEb9"
var mysql_connect = mysql_user + ":" + mysql_password + "@tcp(" + mysql_server + ")/cmpe281"

// RabbitMQ Config
var rabbitmq_server = os.Getenv("RABBITMQ_SERVER")
var rabbitmq_port = "5672"
var rabbitmq_exchange = "message_bus"
var rabbitmq_queue = "datastore_queue"
var rabbitmq_user = "user"
var rabbitmq_pass = "password"

func main() {

	// connect to rabbitmq
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// connect to mysql
	db, err := sql.Open("mysql", mysql_connect)
	failOnError(err, "Error opening mysql connection")
	defer db.Close()

	// connect to event store
/*
	tlsConfig := &tls.Config{}
	dialInfo := &mgo.DialInfo{
	    Addrs: []string{"mongostorage-shard-00-00-cscrb.mongodb.net:27017",
	                    "mongostorage-shard-00-01-cscrb.mongodb.net:27017",
	                    "mongostorage-shard-00-02-cscrb.mongodb.net:27017"},
	    Database: "admin",
	    Username: "cmpe281_user",
	    Password: "GTe6oAN1ZxqTRfyJ",
	}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
	    conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
	    return conn, err
	}
	session, err := mgo.DialWithInfo(dialInfo)
	failOnError(err, "Error connecting to mongodb")
	defer session.Close()
*/
	session, err := mgo.Dial(mongodb_user+":"+mongodb_password+"@"+mongodb_server)
	failOnError(err, "Error connecting to mongodb")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mongodb_database).C(mongodb_collection)

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
		q.Name, 					 // queue name
	  "*.shortlink.*", 	 // routing key
	  rabbitmq_exchange, // exchange
	  false,
	  nil,
	)

	msgs, _ := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// unmarshal body
			var msg shortlinkMsg
		  _ = json.Unmarshal(d.Body, &msg)

			switch d.RoutingKey {
			case "cp.shortlink.create":
				// insert new shortlink into mysql
				res, _ := db.Exec("insert into tiny_urls ( orig_url, short_url ) values ( ?, ? ) ;", msg.OrigUrl, msg.ShortUrl)
				lastInsId, _ := res.LastInsertId()
				rowsInsed, _ := res.RowsAffected()
				fmt.Println(d.RoutingKey, " Last Inserted: ", lastInsId, " Rows Affected: ", rowsInsed )
			case "lr.shortlink.update":
				// update visits in MySQL
				res, _ := db.Exec("update tiny_urls set visits = visits + 1 where short_url = ?", msg.ShortUrl)
				lastInsId, _ := res.LastInsertId()
				rowsInsed, _ := res.RowsAffected()
				fmt.Println(d.RoutingKey, " Last Inserted: ", lastInsId, " Rows Affected: ", rowsInsed )
			default:
				log.Println("Invalid Routing Key: %s", d.RoutingKey)
			}
			go func() {
				// create new event log
				_ = c.Insert(bson.M{"routingkey": d.RoutingKey,
													  "body": msg,
													 })
			}()
			d.Ack(false)
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

func warnOnError(err error, msg string) {
	if err != nil {
		log.Println("%s: %s", msg, err)
	}
}

/*

	-- Create Database Schema (DB User: root, DB Pass: cmpe281)

		Database Schema: cmpe281

	-- Create Database Table

		CREATE TABLE tiny_urls (
		id bigint(20) NOT NULL AUTO_INCREMENT, orig_url varchar(512) NOT NULL, short_url varchar(45) NOT NULL, visits int(11) NOT NULL DEFAULT 0, created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (id), UNIQUE KEY short_url (short_url) ) ;

	-- Create Procedure

*/
