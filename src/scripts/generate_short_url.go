package main

import (
  "fmt"
  "time"
  "math/rand"
  "strings"
  "log"
  "math/big"
  "database/sql"
_ "github.com/go-sql-driver/mysql"
)

//var mysql_connect = "root:cmpe281@tcp(localhost:3307)/cmpe281"
var mysql_connect = "root:cmpe281@tcp(mysql:3307)/cmpe281"

func main() {

  db, err := sql.Open("mysql", mysql_connect)
  defer db.Close()

  failOnError( err, "unable to connect to db")

  sqlStr := "insert into short_links ( short_url ) values "
  vals := []string{}

  for i := int64(62); i < 3844; i++ {
    short_url := big.NewInt(i).Text(62)

    sqlStr += "( ? ),"
    vals = append( vals, short_url )
  }

  Shuffle( vals )

  sqlVals := []interface{}{}

  for _, d := range vals {
    sqlVals = append( sqlVals, d )
  }

  // fmt.Println( sqlVals )

  sqlStr = strings.TrimSuffix( sqlStr, "," )

  stmt, err := db.Prepare( sqlStr )
  failOnError( err, "Error preparing sql statement")

  res, err := stmt.Exec( sqlVals... )
  failOnError( err, "Error inserting values into db")

  fmt.Println( res )

}

// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// https://www.calhoun.io/how-to-shuffle-arrays-and-slices-in-go/
func Shuffle(vals []string) {
  r := rand.New(rand.NewSource(time.Now().Unix()))
  for len(vals) > 0 {
    n := len(vals)
    randIndex := r.Intn(n)
    vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
    vals = vals[:n-1]
  }
}
