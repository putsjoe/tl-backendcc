package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type InstanceInfo struct {
	Filenames map[string]bool
	Port      int
}

type State struct {
	Instances map[string]InstanceInfo
	Mu        sync.Mutex
}

func (f *State) remove(instance string, filename string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	delete(f.Instances[instance].Filenames, filename)
}

func (f *State) add(instance string, filename string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	f.Instances[instance].Filenames[filename] = true
}

func (f *State) insert(hreq HelloRequest) bool {
	if _, ok := f.Instances[hreq.Instance]; !ok {
		f.Instances[hreq.Instance] = InstanceInfo{
			Filenames: make(map[string]bool),
			Port:      hreq.Port,
		}
		return true
	}
	// Return if the instance is new
	return false
}

func (f *State) removeInstance(instance string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	delete(f.Instances, instance)
}

func (f *State) prepFiles() []byte {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	jr := map[string][]map[string]string{"files": make([]map[string]string, 0)}
	for i := range f.Instances {
		for f := range f.Instances[i].Filenames {
			a := map[string]string{"filename": f}
			jr["files"] = append(jr["files"], a)
		}
	}
	js, _ := json.Marshal(jr)
	return js
}

func (f *State) Hello(w http.ResponseWriter, r *http.Request) {

	var hreq HelloRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hreq); err != nil {
		log.Println("[Decode Error]: ", err)
	}
	if !f.insert(hreq) {
		return
	}
	// If instance is new, then return list of current files from server
	location := "http://localhost:" + strconv.Itoa(hreq.Port) + "/files"
	resp, err := http.Get(location)
	if err != nil {
		log.Println("["+location+"/files GET]: ", err)
	}
	defer resp.Body.Close()

	var result map[string][]map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	for _, r := range result["files"] {
		f.add(hreq.Instance, r["filename"])
	}

}

func (f *State) Files(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		js := f.prepFiles()
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	var req []PutRequest
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

func (f *State) Bye(w http.ResponseWriter, r *http.Request) {
	var breq ByeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&breq); err != nil {
		log.Fatal(err)
	}
	f.removeInstance(breq.Instance)

}
