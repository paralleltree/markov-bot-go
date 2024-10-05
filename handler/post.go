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

type generatePostConf struct {
	minWordsCount int
}

func WithMinWordsCount(minWordsCount int) func(c *generatePostConf) {
	return func(c *generatePostConf) {
		c.minWordsCount = minWordsCount
	}
}

func GenerateAndPost(ctx context.Context, client blog.BlogClient, store persistence.PersistentStore, optFns ...func(*generatePostConf)) error {
	conf := &generatePostConf{
		minWordsCount: 1,
	}
	for _, f := range optFns {
		f(conf)
	}

	model, err := loadModel(ctx, store)
	if err != nil {
		return fmt.Errorf("load model: %w", err)
	}

	for i := 0; i < maxAttemptsCount; i++ {
		generated := model.Generate()
		if len(generated) < conf.minWordsCount {
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
