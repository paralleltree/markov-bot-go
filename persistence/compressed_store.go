package persistence

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"time"
)

type compressedStore struct {
	store PersistentStore
}

func NewCompressedStore(store PersistentStore) PersistentStore {
	return &compressedStore{
		store: store,
	}
}

func (s *compressedStore) Load() ([]byte, error) {
	raw, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	r, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("new gzip reader: %w", err)
	}
	defer r.Close()

	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read compressed stream: %w", err)
	}
	return body, nil
}

func (s *compressedStore) ModTime() (time.Time, bool, error) {
	return s.store.ModTime()
}

func (s *compressedStore) Save(data []byte) error {
	buf := new(bytes.Buffer)
	if err := compressStream(buf, data); err != nil {
		return fmt.Errorf("compress stream: %w", err)
	}

	if err := s.store.Save(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func compressStream(w io.Writer, data []byte) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()

	if _, err := gw.Write(data); err != nil {
		return fmt.Errorf("write compressed stream: %w", err)
	}

	return nil
}
