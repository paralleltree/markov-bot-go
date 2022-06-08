package lib

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"time"
)

type PersistentStore interface {
	// Returns data stream and its modified time.
	Load() ([]byte, time.Time, error)

	// Saves given data stream.
	Save(data []byte) error
}

type FileStore struct {
	path string
}

func NewFileStore(path string) PersistentStore {
	return &FileStore{
		path: path,
	}
}

func (s *FileStore) Load() ([]byte, time.Time, error) {
	f, err := os.Open(s.path)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer f.Close()

	stat, err := os.Stat(s.path)
	if err != nil {
		return nil, time.Time{}, err
	}

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, time.Time{}, err
	}

	stream, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, time.Time{}, err
	}
	return stream, stat.ModTime(), nil
}

func (s *FileStore) Save(data []byte) error {
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := gzip.NewWriter(f)
	defer w.Close()
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}
