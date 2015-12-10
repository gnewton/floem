package main

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"log"
	"time"
)

type Config struct {
	DbFile string
	Urls   []UrlRequest
}

type UrlRequest struct {
	Id                        int64  `gorm:"primary_key"`
	ClientIp                  string `sql:"index"` // json ignore
	ClientHostname            string `sql:"index"` //json ignore
	Url                       string `sql:"index"`
	HttpMethod                string `sql:"-"`
	RepeatTimeIntervalSeconds int    `sql:"-"`
	TimeoutSeconds            int    `sql:"-"`
	PostSizeMb                int    `sql:"-"`
}

type Results struct {
	Id               int64     `gorm:"primary_key"`
	RequestTimeStamp time.Time `sql:"index"`
	//Url               string `sql:"index"`
	Url        UrlRequest
	UrlId      sql.NullInt64
	HttpMethod string

	RequestTimeMillis int64
	HttpStatus        int `sql:"index"`

	Success          bool `sql:"index"`
	TimeoutAtSeconds int  `sql:"index"`
	PostSizeMb       int
	HttpHeader       string
	Error            string
	// Unsupported protocol
	// No such host
	// No route to host

}

func (r *Results) SetHttpMethodGet() {
	r.HttpMethod = "GET"
}

func (r *Results) SetHttpMethodPost() {
	r.HttpMethod = "POST"
}

func dbInit() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "/tmp/zazl.sqlite3")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//db.LogMode(true)

	db.DB()
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(500)

	db.CreateTable(&Results{})
	db.CreateTable(&UrlRequest{})

	//
	db.Exec("PRAGMA locking_mode = EXCLUSIVE;")
	db.Exec("PRAGMA count_changes = OFF;")
	db.Exec("PRAGMA page_size = 4096;")
	// indexes
	db.Table("results").AddIndex("request_time_stamp", "request_time_stamp")
	db.Table("results").AddUniqueIndex("url_id", "url_id")
	db.Table("results").AddIndex("http_status", "http_status")
	db.Table("results").AddIndex("success", "success")
	db.Table("url_requests").AddIndex("client_ip", "client_ip")
	db.Table("url_requests").AddIndex("url", "url")

	return &db, nil
}
