package server

import (
	"encoding/json"
	"log"
	"net/http"

	"thirdlight.com/watcher-node/filestore"
	"thirdlight.com/watcher-node/lib"
)

func FilesHandler(store *filestore.Store) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(r.Method == http.MethodGet) {
			log.Println("[ERROR] invalid request method :", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		filesMeta := []lib.FileMetadata{}
		list, seqNo := store.GetList()
		for name, _ := range list {
			filesMeta = append(filesMeta, lib.FileMetadata{
				Filename: name,
			})
		}
		json.NewEncoder(w).Encode(lib.ListResponse{
			Files:       filesMeta,
			Sequence:    seqNo,
			BaseMessage: lib.BaseMessage{store.Instance()},
		})
	})
}
