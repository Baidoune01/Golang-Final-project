package main

import (
	"encoding/gob"
	"os"
	"sync"
)

// DiskStorage is an implementation of Storage interface that persists data to disk.
type DiskStorage struct {
	sync.RWMutex
	filePath string
	data     map[string]string
}

// NewDiskStorage creates a new DiskStorage instance and initializes it with data from the disk if it exists.
func NewDiskStorage(filePath string) (*DiskStorage, error) {
	ds := &DiskStorage{
		filePath: filePath,
		data:     make(map[string]string),
	}

	// Try to load existing data from disk
	if err := ds.load(); err != nil {
		return nil, err
	}

	return ds, nil
}

// Read retrieves the value for a key from the storage.
func (ds *DiskStorage) Read(key string) (string, bool) {
	ds.RLock()
	defer ds.RUnlock()
	value, ok := ds.data[key]
	return value, ok
}

// Write sets the value for a key in the storage.
func (ds *DiskStorage) Write(key string, value string) error {
	ds.Lock()
	defer ds.Unlock()

	ds.data[key] = value
	return ds.save()
}

// Delete removes a key from the storage.
func (ds *DiskStorage) Delete(key string) error {
	ds.Lock()
	defer ds.Unlock()

	delete(ds.data, key)
	return ds.save()
}

// save persists the current state of the storage to disk.
func (ds *DiskStorage) save() error {
	file, err := os.Create(ds.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(ds.data)
}

// load reads data from disk and loads it into the storage.
func (ds *DiskStorage) load() error {
	file, err := os.Open(ds.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // It's okay if the file doesn't exist yet
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&ds.data)
}
