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
