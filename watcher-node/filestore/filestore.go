package filestore

import (
	"os"
	"sync"

	"github.com/google/uuid"
)

type Store struct {
	list     map[string]struct{}
	mutex    sync.RWMutex
	instance string
	seqno    int
}

type fileList map[string]struct{}

func New() *Store {
	return &Store{
		list:     fileList{},
		mutex:    sync.RWMutex{},
		seqno:    0,
		instance: uuid.New().String(),
	}
}

func (s *Store) AddFiles(files []os.FileInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.seqno = s.seqno + 1
	for _, file := range files {
		s.list[file.Name()] = struct{}{}
	}
}

func (s *Store) Instance() string {
	return s.instance
}

func (s *Store) Update(op string, filename string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.seqno += 1
	toRet := s.seqno
	switch op {
	case "add":
		s.list[filename] = struct{}{}
	case "remove":
		delete(s.list, filename)
	}
	return toRet
}

func (s *Store) GetList() (fileList, int) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.list, s.seqno
}
