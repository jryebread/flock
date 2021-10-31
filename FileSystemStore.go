package flock

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileSystemUserStore struct {
	DatabaseWriter  *json.Encoder
	DatabaseFile *os.File
}

func NewFileSystemUserStore(database *os.File) *FileSystemUserStore {
	database.Seek(0, 0)
	return &FileSystemUserStore{
		DatabaseWriter:  json.NewEncoder(&tape{database}),
		DatabaseFile: database,
	}
}

func FileSystemUserStoreFromFile(path string) (*FileSystemUserStore, func(), error) {
	db, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, nil, fmt.Errorf("problem opening %s %v", path, err)
	}

	closeFunc := func() {
		db.Close()
	}

	store := NewFileSystemUserStore(db)

	return store, closeFunc, nil
}