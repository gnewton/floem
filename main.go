package main

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var wg sync.WaitGroup
var db *gorm.DB

func main() {
	var err error
	db, err = dbInit()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	postBodyBytes := makeBodyPostMBytes(3)
	urls := []string{
		"foo",
		"http://www.nrc.can/robots.txt",
		"http://www.nrc.ca/robots.txt",
		"http://192.168.0.1/",
		"http://192.168.0.11/",
	}
	for _, url := range urls {
		go getUrl(url, postBodyBytes)
		wg.Add(1)
	}
	wg.Wait()
}

const Million = 1000000

func getUrl(urlString string, postBody []byte) {
	defer wg.Done()
	url := &Url{Url: urlString}
	results := &Results{
		Url:              *url,
		Success:          false,
		TimeoutAtSeconds: -1}

	defer db.Create(results)
	done := make(chan bool)
	go func() {

		defer func() {
			done <- true
		}()

		t0 := time.Now()
		results.RequestTimeStamp = t0

		res, err := http.Get(url.Url)
		if err != nil {
			results.Error = err.Error()
			log.Println(err)
			return
		}
		_, err = ioutil.ReadAll(res.Body)
		t1 := time.Now()
		res.Body.Close()
		if err != nil {
			results.Error = err.Error()
			log.Println("-------------")
			log.Println(",ERROR::::: ")
			log.Println(err)
			log.Println("-------------")
			return
		}
		fmt.Printf("\nresponsecode=%d  %s\n", res.StatusCode, url)
		fmt.Printf("Time to respond: %v\n\n", t1.Sub(t0))

		t0 = time.Now()
		res, err = http.Post(url.Url, "application/octet-stream", bytes.NewBuffer(postBody))
		//client := &http.Client{}

		//res, err = client.Do(req)
		if err != nil {
			results.Error = err.Error()
			log.Println("-------------")
			log.Println("ERROR::::: ")
			log.Println(err)
			log.Println(res)
			log.Println("-------------")
			return
		}
		defer res.Body.Close()

		t1 = time.Now()
		results.RequestTimeMillis = t1.Sub(t0).Nanoseconds() / Million
		results.HttpStatus = res.StatusCode
		results.Success = true
		results.SetHttpMethodPost()
		headers := fmt.Sprintf("%v", res.Header)
		if len(headers) > 0 {
			headers = headers[4 : len(headers)-1]
			results.HttpHeader = headers
		}

		fmt.Printf("Time to respond: %v\n\n", t1.Sub(t0))
		fmt.Println("response Status:", res.Status)

		fmt.Println("response Headers:", headers)
		_, _ = ioutil.ReadAll(res.Body)

	}()

	select {
	case _ = <-done:

	case <-time.After(time.Second * 100):
		fmt.Println("timeout 100: " + urlString)
		results.Success = false
		results.TimeoutAtSeconds = 100
	}

	//fmt.Printf("%s", robots)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var postBodyCache map[int][]byte = make(map[int][]byte)

func makeBodyPostMBytes(n int) []byte {
	if postBody, ok := postBodyCache[n]; ok {
		return postBody
	}
	size := n * 1024 * 1024
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	postBodyCache[n] = b
	return b
}
