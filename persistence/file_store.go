package persistence

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

type fileStore struct {
	path string
}

func NewFileStore(path string) PersistentStore {
	return &fileStore{
		path: path,
	}
}

func (s *fileStore) Load(ctx context.Context) ([]byte, error) {
	f, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	stream, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return stream, nil
}

func (s *fileStore) ModTime(ctx context.Context) (time.Time, bool, error) {
	stat, err := os.Stat(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, false, nil
		}
		return time.Time{}, true, fmt.Errorf("stat file: %w", err)
	}

	return stat.ModTime(), true, nil
}

func (s *fileStore) Save(ctx context.Context, data []byte) error {
	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}
	return nil
}
