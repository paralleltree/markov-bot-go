package persistence

import (
	"context"
	"time"
)

type PersistentStore interface {
	// Returns data stream.
	Load(ctx context.Context) ([]byte, error)

	// Returns modified time and its existence of the stream.
	// If the second returned value is false, the data does not exists.
	ModTime(ctx context.Context) (time.Time, bool, error)

	// Saves given data stream.
	Save(ctx context.Context, data []byte) error
}
