package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type fileDetails struct {
	filenames map[string]bool
	port      int
}

type files struct {
	state map[string]fileDetails
	mu    sync.Mutex
}

type helloRequest struct {
	Instance string `json:"instance"`
	Port     int    `json:"port"`
}

type byeRequest struct {
	Instance string `json:"instance"`
}

type putRequest struct {
	Instance  string            `json:"instance"`
	Operation string            `json:"op"`
	Sequence  int               `json:"seqno"`
	Value     map[string]string `json:"value"`
}

func main() {
	f := &files{state: make(map[string]fileDetails)}
	http.HandleFunc("/hello", f.hello)
	http.HandleFunc("/bye", f.bye)
	http.HandleFunc("/files", f.files)

	fmt.Println("Serving on port 8000...")
	http.ListenAndServe(":8000", nil)
}

func (f *files) remove(instance string, filename string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.state[instance].filenames, filename)
}

func (f *files) add(instance string, filename string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state[instance].filenames[filename] = true
}

func (f *files) insert(hreq helloRequest) bool {
	// Return whether or not the instance is new
	if _, ok := f.state[hreq.Instance]; !ok {
		f.state[hreq.Instance] = fileDetails{
			filenames: make(map[string]bool),
			port:      hreq.Port,
		}
		return true
	}
	return false
}

func (f *files) hello(w http.ResponseWriter, r *http.Request) {

	var hreq helloRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hreq); err != nil {
		log.Fatal(err)
	}
	if !f.insert(hreq) {
		return
	}
	// If new, then return list of current files from server
	// This shouldnt be hardcoded.
	location := "http://localhost:" + strconv.Itoa(hreq.Port) + "/files"
	resp, err := http.Get(location)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result map[string][]map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	for _, r := range result["files"] {
		f.add(hreq.Instance, r["filename"])
	}

}

func (f *files) removeInstance(instance string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.state, instance)
}

func (f *files) bye(w http.ResponseWriter, r *http.Request) {
	var breq byeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&breq); err != nil {
		log.Fatal(err)
	}
	f.removeInstance(breq.Instance)

}

func (f *files) prepFiles() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	jr := map[string][]map[string]string{"files": make([]map[string]string, 0)}
	for i := range f.state {
		for f := range f.state[i].filenames {
			a := map[string]string{"filename": f}
			jr["files"] = append(jr["files"], a)
		}
	}
	js, _ := json.Marshal(jr)
	return js
}

func (f *files) files(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		js := f.prepFiles()
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	var req []putRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Fatal(err)
	}
	for _, p := range req {
		switch p.Operation {
		case "add":
			f.add(p.Instance, p.Value["filename"])
		case "remove":
			f.remove(p.Instance, p.Value["filename"])
		}
	}
}

// TODO
// Seperate out
// Move notes to bottom of readme

/** Improvements **/
// Sorting of filenames returned
// hello function more concurrency
// data races
// Alternative to mutex

/** Questions: **/
// prepFiles - better solution and sorting.
// files function - Too short for switch statement?
// Overused mutex? RWMutex? Best alternative?
/*
Which is the better syntax:
  var state = make(map[string]map[string]bool)
  var state = map[string]map[string]bool{}
*/
/*
	Using the mutex right at the top can just block the whole thing on a http
	handler function.
/*
make stop
pkill -f watcher-node
2020/08/20 22:33:13 http: Server closed
2020/08/20 22:33:13 http: Server closed
joe@joe:~/thirdlight-bcc$ 2020/08/20 22:33:13 [ERROR] <nil>
*/

/** References **
https://www.alexedwards.net/blog/understanding-mutexes
https://golang.org/doc/articles/race_detector.html
https://golang.org/doc/effective_go.html#maps
https://gobyexample.com/
https://play.golang.org/p/Y7gn8_0cKJm
*/
