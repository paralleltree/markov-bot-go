package handler

import (
	"fmt"
	"strings"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/markov"
	"github.com/paralleltree/markov-bot-go/persistence"
)

func GenerateAndPost(client *blog.MastodonClient, store persistence.PersistentStore, minWordsCount int, dryRun bool) error {
	model, err := loadModel(store)
	if err != nil {
		return fmt.Errorf("load model: %v", err)
	}

	for {
		generated := model.Generate()
		if len(generated) < minWordsCount {
			continue
		}
		text := strings.Join(generated, "")

		if dryRun {
			fmt.Println(text)
		} else {
			if _, err := client.CreateStatus(text, blog.StatusUnlisted); err != nil {
				return fmt.Errorf("create status: %v", err)
			}
		}
		break
	}
	return nil
}

func loadModel(store persistence.PersistentStore) (*markov.Chain, error) {
	data, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("load chain data: %v", err)
	}
	chain, err := markov.LoadChain(data)
	if err != nil {
		return nil, fmt.Errorf("reconstruct chain: %v", err)
	}
	return chain, nil
}
