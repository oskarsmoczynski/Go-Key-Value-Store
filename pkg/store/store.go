package store

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Item struct {
	Value     string
	ExpiresAt time.Time
}

type AOFEntry struct {
	Op        string
	Key       string
	Value     string
	ExpiresAt time.Time
}

type Store struct {
	items       map[string]Item
	mu          sync.RWMutex
	aofFile     *os.File
	persistance Persistance
}

type Persistance interface {
	AOFAppend(file *os.File, entry AOFEntry) error
}

func New(aofPath string, persistence Persistance) (*Store, error) {
	absPath, err := filepath.Abs(aofPath)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	store := Store{
		items:       make(map[string]Item),
		aofFile:     f,
		persistance: persistence,
	}
	err = store.loadAOF()
	if err != nil {
		return nil, err
	}

	return &store, nil
}

func (s *Store) Set(key string, value string, ttlSeconds uint64, override bool) {
	var expiresAt time.Time
	if ttlSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}

	if !override {
		_, ok := s.Get(key)
		if ok {
			return
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = Item{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	if s.persistance != nil {
		for range 5 {
			err := s.persistance.AOFAppend(s.aofFile, AOFEntry{
				Op:        "set",
				Key:       key,
				Value:     value,
				ExpiresAt: expiresAt,
			})
			if err == nil {
				break
			}
		}
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[key]

	if !ok {
		return "", false
	}
	if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(time.Now()) {
		s.Delete(key)
		return "", false
	}
	return item.Value, true
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)

	if s.persistance != nil {
		for range 5 {
			err := s.persistance.AOFAppend(s.aofFile, AOFEntry{
				Op:  "delete",
				Key: key,
			})
			if err == nil {
				break
			}
		}
	}
}

func (s *Store) loadAOF() error {
	_, err := s.aofFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(s.aofFile)
	for scanner.Scan() {
		var entry AOFEntry
		err := json.Unmarshal(scanner.Bytes(), &entry)
		if err != nil {
			return err
		}

		// Skip expired entries
		if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(time.Now()) {
			continue
		}

		switch entry.Op {
		case "set":
			s.mu.Lock()
			s.items[entry.Key] = Item{
				Value:     entry.Value,
				ExpiresAt: entry.ExpiresAt,
			}
			s.mu.Unlock()
		case "delete":
			s.mu.Lock()
			delete(s.items, entry.Key)
			s.mu.Unlock()
		}
	}
	return nil
}
