package persistence

import (
	"bytes"
	"compress/gzip"
	"context"
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

func (s *compressedStore) Load(ctx context.Context) ([]byte, error) {
	raw, err := s.store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load data: %w", err)
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

func (s *compressedStore) ModTime(ctx context.Context) (time.Time, bool, error) {
	return s.store.ModTime(ctx)
}

func (s *compressedStore) Save(ctx context.Context, data []byte) error {
	buf := new(bytes.Buffer)
	if err := compressStream(buf, data); err != nil {
		return fmt.Errorf("compress stream: %w", err)
	}

	if err := s.store.Save(ctx, buf.Bytes()); err != nil {
		return fmt.Errorf("save compressed data: %w", err)
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
