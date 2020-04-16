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

type shortlinkMsg struct {
	OrigUrl string
	ShortUrl string
}

type shortlinkDoc struct {
	OrigUrl string
	ShortUrl string
	visits int
}
