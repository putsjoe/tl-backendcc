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

	// Create operation channels
	ch := internal.OperationChannels{
		AddInstanceChan:    make(chan internal.AddI),
		AddFileChan:        make(chan internal.FileN),
		RemoveFileChan:     make(chan internal.FileN),
		RemoveInstanceChan: make(chan string)}
	go internal.StateSetup(ch)

	// f := &internal.State{Instances: make(map[string]internal.InstanceInfo)}
	http.HandleFunc("/hello", ch.Hello)
	http.HandleFunc("/bye", ch.Bye)
	http.HandleFunc("/files", ch.Files)

	log.Println("[INFO] Starting aggregator")

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))

}
