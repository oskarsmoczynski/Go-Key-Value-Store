package persistance

import (
	"encoding/json"
	"os"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/store"
)

type AOFPersistance struct{}

func NewAOFPersistance() *AOFPersistance {
    // TODO: add file creation
	return &AOFPersistance{}
}

func (a *AOFPersistance) AOFAppend(file *os.File, entry store.AOFEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(data) + "\n")
	return err
}
