package main

import (
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
