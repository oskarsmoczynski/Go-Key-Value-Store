package persistance

import (
	"encoding/json"
	"os"
    "time"
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
