package main

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// CompactionManager is responsible for the compaction of SST files.
type CompactionManager struct {
	sstDir string // Directory where SST files are stored
}

// NewCompactionManager creates a new instance of a CompactionManager for a given SST directory.
func NewCompactionManager(sstDir string) *CompactionManager {
	return &CompactionManager{
		sstDir: sstDir,
	}
}

// Compact merges several SST files into one, removing outdated or deleted entries.
func (c *CompactionManager) Compact() error {
	// Get a list of all SST files
	files, err := filepath.Glob(filepath.Join(c.sstDir, "*.sst"))
	if err != nil {
		return err
	}

	sort.Strings(files) // Sort files to process them in order

	// Open a new SST file for the compacted data
	compactedFileName := filepath.Join(c.sstDir, "compacted.sst")
	compactedFile, err := os.Create(compactedFileName)
	if err != nil {
		return err
	}
	defer compactedFile.Close()

	// Process each SST file
	for _, file := range files {
		if err := c.processSSTFile(file, compactedFile); err != nil {
			return err
		}
	}

	// Optionally, remove the old SST files after successful compaction
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			// Log the error but continue with the process
		}
	}

	return nil
}

// processSSTFile reads an SST file and writes its content to the compacted SST file.
func (c *CompactionManager) processSSTFile(filePath string, compactedFile *os.File) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		// Read operation marker, key, and value
		opMarker := make([]byte, 1)
		if _, err := file.Read(opMarker); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Read key
		var keyLength uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLength); err != nil {
			return err
		}
		keyBuf := make([]byte, keyLength)
		if _, err := file.Read(keyBuf); err != nil {
			return err
		}

		// Read value if it's a SET operation
		if opMarker[0] == 's' {
			var valueLength uint32
			if err := binary.Read(file, binary.LittleEndian, &valueLength); err != nil {
				return err
			}
			valueBuf := make([]byte, valueLength)
			if _, err := file.Read(valueBuf); err != nil {
				return err
			}

			// Write the key-value pair to the compacted file
			if _, err := compactedFile.Write(opMarker); err != nil {
				return err
			}
			if err := binary.Write(compactedFile, binary.LittleEndian, keyLength); err != nil {
				return err
			}
			if _, err := compactedFile.Write(keyBuf); err != nil {
				return err
			}
			if err := binary.Write(compactedFile, binary.LittleEndian, valueLength); err != nil {
				return err
			}
			if _, err := compactedFile.Write(valueBuf); err != nil {
				return err
			}
		}
	}

	return nil
}
