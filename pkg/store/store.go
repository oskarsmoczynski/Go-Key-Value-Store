package store

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/persistance"
	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/util"
)

type Item struct {
	Value     string
	ExpiresAt time.Time
}

type Store struct {
	items               map[string]Item
	mu                  sync.RWMutex
	aofFile             *os.File
	snapshotDir         string
	aofPersistance      AOFPersistance
	snapshotPersistance SnapshotPersistance
}

type AOFPersistance interface {
	AOFAppend(file *os.File, entry persistance.AOFEntry) error
}

type SnapshotPersistance interface {
	SaveSnapshot(dir string, entries []persistance.SnapshotEntry) error
	LoadSnapshot(dir string) ([]persistance.SnapshotEntry, error)
}

func New(aofPath string, snapshotPath string) (*Store, error) {
	aofFile, err := util.OpenOrCreate(aofPath)
	if err != nil {
		return nil, err
	}
	snapshotDir, err := util.MakeDirs(snapshotPath)
	if err != nil {
		return nil, err
	}
	store := Store{
		items:               make(map[string]Item),
		aofFile:             aofFile,
		snapshotDir:         snapshotDir,
		aofPersistance:      persistance.NewAOFPersistance(),
		snapshotPersistance: persistance.NewSnapshotPersistance(),
	}
	if err = store.LoadSnapshot(); err != nil {
		return nil, err
	}
	if err = store.loadAOF(); err != nil {
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

	if s.aofPersistance != nil {
		for range 5 {
			err := s.aofPersistance.AOFAppend(s.aofFile, persistance.AOFEntry{
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

	if s.aofPersistance != nil {
		for range 5 {
			err := s.aofPersistance.AOFAppend(s.aofFile, persistance.AOFEntry{
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
		var entry persistance.AOFEntry
		err := json.Unmarshal(scanner.Bytes(), &entry)
		if err != nil {
			return err
		}

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

func (s *Store) SaveSnapshot() error {
	entries := make([]persistance.SnapshotEntry, 0, len(s.items))
	for k, v := range s.items {
		entries = append(entries, persistance.SnapshotEntry{
			Key:       k,
			Value:     v.Value,
			ExpiresAt: v.ExpiresAt,
		})
	}
	return s.snapshotPersistance.SaveSnapshot(s.snapshotDir, entries)
}

func (s *Store) LoadSnapshot() error {
	entries, err := s.snapshotPersistance.LoadSnapshot(s.snapshotDir)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range entries {
		if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(time.Now()) {
			continue
		}

		s.items[entry.Key] = Item{
			Value:     entry.Value,
			ExpiresAt: entry.ExpiresAt,
		}
	}
	return nil
}
