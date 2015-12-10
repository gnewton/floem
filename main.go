package main

import (
	"bytes"
	"fmt"
	"os"
	"net"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var DefaultTransport http.RoundTripper

func init() {
	rand.Seed(time.Now().UnixNano())
	ClientHostname, _ = os.Hostname()
	ClientIp =  GetLocalIP()
	DefaultTransport = &http.Transport{
        Dial: (&net.Dialer{
                Timeout:   107 * time.Second,
                KeepAlive: 200 * time.Second,
        }).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}


var wg sync.WaitGroup
var db *gorm.DB
var ClientHostname string
var ClientIp string
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
		"https://gcdocs.gc.ca/nrcan-rncan/llisapi.dll",

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
	url := &UrlRequest{Url: urlString,
		ClientHostname: ClientHostname,
		ClientIp: ClientIp,
	}
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
		timeout := time.Duration(150 * time.Second)
		client := http.Client{
			Timeout: timeout,
			Transport:DefaultTransport,
		}
		req, err := http.NewRequest("GET", url.Url, nil)
		if err != nil{
			results.Error = err.Error()
			log.Println(err)
			return
		}
		req.Header.Add("User-Agent", "floem")
		res, err := client.Do(req)
		t1 := time.Now()
		results.RequestTimeMillis = t1.Sub(t0).Nanoseconds() / Million
		if err != nil {
			results.Error = err.Error()
			log.Println("-------------")
			log.Println("zzzz ERROR::::: ")
			log.Println(err)
			log.Println("-------------")
			return
		}
		_, err = ioutil.ReadAll(res.Body)

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
		req, err = http.NewRequest("POST", url.Url, bytes.NewBuffer(postBody))
		req.Header.Add("User-Agent", "floem")
		req.Header.Add("Content-Type", "application/octet-stream")
		res, err = client.Do(req)
		//client := &http.Client{}

		//res, err = client.Do(req)
		t1 = time.Now()
		results.RequestTimeMillis = t1.Sub(t0).Nanoseconds() / Million
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

//From: http://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}
