package persistance

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"
    "strconv"
    "strings"

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
    timestamp := time.Now().Unix()
    filename := makeFilename(timestamp)
    path := filepath.Join(dir, filename)

    file, err := util.OpenOrCreate(path)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := gob.NewEncoder(file)
    if err := encoder.Encode(entries); err != nil {
        return err
    }
    return nil
}

func (sp *SnapshotPersistance) LoadSnapshot(dir string) ([]SnapshotEntry, error) {
    files, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    maxTimestamp := int64(0)
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        parts := strings.Split(file.Name(), ".")
        timestamp, err := strconv.ParseInt(parts[0], 10, 64)
        if err != nil {
            continue
        }
        if timestamp > maxTimestamp {
            maxTimestamp = timestamp
        }
    }
    if maxTimestamp == 0 {
        return nil, nil
    }

    filename := makeFilename(maxTimestamp)
    path := filepath.Join(dir, filename)

    file, err := os.Open(path)
    if err != nil {
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

func makeFilename(timestamp int64) string {
    return fmt.Sprintf("%d.snapshot", timestamp)
}