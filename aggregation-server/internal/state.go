package internal

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"

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
}

type FilesResponse struct {
	Files []lib.FileMetadata `json:"files"`
}

type FileN struct {
	instance string
	filename string
}

type AddI struct {
	hreq lib.HelloOperation
	url  string
	resp chan bool
}

type DataR struct {
	resp chan FilesResponse
}

type OperationChannels struct {
	AddInstanceChan    chan AddI
	AddFileChan        chan FileN
	RemoveFileChan     chan FileN
	RemoveInstanceChan chan string
	PrepResponseChan   chan DataR
}

func StateSetup(oc OperationChannels) {
	f := &State{Instances: make(map[string]InstanceInfo)}
	for {
		select {
		case add := <-oc.AddInstanceChan:
			if _, ok := f.Instances[add.hreq.Instance]; !ok {
				f.Instances[add.hreq.Instance] = InstanceInfo{
					Filenames: make(map[string]bool),
					URL:       add.url,
					Port:      int(add.hreq.Port),
				}
				add.resp <- true
			}
			add.resp <- false

		case a := <-oc.AddFileChan:
			f.Instances[a.instance].Filenames[a.filename] = true

		case rf := <-oc.RemoveFileChan:
			delete(f.Instances[rf.instance].Filenames, rf.filename)

		case instance := <-oc.RemoveInstanceChan:
			if _, ok := f.Instances[instance]; ok {
				delete(f.Instances, instance)
			}

		case data := <-oc.PrepResponseChan:
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

			data.resp <- FilesResponse{Files: fr}
		}
	}
}

/*
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
*/

func (oc OperationChannels) Hello(w http.ResponseWriter, r *http.Request) {
	var hreq lib.HelloOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hreq); err != nil {
		log.Println("[Decode Error]: ", err)
	}

	addInst := AddI{
		hreq: hreq,
		url:  r.RemoteAddr,
		resp: make(chan bool)}
	oc.AddInstanceChan <- addInst

	if !<-addInst.resp {
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
		// Also remove this instance as it is no longer working?
	}
	defer resp.Body.Close()

	var result FilesResponse
	json.NewDecoder(resp.Body).Decode(&result)
	for _, r := range result.Files {
		oc.AddFileChan <- FileN{
			instance: hreq.Instance,
			filename: r.Filename}
	}
}

func (oc OperationChannels) Files(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getjson := DataR{resp: make(chan FilesResponse)}
		oc.PrepResponseChan <- getjson
		js := <-getjson.resp
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
			oc.AddFileChan <- FileN{
				instance: p.Instance,
				filename: p.Value.Filename}
		case remove:
			oc.RemoveFileChan <- FileN{
				instance: p.Instance,
				filename: p.Value.Filename}
		}
	}
}

func (oc OperationChannels) Bye(w http.ResponseWriter, r *http.Request) {
	var breq lib.ByeOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&breq); err != nil {
		log.Fatal(err)
	}

	oc.RemoveInstanceChan <- breq.Instance

}
