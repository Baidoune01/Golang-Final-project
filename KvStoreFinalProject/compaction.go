package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type CompactionManager struct {
	sstDir string
}

func NewCompactionManager(sstDir string) *CompactionManager {
	return &CompactionManager{sstDir: sstDir}
}

func (c *CompactionManager) Compact() error {
	files, err := filepath.Glob(filepath.Join(c.sstDir, "*.sst"))
	if err != nil {
		return err
	}

	sort.Strings(files)

	compactedFileName := filepath.Join(c.sstDir, fmt.Sprintf("compacted_%d.sst", time.Now().UnixNano()))
	compactedFile, err := os.Create(compactedFileName)
	if err != nil {
		return err
	}
	defer compactedFile.Close()

	keyOperations := make(map[string]MemTableEntry)

	for _, file := range files {
		if err := c.processSSTFile(file, keyOperations); err != nil {
			return err
		}
	}

	for key, entry := range keyOperations {
		if err := c.writeCompactedEntry(compactedFile, key, entry); err != nil {
			return err
		}
	}

	for _, file := range files {
		if file != compactedFileName {
			os.Remove(file)
		}
	}

	return nil
}

func (c *CompactionManager) processSSTFile(filePath string, keyOperations map[string]MemTableEntry) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		opMarker := make([]byte, 1)
		if _, err := file.Read(opMarker); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var keyLength uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLength); err != nil {
			return err
		}

		keyBuf := make([]byte, keyLength)
		if _, err := file.Read(keyBuf); err != nil {
			return err
		}
		currentKey := string(keyBuf)

		if opMarker[0] == 's' {
			var valueLength uint32
			if err := binary.Read(file, binary.LittleEndian, &valueLength); err != nil {
				return err
			}
			valueBuf := make([]byte, valueLength)
			if _, err := file.Read(valueBuf); err != nil {
				return err
			}
			keyOperations[currentKey] = MemTableEntry{Value: string(valueBuf), Op: OpSet}
		} else {
			keyOperations[currentKey] = MemTableEntry{Op: OpDelete}
		}
	}

	return nil
}

func (c *CompactionManager) writeCompactedEntry(file *os.File, key string, entry MemTableEntry) error {
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
