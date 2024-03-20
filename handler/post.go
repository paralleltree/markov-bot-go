package handler

import (
	"fmt"
	"strings"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/markov"
	"github.com/paralleltree/markov-bot-go/persistence"
)

func GenerateAndPost(client blog.BlogClient, store persistence.PersistentStore, minWordsCount int) error {
	model, err := loadModel(store)
	if err != nil {
		return fmt.Errorf("load model: %w", err)
	}

	for {
		generated := model.Generate()
		if len(generated) < minWordsCount {
			continue
		}
		text := strings.Join(generated, "")

		if err := client.CreatePost(text); err != nil {
			return fmt.Errorf("create status: %w", err)
		}
		break
	}
	return nil
}

func loadModel(store persistence.PersistentStore) (*markov.Chain, error) {
	data, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("load chain data: %w", err)
	}
	chain, err := markov.LoadChain(data)
	if err != nil {
		return nil, fmt.Errorf("reconstruct chain: %w", err)
	}
	return chain, nil
}
