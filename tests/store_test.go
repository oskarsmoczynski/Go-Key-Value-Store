package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/persistance"
	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/store"
)

// fake implementations for dependencies
type fakeAOF struct {
	entries   []persistance.AOFEntry
	appendErr error
	cleared   bool
}

func (f *fakeAOF) AOFAppend(_ *os.File, entry persistance.AOFEntry) error {
	if f.appendErr != nil {
		return f.appendErr
	}
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeAOF) LoadAOF(_ *os.File) ([]persistance.AOFEntry, error) {
	now := time.Now()
	result := make([]persistance.AOFEntry, 0, len(f.entries))
	for _, e := range f.entries {
		if !e.ExpiresAt.IsZero() && e.ExpiresAt.Before(now) {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

func (f *fakeAOF) ClearAOF(_ *os.File) error {
	f.cleared = true
	f.entries = nil
	return nil
}

type fakeSnapshot struct {
	saved   []persistance.SnapshotEntry
	toLoad  []persistance.SnapshotEntry
	saveErr error
	loadErr error
}

func (f *fakeSnapshot) SaveSnapshot(_ string, entries []persistance.SnapshotEntry) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	// copy to avoid aliasing
	f.saved = append([]persistance.SnapshotEntry(nil), entries...)
	return nil
}

func (f *fakeSnapshot) LoadSnapshot(_ string) ([]persistance.SnapshotEntry, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return append([]persistance.SnapshotEntry(nil), f.toLoad...), nil
}

func newInMemoryStore() *store.Store {
	dir := os.TempDir()
	s, _ := store.New(filepath.Join(dir, "test-aof.log"), filepath.Join(dir, "test-snapshots"))
	return s
}

func TestSetAndGet(t *testing.T) {
	s := newInMemoryStore()

	s.Set("foo", "bar", 0, true)

	v, ok := s.Get("foo")
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if v != "bar" {
		t.Fatalf("expected value 'bar', got %q", v)
	}
}

func TestSetOverrideBehavior(t *testing.T) {
	s := newInMemoryStore()

	s.Set("k", "v1", 0, true)
	s.Set("k", "v2", 0, false)
	if v, _ := s.Get("k"); v != "v1" {
		t.Fatalf("expected value to remain 'v1', got %q", v)
	}
	s.Set("k", "v3", 0, true)
	if v, _ := s.Get("k"); v != "v3" {
		t.Fatalf("expected value to be 'v3', got %q", v)
	}
}

func TestDelete(t *testing.T) {
	s := newInMemoryStore()
	s.Set("a", "b", 0, true)
	s.Delete("a")
	if _, ok := s.Get("a"); ok {
		t.Fatalf("expected key to be deleted")
	}
}

func TestTTLExpiryViaGet(t *testing.T) {
	s := newInMemoryStore()
	s.Set("ttl", "value", 1, true)
	time.Sleep(1100 * time.Millisecond)
	if _, ok := s.Get("ttl"); ok {
		t.Fatalf("expected key to have expired")
	}
}
