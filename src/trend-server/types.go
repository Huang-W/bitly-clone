/*
	Butly API ( Version 3.0 )
*/

package main

type trendServerEntry struct {
	Id string `json:"id" bson:"_id"`
	OrigUrl string
	Visits int
}

type shortlinkTrend struct {
	OrigUrl string
  ShortUrl string
  Visits int
}

type shortUrl struct {
	ShortUrl string
}

type shortlinkMsg struct {
	OrigUrl string
	ShortUrl string
}

type trendResult struct {
	OrigUrl string
	Visits int
}
