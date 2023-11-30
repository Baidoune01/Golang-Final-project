package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var memTable *MemTable
var wal *WAL
var sstManager *SSTManager

const MemTableFlushThreshold = 10

func main() {
	memTable = NewMemTable()
	wal, _ = NewWAL()
	sstManager = NewSSTManager("./sst")

	router := http.NewServeMux()
	router.HandleFunc("/get", handleGet)
	router.HandleFunc("/set", handleSet)
	router.HandleFunc("/del", handleDelete)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	memTable.RLock()
	value, op, found := memTable.Get(key)
	memTable.RUnlock()

	if !found {
		value, found = sstManager.GetFromSST(key)
		if !found {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
	}

	if op == OpDelete {
		http.Error(w, "Key was deleted", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var data map[string]string
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	// Check if at least one key-value pair is provided and value is not empty
	if len(data) == 0 {
		http.Error(w, "At least one key-value pair is required", http.StatusBadRequest)
		return
	}

	for key, value := range data {
		if value == "" {
			http.Error(w, fmt.Sprintf("Value for key '%s' is required", key), http.StatusBadRequest)
			return
		}

		wal.Append(fmt.Sprintf("SET %s %s", key, value))
		memTable.Set(key, value)

		if shouldFlushMemTable() {
			sstManager.FlushMemTableToSST(memTable)
			memTable.Clear()
			wal.UpdateWatermark()
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Key-value pair set successfully"))
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Check if the key exists in MemTable
	memTable.RLock()
	_, op, foundInMemTable := memTable.Get(key)
	memTable.RUnlock()

	// Check if the key exists in SST
	_, foundInSST := sstManager.GetFromSST(key)

	// If key is not found in MemTable and not found in SST, return error
	if !foundInMemTable && !foundInSST {
		http.Error(w, "Key does not exist", http.StatusNotFound)
		return
	}

	// If key is found but already marked as deleted, return key deleted
	if foundInMemTable && op == OpDelete {
		http.Error(w, "Key already deleted", http.StatusNotFound)
		return
	}

	// Proceed with delete operation
	wal.Append(fmt.Sprintf("DEL %s", key))
	memTable.Delete(key)

	if shouldFlushMemTable() {
		sstManager.FlushMemTableToSST(memTable)
		memTable.Clear()
		wal.UpdateWatermark()
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Key deleted successfully"))
}

func shouldFlushMemTable() bool {
	memTable.RLock()
	defer memTable.RUnlock()
	return len(memTable.table) >= MemTableFlushThreshold
}
