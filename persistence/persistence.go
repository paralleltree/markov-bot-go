package persistence

import (
	"time"
)

type PersistentStore interface {
	// Returns data stream.
	Load() ([]byte, error)

	// Returns modified time of the stream.
	ModTime() (time.Time, error)

	// Saves given data stream.
	Save(data []byte) error
}
