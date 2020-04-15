/*
	Butly API ( Version 1.0 )
	CP, LR -> MySQL
	shortlink = insert Id
*/

package main

import (
	"strconv"
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/satori/go.uuid"
    "database/sql"
	_ "github.com/go-sql-driver/mysql"
)

/*
	Go's SQL Package:
		Tutorial: http://go-database-sql.org/index.html
		Reference: https://golang.org/pkg/database/sql/
*/

var mysql_connect = "root:cmpe281@tcp(localhost:3306)/cmpe281"
//var mysql_connect = "root:cmpe281@tcp(mysql:3306)/cmpe281"

// RabbitMQ Config
var rabbitmq_server = "rabbit"
var rabbitmq_port = "5672"
var rabbitmq_queue = "gumball"
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
			count int
			model string
			serial string
		)
		rows, err := db.Query("select id, count_gumballs, model_number, serial_number from gumball where id = ?", 1)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&id, &count, &model, &serial)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(id, count, model, serial)
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
	mx.HandleFunc("/urls", butlyShortenUrlHandler(formatter)).Methods("POST")
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
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"API version 1.0 alive!"})
	}
}

// API Shorten a URL
func butlyShortenUrlHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var reqBody shortenUrlReq

		err := json.NewDecoder(req.Body).Decode(&reqBody)
		failOnError(err, "Error parsing request body")
		// reqBody, err := ioutil.ReadAll(req.Body)
		// failOnError(err, "Error reading request body")

		fmt.Printf("Short Request: %+v \n", reqBody)

		var (
			id int
			orig_url string
			short_url string
			visits int
		)

		id = insertId + 1
		orig_url = reqBody.OrigUrl
		short_url = strconv.Itoa(id)
		visits = 0

		fmt.Println("Insert Values: ", id, orig_url, short_url, visits)

		db, err := sql.Open("mysql", mysql_connect)
		defer db.Close()
		if err != nil {
			log.Fatal(err)
		} else {
			result, insErr := db.Exec("insert into tiny_urls ( id, orig_url, short_url, visits ) values ( ?, ?, ?, ? ) ;", id, orig_url, short_url, visits )
			if insErr != nil {
				log.Fatal( insErr )
			} else {
				lastId, _ := result.LastInsertId()
				insertId = int(lastId)
			}
		}

		result := shortenUrlResp {
			ShortUrl: short_url,
		}
		fmt.Println("Shortened Url: ", result)
		formatter.JSON( w, http.StatusOK, result)
	}
}

/*

	-- Create Database Schema (DB User: root, DB Pass: cmpe281)

		Database Schema: cmpe281

	-- Create Database Table

		CREATE TABLE tiny_urls ( id bigint(20) NOT NULL AUTO_INCREMENT, orig_url varchar(512) NOT NULL, short_url varchar(45) NOT NULL, visits int(11) NOT NULL, created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (id), UNIQUE KEY short_url (short_url) ) ;

		CREATE TABLE gumball ( id bigint(20) NOT NULL AUTO_INCREMENT, version bigint(20) NOT NULL, count_gumballs int(11) NOT NULL, model_number varchar(255) NOT NULL, serial_number varchar(255) NOT NULL, PRIMARY KEY (id), UNIQUE KEY serial_number (serial_number) ) ;

	-- Load Data

		insert into gumball ( id, version, count_gumballs, model_number, serial_number ) values ( 1, 0, 1000, 'M102988', '1234998871109' ) ;

	-- Verify Data

		select * from gumball ;


*/
