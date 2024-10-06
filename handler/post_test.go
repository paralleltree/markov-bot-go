package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/handler"
	"github.com/paralleltree/markov-bot-go/morpheme"
	"github.com/paralleltree/markov-bot-go/persistence"
)

var analyzer = morpheme.NewMecabAnalyzer("mecab-ipadic-neologd")

func TestGenerateAndPost_WhenModelIsEmpty_ReturnsGenerateFailedError(t *testing.T) {
	// arrange
	ctx := context.Background()
	postClient := blog.NewRecordableBlogClient(nil)
	fetchClient := blog.NewRecordableBlogClient([]string{""})
	store := persistence.NewMemoryStore()

	if err := handler.BuildChain(ctx, fetchClient, analyzer, store); err != nil {
		t.Errorf("BuildChain() should not return error, but got: %v", err)
	}

	// act
	err := handler.GenerateAndPost(ctx, postClient, store)

	// assert
	if err == nil {
		t.Errorf("run() should return error, but got nil")
	}
	if !errors.Is(err, handler.ErrGenerationFailed) {
		t.Errorf("run() should return ErrGenerateFailed, but got: %v", err)
	}
}

func TestGenerateAndPost_ReturnsTextFromModel(t *testing.T) {
	cases := []struct {
		inputText string
	}{
		{
			inputText: "アルミ缶の上にあるミカン",
		},
		{
			inputText: "GetPostsFetcherの呼び出し",
		},
		{
			// English sentence
			inputText: "The quick brown fox jumps over the lazy dog.",
		},
		{
			inputText: "12:00:50に出力されたログ",
		},
		{
			inputText: "EZ DO DANCE",
		},
	}

	for _, tt := range cases {
		t.Run(tt.inputText, func(t *testing.T) {
			// arrange
			ctx := context.Background()
			postClient := blog.NewRecordableBlogClient(nil)
			fetchClient := blog.NewRecordableBlogClient([]string{tt.inputText})
			store := persistence.NewMemoryStore()

			if err := handler.BuildChain(ctx, fetchClient, analyzer, store); err != nil {
				t.Errorf("BuildChain() should not return error, but got: %v", err)
			}

			// act
			err := handler.GenerateAndPost(ctx, postClient, store)

			// assert
			if err != nil {
				t.Errorf("run() should not return error, but got: %v", err)
			}
			if len(postClient.PostedContents) != 1 {
				t.Errorf("unexpected items count: want %d, but got %d", 1, len(postClient.PostedContents))
			}
			if tt.inputText != postClient.PostedContents[0] {
				t.Errorf("unexpected output: want %s, but got %s", tt.inputText, postClient.PostedContents[0])
			}
		})
	}
}
