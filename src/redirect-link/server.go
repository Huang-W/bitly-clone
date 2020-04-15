/*
	Butly API ( Version 1.0 )
	CP, LR -> MySQL
	shortlink = insert Id
*/

package main

import (
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
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"API version 1.0 alive!"})
	}
}

// API Redirect a URL
func butlyRedirectUrlHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		params := mux.Vars(req)
		var shortUrl string = params["key"]
		fmt.Println( "Short Url Key: ", shortUrl )
		var orig_url string

		if shortUrl == ""  {
			formatter.JSON(w, http.StatusBadRequest, "")
		} else {
			db, err := sql.Open("mysql", mysql_connect)
			defer db.Close()
			if err != nil {
				log.Fatal(err)
			} else {
				rows, _ := db.Query("select orig_url from tiny_urls where short_url = ? ;", shortUrl )
				defer rows.Close()
				for rows.Next() {
					rows.Scan(&orig_url)
					log.Println(orig_url)
				}
				result := redirectUrlResp {
					OrigUrl: orig_url,
				}
				formatter.JSON(w, http.StatusOK, result)
			}
		}
	}
}

/*

	-- Create Database Schema (DB User: root, DB Pass: cmpe281)

		Database Schema: cmpe281

	-- Create Database Table

		CREATE TABLE tiny_urls ( id bigint(20) NOT NULL AUTO_INCREMENT, orig_url varchar(512) NOT NULL, short_url varchar(45) NOT NULL, visits int(11) NOT NULL, created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (id), UNIQUE KEY short_url (short_url) ) ;

	-- Load Data

		insert into tiny_urls ( id, orig_url, short_url, visits ) values ( 1, 'ifconfig.co', '1', 0 ) ;

	-- Verify Data

		select * from tiny_urls ;


*/
