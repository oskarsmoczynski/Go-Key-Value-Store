package store

import (
	"fmt"
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
	LoadAOF(file *os.File) ([]persistance.AOFEntry, error)
	ClearAOF(file *os.File) error
}

type SnapshotPersistance interface {
	SaveSnapshot(dir string, entries []persistance.SnapshotEntry) error
	LoadSnapshot(dir string) ([]persistance.SnapshotEntry, error)
}

func New(aofPath string, snapshotPath string) (*Store, error) {
	// Make sure that the file exists and open it
	aofFile, err := util.OpenOrCreate(aofPath)
	if err != nil {
		return nil, err
	}

	// Make sure that the directory exists and create it if it doesn't
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

	// Load the content of the snapshot file into memory
	if err = store.LoadSnapshot(); err != nil {
		return nil, err
	}

	// Load the content of the AOF file into memory
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
		_, exists := s.Get(key)
		if exists {
			// If the item already exists, don't override it
			return
		}
	}

	s.mu.Lock()
	s.items[key] = Item{
		Value:     value,
		ExpiresAt: expiresAt,
	}
	s.mu.Unlock()

	if s.aofPersistance != nil {

		// Retry if writing to the AOF file fails
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
	item, ok := s.items[key]
	s.mu.RUnlock()

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
	delete(s.items, key)
	s.mu.Unlock()

	if s.aofPersistance != nil {

		// Retry if writing to the AOF file fails
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
	entries, err := s.aofPersistance.LoadAOF(s.aofFile)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range entries {
		switch entry.Op {
		case "set":
			s.items[entry.Key] = Item{
				Value:     entry.Value,
				ExpiresAt: entry.ExpiresAt,
			}
		case "delete":
			delete(s.items, entry.Key)
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

	if err := s.snapshotPersistance.SaveSnapshot(s.snapshotDir, entries); err != nil {
		return err
	}

	// Clear content of the AOF file on successfull snapshot save
	if s.aofPersistance != nil {
		if err := s.aofPersistance.ClearAOF(s.aofFile); err != nil {
			return err
		}
	}

	return nil
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

func (s *Store) SaveSnapshotRegularly() {
	for {
		time.Sleep(30 * time.Second)
		if err := s.SaveSnapshot(); err != nil {
			fmt.Println("Error saving snapshot:", err)
		}
	}
}

func (s *Store) CleanExpiredItems() {
	for {
		time.Sleep(1 * time.Second)
		s.mu.Lock()
		for k, v := range s.items {
			if !v.ExpiresAt.IsZero() && v.ExpiresAt.Before(time.Now()) {
				delete(s.items, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Store) InitBackgroundTasks() {
	go s.SaveSnapshotRegularly()
	go s.CleanExpiredItems()
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.aofFile != nil {
		err := s.aofFile.Close()
		s.aofFile = nil
		return err
	}
	return nil
}
