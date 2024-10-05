package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/config"
	"github.com/paralleltree/markov-bot-go/handler"
	"github.com/paralleltree/markov-bot-go/lib"
)

func TestRun_WhenModelNotExists_CreatesModel(t *testing.T) {
	// arrange
	ctx := context.Background()
	inputText := "アルミ缶の上にあるミカン"
	postClient := blog.NewRecordableBlogClient(nil)
	conf := &config.BotConfig{
		FetchClient: blog.NewRecordableBlogClient([]string{inputText}),
		PostClient:  postClient,
		ChainConfig: config.DefaultChainConfig(),
	}
	store := NewMemoryStore()

	// act
	if err := run(ctx, conf, store); err != nil {
		t.Errorf("run() should not return error, but got: %v", err)
	}

	// assert
	wantResult := []string{inputText}
	if !reflect.DeepEqual(wantResult, postClient.PostedContents) {
		t.Errorf("unexpected output: want %s, but got %s", inputText, postClient.PostedContents[0])
	}
}

func TestRun_WhenModelIsEmpty_ReturnsGenerateFailedError(t *testing.T) {
	// arrange
	ctx := context.Background()
	postClient := blog.NewRecordableBlogClient(nil)
	conf := &config.BotConfig{
		FetchClient: blog.NewRecordableBlogClient(nil),
		PostClient:  postClient,
		ChainConfig: config.DefaultChainConfig(),
	}
	store := NewMemoryStore()

	// act
	err := run(ctx, conf, store)

	// assert
	if err == nil {
		t.Errorf("run() should return error, but got nil")
	}
	if !errors.Is(err, handler.ErrGenerationFailed) {
		t.Errorf("run() should return ErrGenerateFailed, but got: %v", err)
	}
}

func TestRun_WhenModelAlreadyExistsAndBuildingModelFails_PostsWithExistingModelAndReturnsNoError(t *testing.T) {
	// arrange
	ctx := context.Background()
	inputText := "アルミ缶の上にあるミカン"
	postClient := blog.NewRecordableBlogClient(nil)
	conf := &config.BotConfig{
		FetchClient: blog.NewRecordableBlogClient([]string{inputText}),
		PostClient:  blog.NewRecordableBlogClient(nil), // discard posted content
		ChainConfig: config.DefaultChainConfig(),
	}
	store := NewMemoryStore()

	// build model
	if err := run(ctx, conf, store); err != nil {
		t.Errorf("run() should not return error, but got: %v", err)
	}

	conf = &config.BotConfig{
		FetchClient: &errorBlogClient{},
		PostClient:  postClient,
		ChainConfig: config.ChainConfig{
			FetchStatusCount: 1,
			ExpiresIn:        0, // force building chain
		},
	}

	// act
	if err := run(ctx, conf, store); err != nil {
		t.Errorf("run() should not return error, but got: %v", err)
	}

	// assert
	wantResult := []string{inputText}
	if !reflect.DeepEqual(wantResult, postClient.PostedContents) {
		t.Errorf("unexpected output: want %s, but got %s", inputText, postClient.PostedContents[0])
	}
}

type errorBlogClient struct{}

func (e *errorBlogClient) GetPostsFetcher(ctx context.Context) lib.ChunkIteratorFunc[string] {
	return func() ([]string, bool, error) {
		return nil, false, fmt.Errorf("failed to fetch posts")
	}
}

func (e *errorBlogClient) CreatePost(ctx context.Context, body string) error {
	return fmt.Errorf("failed to create post")
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
