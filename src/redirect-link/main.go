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
	"os"
)

func main() {

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3001"
	}

	server := NewServer()
	server.Run(":" + port)
}
