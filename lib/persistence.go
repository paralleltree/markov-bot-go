package lib

import (
	"compress/gzip"
	"io/ioutil"
	"os"
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

type FileStore struct {
	path string
}

func NewFileStore(path string) PersistentStore {
	return &FileStore{
		path: path,
	}
}

func (s *FileStore) Load() ([]byte, error) {
	f, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	stream, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (s *FileStore) ModTime() (time.Time, error) {
	stat, err := os.Stat(s.path)
	if err != nil {
		return time.Time{}, err
	}

	return stat.ModTime(), nil
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
