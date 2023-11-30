package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// WAL represents a Write-Ahead Log for recording operations
type WAL struct {
	sync.Mutex
	file      *os.File
	watermark int64
}

// NewWAL creates and initializes a new WAL instance
func NewWAL() (*WAL, error) {
	file, err := os.OpenFile("wal.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &WAL{
		file: file,
	}, nil
}

// Append adds a new entry to the WAL
func (w *WAL) Append(entry string) error {
	w.Lock()
	defer w.Unlock()

	_, err := w.file.WriteString(fmt.Sprintf("%s\n", entry))
	return err
}

// UpdateWatermark updates the watermark in the WAL after flushing to SST
func (w *WAL) UpdateWatermark() error {
	w.Lock()
	defer w.Unlock()

	pos, err := w.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	w.watermark = pos
	flushedMark := fmt.Sprintf("flushed at %d\n", time.Now().Unix())
	_, err = w.file.WriteString(flushedMark)
	return err
}

// Close closes the WAL file
func (w *WAL) Close() error {
	w.Lock()
	defer w.Unlock()

	return w.file.Close()
}
