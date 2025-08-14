package persistance

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/util"
)

type AOFEntry struct {
	Op        string
	Key       string
	Value     string
	ExpiresAt time.Time
}

type AOFPersistance struct{}

func NewAOFPersistance() *AOFPersistance {
	return &AOFPersistance{}
}

func (ap *AOFPersistance) AOFAppend(file *os.File, entry AOFEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(data) + "\n")
	return err
}

func (ap *AOFPersistance) LoadAOF(file *os.File) ([]AOFEntry, error) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	entries := make([]AOFEntry, 0)
	for scanner.Scan() {
		var entry AOFEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, err
		}

		if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(time.Now()) {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (ap *AOFPersistance) ClearAOF(file *os.File) error {
	filePath := file.Name()
	
	// Truncating the file didn't work so here's a dirty hack around it
	file.Close()
	if err := os.Remove(filePath); err != nil {
		return err
	}
	newFile, err := util.OpenOrCreate(filePath)
	if err != nil {
		return err
	}

	// Replace the old pointer to the file with the new one.
	// If I figure out how to truncate the file then that won't be needed.
	*file = *newFile

	return nil
}
