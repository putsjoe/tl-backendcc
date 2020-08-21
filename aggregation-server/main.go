package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"thirdlight.com/aggregation-server/internal"
)

func main() {
	f := &internal.Files{State: make(map[string]internal.FileDetails)}
	http.HandleFunc("/hello", f.Hello)
	http.HandleFunc("/bye", f.Bye)
	http.HandleFunc("/files", f.Files)

	port := flag.Int("port", 8000, "The port to run the Aggregator server on")
	flag.Parse()

	log.Println("[INFO] Starting aggregator")
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}

