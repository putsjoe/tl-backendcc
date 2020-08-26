package internal

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"thirdlight.com/watcher-node/lib"
)

const (
	add    = "add"
	remove = "remove"
)

type InstanceInfo struct {
	Filenames map[string]bool
	URL       string
	Port      int
}

type State struct {
	Instances map[string]InstanceInfo
	Mu        sync.Mutex
}

type FilesResponse struct {
	Files []lib.FileMetadata `json:"files"`
}

func (f *State) removeFile(instance string, filename string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	delete(f.Instances[instance].Filenames, filename)
}

func (f *State) addFile(instance string, filename string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	f.Instances[instance].Filenames[filename] = true
}

func (f *State) addInstance(hreq lib.HelloOperation, url string) bool {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	if _, ok := f.Instances[hreq.Instance]; !ok {
		f.Instances[hreq.Instance] = InstanceInfo{
			Filenames: make(map[string]bool),
			URL:       url,
			Port:      int(hreq.Port),
		}
		return true
	}
	return false
}

func (f *State) removeInstance(instance string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	if _, ok := f.Instances[instance]; ok {
		delete(f.Instances, instance)
	}
}

func (f *State) prepResponse() FilesResponse {
	f.Mu.Lock()
	defer f.Mu.Unlock()

	fr := make([]lib.FileMetadata, 0)
	us := make([]string, 0)

	for i := range f.Instances {
		for f := range f.Instances[i].Filenames {
			us = append(us, f)
		}
	}

	sort.Strings(us)
	for _, s := range us {
		fr = append(fr, lib.FileMetadata{s})
	}

	return FilesResponse{Files: fr}
}

func (f *State) Hello(w http.ResponseWriter, r *http.Request) {

	var hreq lib.HelloOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hreq); err != nil {
		log.Println("[Decode Error]: ", err)
	}
	if !f.addInstance(hreq, r.RemoteAddr) {
		return
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("splitHost: %v", err)
	}

	// Not sure this is right but getting an error just using r.RemoteAddr
	location := "http://" + host + ":" + strconv.FormatUint(uint64(hreq.Port), 10) + "/files"
	resp, err := http.Get(location)
	if err != nil {
		log.Println("["+location+"/files GET]: ", err)
	}
	defer resp.Body.Close()

	var result FilesResponse
	json.NewDecoder(resp.Body).Decode(&result)
	for _, r := range result.Files {
		f.addFile(hreq.Instance, r.Filename)
	}

}

func (f *State) Files(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		js := f.prepResponse()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(js)
		return
	}

	var req []lib.PatchOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Fatal(err)
	}
	for _, p := range req {
		switch p.Op {
		case add:
			f.addFile(p.Instance, p.Value.Filename)
		case remove:
			f.removeFile(p.Instance, p.Value.Filename)
		}
	}
}

func (f *State) Bye(w http.ResponseWriter, r *http.Request) {
	var breq lib.ByeOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&breq); err != nil {
		log.Fatal(err)
	}
	f.removeInstance(breq.Instance)

}
