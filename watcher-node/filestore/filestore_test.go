package filestore

import (
	"reflect"
	"testing"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		scenario          string
		op                string
		filename          string
		actualStoreList   fileList
		expectedStoreList fileList
	}{
		{
			scenario:        "add file",
			op:              "add",
			filename:        "file.txt",
			actualStoreList: fileList{},
			expectedStoreList: fileList{
				"file.txt": {},
			},
		},
		{
			scenario: "delete file",
			op:       "remove",
			filename: "file.txt",
			actualStoreList: fileList{
				"file.txt": {},
			},
			expectedStoreList: fileList{},
		},
		{
			scenario: "unknown op",
			op:       "move",
			filename: "blah.txt",
			actualStoreList: fileList{
				"file.txt": {},
			},
			expectedStoreList: fileList{
				"file.txt": {},
			},
		},
	}

	for _, test := range tests {
		store := Store{
			list: test.actualStoreList,
		}
		store.Update(test.op, test.filename)
		if !reflect.DeepEqual(store.GetList(), test.expectedStoreList) {
			t.Errorf(
				"%s, expected: %v, got: %v",
				test.scenario,
				test.expectedStoreList,
				store.GetList(),
			)
		}
	}
}
