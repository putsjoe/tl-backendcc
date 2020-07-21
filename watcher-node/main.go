package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"thirdlight.com/watcher-node/aggregator"
	"thirdlight.com/watcher-node/filestore"
	"thirdlight.com/watcher-node/server"
)

const (
	mountedDir = "/host/watched-folder"
	add        = "add"
	remove     = "remove"
)

var defaultPort uint = 4000

func main() {
	log.Println("[INFO] Starting watcher node")
	var directory = flag.String("dir", mountedDir, "the path of the directory to watch")
	var port = flag.Uint("p", defaultPort, "the port")
	var aggregationServer = flag.String("aggregator", "", "the aggregation server address")
	flag.Parse()

	aggregatorClient, err := aggregator.New(&http.Client{}, *aggregationServer)
	if err != nil {
		log.Fatalln("[ERROR]", err)
	}

	store, err := initializeStoreForDirectory(*directory)
	if err != nil {
		log.Fatalln("[ERROR]", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/files", server.FilesHandler(store))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("[ERROR]", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				handleEvent(event, store, aggregatorClient)
			case err := <-watcher.Errors:
				log.Println("[ERROR]", err)
			}
		}
	}()

	if err := watcher.Add(*directory); err != nil {
		log.Println("[ERROR]", err)
	}
	log.Println("[INFO] Now watching", *directory)

	if port == nil {
		port = &defaultPort
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	go func() {
		<-sigChan
		srv.Close()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Println(srv.ListenAndServe())
		wg.Done()
	}()

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			helloErr := aggregatorClient.Hello(store.Instance(), *port)
			if helloErr != nil {
				log.Println("[ERROR]", helloErr)
			}
		}
	}()
	defer aggregatorClient.Bye(store.Instance())

	wg.Wait()
	close(sigChan)
	ticker.Stop()
}

func initializeStoreForDirectory(directory string) (*filestore.Store, error) {
	store := filestore.New()

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	store.AddFiles(files)
	return store, nil
}

func handleEvent(
	event fsnotify.Event,
	store *filestore.Store,
	aggregator *aggregator.Aggregator,
) {
	op := getOp(event)
	if op == "" {
		return
	}
	filename := filepath.Base(event.Name)

	opSeqNo := store.Update(op, filename)

	err := aggregator.NotifyUpdate(op, filename, opSeqNo, store.Instance())
	if err != nil {
		log.Println("[ERROR]: ", err)
	}
}

func getOp(event fsnotify.Event) string {
	switch event.Op {
	case fsnotify.Create:
		return add
	case fsnotify.Remove, fsnotify.Rename:
		return remove
	}
	return ""
}
