package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/markov"
	"github.com/paralleltree/markov-bot-go/persistence"
)

const maxAttemptsCount = 100

var ErrGenerationFailed = fmt.Errorf("failed to generate a post")

func GenerateAndPost(ctx context.Context, client blog.BlogClient, store persistence.PersistentStore, minWordsCount int) error {
	model, err := loadModel(ctx, store)
	if err != nil {
		return fmt.Errorf("load model: %w", err)
	}

	for i := 0; i < maxAttemptsCount; i++ {
		generated := model.Generate()
		if len(generated) < minWordsCount {
			continue
		}
		text := strings.Join(generated, "")

		if err := client.CreatePost(ctx, text); err != nil {
			return fmt.Errorf("create status: %w", err)
		}
		return nil
	}
	return ErrGenerationFailed
}

func loadModel(ctx context.Context, store persistence.PersistentStore) (*markov.Chain, error) {
	data, err := store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load chain data: %w", err)
	}
	chain, err := markov.LoadChain(data)
	if err != nil {
		return nil, fmt.Errorf("reconstruct chain: %w", err)
	}
	return chain, nil
}
