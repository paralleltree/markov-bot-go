package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/config"
	"github.com/paralleltree/markov-bot-go/lib"
	"github.com/paralleltree/markov-bot-go/persistence"
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
	store := persistence.NewMemoryStore()

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
	store := persistence.NewMemoryStore()

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
