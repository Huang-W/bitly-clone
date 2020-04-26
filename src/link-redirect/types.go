/*
	Butly API ( Version 3.0 )
*/

package main

type shortlinkMsg struct {
	OrigUrl string
	ShortUrl string
}

type shortUrl struct {
	ShortUrl string
}

type originalUrl struct {
	OrigUrl string
}

type shortlinkDoc struct {
	OrigUrl string
	ShortUrl string
	visits int
}
