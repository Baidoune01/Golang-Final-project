package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
)

// SSTManager manages SST files
type SSTManager struct {
	sync.Mutex
	sstDir         string
	sstFileCounter int
}

// NewSSTManager creates a new SSTManager instance
func NewSSTManager(sstDir string) *SSTManager {
	os.MkdirAll(sstDir, os.ModePerm)
	return &SSTManager{
		sstDir:         sstDir,
		sstFileCounter: 0,
	}
}

// FlushMemTableToSST flushes contents of MemTable to an SST file
func (s *SSTManager) FlushMemTableToSST(memTable *MemTable) error {
	s.Lock()
	defer s.Unlock()

	s.sstFileCounter++
	fileName := s.sstDir + "/sst_" + fmt.Sprint(s.sstFileCounter) + ".sst"

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	for key, entry := range memTable.table {
		if err := writeKeyValuePairWithOperation(file, key, entry); err != nil {
			return err
		}
	}

	return s.checkAndTriggerCompaction()
}

func (s *SSTManager) checkAndTriggerCompaction() error {
	files, err := os.ReadDir(s.sstDir)
	if err != nil {
		return fmt.Errorf("error reading SST directory: %v", err)
	}

	if len(files) >= 4 {
		compactionManager := NewCompactionManager(s.sstDir)
		if err := compactionManager.Compact(); err != nil {
			return fmt.Errorf("compaction error: %v", err)
		}
	}

	return nil
}

// writeKeyValuePairWithOperation writes a key-value pair with operation to the file
func writeKeyValuePairWithOperation(file *os.File, key string, entry MemTableEntry) error {
	var opMarker byte
	if entry.Op == OpSet {
		opMarker = 's'
	} else {
		opMarker = 'd'
	}

	if _, err := file.Write([]byte{opMarker}); err != nil {
		return err
	}

	keyLen := uint32(len(key))
	if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
		return err
	}
	if _, err := file.Write([]byte(key)); err != nil {
		return err
	}

	if entry.Op == OpSet {
		valueLen := uint32(len(entry.Value))
		if err := binary.Write(file, binary.LittleEndian, valueLen); err != nil {
			return err
		}
		if _, err := file.Write([]byte(entry.Value)); err != nil {
			return err
		}
	}

	return nil
}

func (s *SSTManager) GetFromSST(key string) (string, bool) {
	s.Lock()
	defer s.Unlock()

	files, err := os.ReadDir(s.sstDir)
	if err != nil {
		return "", false
	}

	for i := len(files) - 1; i >= 0; i-- {
		fileInfo := files[i]
		if fileInfo.IsDir() {
			continue
		}

		file, err := os.Open(s.sstDir + "/" + fileInfo.Name())
		if err != nil {
			continue
		}

		found, value, deleted := processFile(file, key)
		file.Close()

		if found {
			if deleted {
				return "", false // Key was marked as deleted
			}
			return value, true
		}
	}

	return "", false
}

func processFile(file *os.File, key string) (bool, string, bool) {
	for {
		opMarker := make([]byte, 1)
		if _, err := file.Read(opMarker); err != nil {
			if err == io.EOF {
				break
			}
			return false, "", false
		}

		var keyLength uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLength); err != nil {
			if err == io.EOF {
				break
			}
			return false, "", false
		}

		keyBuf := make([]byte, keyLength)
		if _, err := file.Read(keyBuf); err != nil {
			if err == io.EOF {
				break
			}
			return false, "", false
		}
		currentKey := string(keyBuf)

		if currentKey == key {
			if opMarker[0] == 'd' {
				return true, "", true // Key was marked as deleted
			}

			var valueLength uint32
			if err := binary.Read(file, binary.LittleEndian, &valueLength); err != nil {
				return false, "", false
			}

			valueBuf := make([]byte, valueLength)
			if _, err := file.Read(valueBuf); err != nil {
				return false, "", false
			}

			return true, string(valueBuf), false
		}

		// If it's a SET operation, skip the value part
		if opMarker[0] == 's' {
			var valueLength uint32
			if err := binary.Read(file, binary.LittleEndian, &valueLength); err != nil {
				return false, "", false
			}
			if _, err := file.Seek(int64(valueLength), io.SeekCurrent); err != nil {
				return false, "", false
			}
		}
	}
	return false, "", false
}
