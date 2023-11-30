package main

import (
	"sync"
)

type OperationType string

const (
	OpSet    OperationType = "SET"
	OpDelete OperationType = "DELETE"
)

type MemTableEntry struct {
	Value string
	Op    OperationType
}

type MemTable struct {
	sync.RWMutex
	table map[string]MemTableEntry
}

// NewMemTable creates a new MemTable instance.
func NewMemTable() *MemTable {
	return &MemTable{
		table: make(map[string]MemTableEntry),
	}
}

// Get retrieves a value for a given key from the MemTable.
// The bool returned represents whether the key was found.
func (m *MemTable) Get(key string) (string, OperationType, bool) {
	m.RLock()
	defer m.RUnlock()
	entry, ok := m.table[key]
	return entry.Value, entry.Op, ok
}

// Set adds a key-value pair to the MemTable.
func (m *MemTable) Set(key string, value string) {
	m.Lock()
	defer m.Unlock()
	m.table[key] = MemTableEntry{Value: value, Op: OpSet}
}

// Delete marks a key-value pair as deleted in the MemTable.
func (m *MemTable) Delete(key string) {
	m.Lock()
	defer m.Unlock()
	// If key exists, mark as deleted. If not, still add a delete marker.
	m.table[key] = MemTableEntry{Op: OpDelete}
}

// ShouldFlush checks if the MemTable should be flushed based on a threshold.
func (m *MemTable) ShouldFlush() bool {
	m.RLock()
	defer m.RUnlock()
	return len(m.table) >= MemTableFlushThreshold
}

// Clear clears the MemTable.
func (m *MemTable) Clear() {
	m.Lock()
	defer m.Unlock()
	m.table = make(map[string]MemTableEntry)
}
