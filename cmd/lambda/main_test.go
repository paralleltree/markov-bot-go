package main

import (
	"testing"
	"time"

	"github.com/paralleltree/markov-bot-go/config"
	"github.com/paralleltree/markov-bot-go/lib"
)

func TestRun_WhenModelNotExists_CreatesModel(t *testing.T) {
	// arrange
	inputText := "アルミ缶の上にあるミカン"
	postClient := NewRecordableBlogClient(nil)
	conf := &config.BotConfig{
		FetchClient: NewRecordableBlogClient([]string{inputText}),
		PostClient:  postClient,
		ChainConfig: config.DefaultChainConfig(),
	}
	store := NewMemoryStore()

	// act
	if err := run(conf, store); err != nil {
		t.Errorf("run() should not return error, but got: %v", err)
	}

	// assert
	if len(postClient.PostedContents) != 1 {
		t.Errorf("PostedContents should have 1 item, but got: %v", postClient.PostedContents)
	}
	if inputText != postClient.PostedContents[0] {
		t.Errorf("unexpected output: want %s, but got %s", inputText, postClient.PostedContents[0])
	}
}

type recordableBlogClient struct {
	contents        []string
	contentsFetched bool
	PostedContents  []string
}

func NewRecordableBlogClient(contents []string) *recordableBlogClient {
	return &recordableBlogClient{
		contents: contents,
	}
}

func (f *recordableBlogClient) GetPostsFetcher() lib.ChunkIteratorFunc[string] {
	return func() ([]string, bool, error) {
		if f.contentsFetched {
			return nil, false, nil
		}
		f.contentsFetched = true
		return f.contents, false, nil
	}
}

func (f *recordableBlogClient) CreatePost(body string) error {
	f.PostedContents = append(f.PostedContents, body)
	return nil
}

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

func (m *memoryStore) Load() ([]byte, error) {
	return m.content, nil
}

func (m *memoryStore) ModTime() (time.Time, bool, error) {
	return m.modTime, len(m.content) > 0, nil
}

func (m *memoryStore) Save(data []byte) error {
	m.content = data
	return nil
}
