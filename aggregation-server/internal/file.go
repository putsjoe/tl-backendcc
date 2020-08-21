package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type FileDetails struct {
	Filenames map[string]bool
	Port      int
}

type Files struct {
	State map[string]FileDetails
	Mu    sync.Mutex
}

func (f *Files) remove(instance string, filename string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	delete(f.State[instance].Filenames, filename)
}

func (f *Files) add(instance string, filename string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	f.State[instance].Filenames[filename] = true
}

func (f *Files) insert(hreq HelloRequest) bool {
	if _, ok := f.State[hreq.Instance]; !ok {
		f.State[hreq.Instance] = FileDetails{
			Filenames: make(map[string]bool),
			Port:      hreq.Port,
		}
		return true
	}
	// Return if the instance is new
	return false
}

func (f *Files) removeInstance(instance string) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	delete(f.State, instance)
}

func (f *Files) prepFiles() []byte {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	jr := map[string][]map[string]string{"files": make([]map[string]string, 0)}
	for i := range f.State {
		for f := range f.State[i].Filenames {
			a := map[string]string{"filename": f}
			jr["files"] = append(jr["files"], a)
		}
	}
	js, _ := json.Marshal(jr)
	return js
}

func (f *Files) Hello(w http.ResponseWriter, r *http.Request) {

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

func (f *Files) Files(w http.ResponseWriter, r *http.Request) {
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

func (f *Files) Bye(w http.ResponseWriter, r *http.Request) {
	var breq ByeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&breq); err != nil {
		log.Fatal(err)
	}
	f.removeInstance(breq.Instance)

}
