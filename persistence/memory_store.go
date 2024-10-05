package persistence

import (
	"context"
	"time"
)

type memoryStore struct {
	content []byte
	modTime time.Time
}

func NewMemoryStore() *memoryStore {
	return &memoryStore{
		content: []byte{},
		modTime: time.Now(),
	}
}

func (m *memoryStore) Load(ctx context.Context) ([]byte, error) {
	return m.content, nil
}

func (m *memoryStore) ModTime(ctx context.Context) (time.Time, bool, error) {
	return m.modTime, len(m.content) > 0, nil
}

func (m *memoryStore) Save(ctx context.Context, data []byte) error {
	m.content = data
	return nil
}
