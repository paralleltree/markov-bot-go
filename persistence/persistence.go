package persistence

import (
	"time"
)

type PersistentStore interface {
	// Returns data stream.
	Load() ([]byte, error)

	// Returns modified time and its existence of the stream.
	// If the second returned value is false, the data does not exists.
	ModTime() (time.Time, bool, error)

	// Saves given data stream.
	Save(data []byte) error
}
