package persistance

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/util"
)

type SnapshotEntry struct {
	Key       string
	Value     string
	ExpiresAt time.Time
}

type SnapshotPersistance struct{}

func NewSnapshotPersistance() *SnapshotPersistance {
	return &SnapshotPersistance{}
}

func (sp *SnapshotPersistance) SaveSnapshot(dir string, entries []SnapshotEntry) error {
    tempFilename := "snapshot.tmp"
    existingFilename := "snapshot.gob"
    tempPath := filepath.Join(dir, tempFilename)
    existingPath := filepath.Join(dir, existingFilename)
    
    file, err := util.OpenOrCreate(tempPath)
    if err != nil {
        return err
    }
    
    encoder := gob.NewEncoder(file)
    if err := encoder.Encode(entries); err != nil {
        return err
    }
    file.Close()

    // Replace the old snapshot file with the new one
    os.Remove(existingPath)
    os.Rename(tempPath, existingPath)

    return nil
}

func (sp *SnapshotPersistance) LoadSnapshot(dir string) ([]SnapshotEntry, error) {
    snapshotFilename := "snapshot.gob"
    path := filepath.Join(dir, snapshotFilename)

    file, err := os.Open(path)
    if err != nil {
        if os.IsNotExist(err) {
            return []SnapshotEntry{}, nil
        }
        return nil, err
    }
    defer file.Close()

    var entries []SnapshotEntry
    decoder := gob.NewDecoder(file)
    if err := decoder.Decode(&entries); err != nil {
        return nil, err
    }

    return entries, nil
}
