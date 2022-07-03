package persistence

import (
	"compress/gzip"
	"io/ioutil"
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

func (s *fileStore) Load() ([]byte, error) {
	f, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	stream, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (s *fileStore) ModTime() (time.Time, bool, error) {
	stat, err := os.Stat(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, false, nil
		}
		return time.Time{}, true, err
	}

	return stat.ModTime(), true, nil
}

func (s *fileStore) Save(data []byte) error {
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
