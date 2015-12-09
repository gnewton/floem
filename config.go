package main

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"log"
	"time"
)

type Config struct {
	DbFile string
	Urls   []Url
}

type Url struct {
	Id                        int64  `gorm:"primary_key"`
	Url                       string `sql:"index"`
	HttpMethod                string `sql:"-"`
	RepeatTimeIntervalSeconds int    `sql:"-"`
	TimeoutSeconds            int    `sql:"-"`
	PostSizeMB                int    `sql:"-"`
}

type Results struct {
	Id               int64     `gorm:"primary_key"`
	RequestTimeStamp time.Time `sql:"index"`
	//Url               string `sql:"index"`
	Url        Url
	UrlId      sql.NullInt64
	HttpMethod string

	RequestTimeMillis int64
	HttpStatus        int `sql:"index"`

	Success          bool `sql:"index"`
	TimeoutAtSeconds int  `sql:"index"`
	PostSizeMB       int
	Error            string
	// Unsupported protocol
	// No such host
	// No route to host

	HttpHeader string
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
	db.DB().SetMaxOpenConns(100)

	db.CreateTable(&Results{})
	db.CreateTable(&Url{})
	return &db, nil
}
