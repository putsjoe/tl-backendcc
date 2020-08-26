package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"thirdlight.com/aggregation-server/internal"
)

func main() {
	port := flag.Int("port", 8000, "The port to run the Aggregator server on")
	flag.Parse()

	f := &internal.State{Instances: make(map[string]internal.InstanceInfo)}
	http.HandleFunc("/hello", f.Hello)
	http.HandleFunc("/bye", f.Bye)
	http.HandleFunc("/files", f.Files)

	log.Println("[INFO] Starting aggregator")
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
