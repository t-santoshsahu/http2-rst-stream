package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
        "time"
)

var (
	listenPort = flag.Int("listen_port", 12345, "Server listening port")
	certFile   = flag.String("cert_file", "server.crt", "TLS server cert")
	keyFile    = flag.String("key_file", "server.key", "TLS server key file")
	numWorkers = flag.Int("num_workers", 1, "number of async workers")
)

func asyncWorker(workerId int, serverUrl *url.URL, req string) {
	// Abuse the checks on body size to send Stream RSTs
	// https://go.googlesource.com/net/+/master/http2/transport.go#1748
	client := initClient()
	for i := 0; ; i++ {
		time.Sleep(10*time.Second)
		request := &http.Request{
			URL:           serverUrl,
			ContentLength: int64(len(req)),
			Body:          io.NopCloser(bytes.NewReader([]byte(req))),
		}
		res, err := client.Do(request)
		log.Printf("[worker %d] request %d: res %v %s", workerId, i, res, err)
	}
	client.CloseIdleConnections()
}
func generateLargeString() string {
	res := ""
	for i:=0; i< 133333; i++ {
		res += "a"
	}
	return res
}

func main() {
	res := generateLargeString()
	server := initServer(res, int16(*listenPort))

	serverUrl, err := url.Parse(fmt.Sprintf("https://localhost:%d", int16(*listenPort)))
	if err != nil {
		log.Fatalf("failed to parse internal URL: %s", err)
	}

	for i := 0; i < *numWorkers; i++ {
		time.Sleep(10*time.Second)
		go asyncWorker(i, serverUrl, res)
	}

	for err := range server.errors {
		log.Printf("server error: %s", err)
	}
}
