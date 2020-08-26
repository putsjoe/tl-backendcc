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

type FilesResponse struct {
	Files []FileMetadata `json:"files"`
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

func (f *State) addInstance(hreq HelloOperation) bool {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	if _, ok := f.Instances[hreq.Instance]; !ok {
		f.Instances[hreq.Instance] = InstanceInfo{
			Filenames: make(map[string]bool),
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

	fr := FilesResponse{
		Files: make([]FileMetadata, 0),
	}

	for i := range f.Instances {
		for f := range f.Instances[i].Filenames {
			fr.Files = append(fr.Files, FileMetadata{f})
		}
	}

	return fr
}

func (f *State) Hello(w http.ResponseWriter, r *http.Request) {

	var hreq HelloOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hreq); err != nil {
		log.Println("[Decode Error]: ", err)
	}
	if !f.addInstance(hreq) {
		return
	}
	// If instance is new, then return list of current files from server
	location := "http://localhost:" + strconv.FormatUint(uint64(hreq.Port), 10) + "/files"
	resp, err := http.Get(location)
	if err != nil {
		log.Println("["+location+"/files GET]: ", err)
	}
	defer resp.Body.Close()

	var result map[string][]map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	for _, r := range result["files"] {
		f.addFile(hreq.Instance, r["filename"])
	}

}

func (f *State) Files(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		js := f.prepResponse()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(js)
		return
	}

	var req []PatchOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Fatal(err)
	}
	for _, p := range req {
		switch p.Op {
		case "add":
			f.addFile(p.Instance, p.Value.Filename)
		case "remove":
			f.removeFile(p.Instance, p.Value.Filename)
		}
	}
}

func (f *State) Bye(w http.ResponseWriter, r *http.Request) {
	var breq ByeOperation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&breq); err != nil {
		log.Fatal(err)
	}
	f.removeInstance(breq.Instance)

}
