/*
	Butly API ( Version 3.0 )
*/

package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
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
	mx.HandleFunc("/t/merge", butlyMergeTrendsHandler(formatter)).Methods("GET")
	mx.HandleFunc("/t/{key}", butlyShortlinkTrendHandler(formatter)).Methods("GET")
}

// API Ping Handler
func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"TS Server: " + this_id.String() + " - API version 3.0 alive!"})
	}
}

// API Merge current visits
func butlyMergeTrendsHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		session, err := mgo.Dial(mongodb_user + ":" + mongodb_password + "@" + mongodb_server)
		if err != nil {
			warnOnError(err, "Error connecting to mongo-ts")
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodb_database).C(mongodb_collection)

		pipe := c.Pipe([]bson.M{
			{"$match": bson.M{}},
			{"$group": bson.M{"_id": "$origurl", "visits": bson.M{"$sum": "$visits"}}},
			{"$sort": bson.M{"visits": -1}},
			{"$limit": 5}})
		resp := []bson.M{}
		err = pipe.All(&resp)
		if err != nil {
			warnOnError(err, "Error with mongo pipe")
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}
		result := []trendResult{}
		for _, d := range resp {
			var document trendResult
			document.OrigUrl = d["_id"].(string)
			document.Visits = d["visits"].(int)
			result = append(result, document)
		}
		fmt.Println(result)
		formatter.JSON(w, http.StatusOK, result)
	}
}

// API Get current visits for a link
func butlyShortlinkTrendHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		params := mux.Vars(req)
		var shortUrl string = params["key"]
		fmt.Println("Short Url Key: ", shortUrl)

		if shortUrl == "" {
			formatter.JSON(w, http.StatusBadRequest, nil)
			return
		}

		session, err := mgo.Dial(mongodb_user + ":" + mongodb_password + "@" + mongodb_server)
		if err != nil {
			warnOnError(err, "Error connecting to mongo-ts")
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodb_database).C(mongodb_collection)
		var doc trendServerEntry
		// handle the case of document not found
		err = c.Find(bson.M{"_id": shortUrl}).One(&doc)
		if err != nil {
			warnOnError(err, "Error finding origurl in mongodb")
			formatter.JSON(w, http.StatusNotFound, nil)
			return
		}
		fmt.Println(doc)
		httpResponse := shortlinkTrend{
			OrigUrl:  doc.OrigUrl,
			ShortUrl: doc.Id,
			Visits:   doc.Visits,
		}
		formatter.JSON(w, http.StatusOK, httpResponse)
	}
}
