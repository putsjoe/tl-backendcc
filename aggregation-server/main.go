package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type files struct {
	state map[string][]string
	mu    sync.Mutex
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
	f := &files{state: make(map[string][]string)}
	http.HandleFunc("/hello", nothing)
	http.HandleFunc("/bye", f.bye)
	http.HandleFunc("/files", f.files)

	fmt.Println("Serving on port 8000...")
	http.ListenAndServe(":8000", nil)
}

func (f *files) remove(instance string, filename string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if val, ok := f.state[instance]; ok {
		for i, e := range val {
			if e == filename {
				val[i] = ""
			}
		}
	}

}

func (f *files) add(instance string, filename string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.state[instance]; ok {
		f.state[instance] = append(f.state[instance], filename)
	} else {
		f.state[instance] = []string{filename}
	}
}

func (f *files) bye(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var req byeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Fatal(err)
	}
	if _, ok := f.state[req.Instance]; ok {
		delete(f.state, req.Instance)
	}

}

func (f *files) files(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Method == http.MethodGet {
		var filenames []string

		for i := range f.state {
			for _, f := range f.state[i] {
				filenames = append(filenames, f)
			}
		}
		for i := range filenames {
			fmt.Fprintf(w, filenames[i]+"\n")
		}

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

func print(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

	return

}

func nothing(w http.ResponseWriter, r *http.Request) {
	return
}

// TODO
// Proper JSON response for files GET endpoint
// remove function - Dont just blank, remove from slice
// Think about alternative to using a slice?
// Move to seperate file?

// Endpoints: files GET, files POST, hello, bye

// Notes:

// files function - Too short for switch statement?

/*
make stop
pkill -f watcher-node
2020/08/20 22:33:13 http: Server closed
2020/08/20 22:33:13 http: Server closed
joe@joe:~/thirdlight-bcc$ 2020/08/20 22:33:13 [ERROR] <nil>
*/
