package handler

import (
	"fmt"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/markov"
	"github.com/paralleltree/markov-bot-go/morpheme"
	"github.com/paralleltree/markov-bot-go/persistence"
)

func BuildChain(client *blog.MastodonClient, fetchStatusCount int, stateSize int, store persistence.PersistentStore) error {
	analyzer := morpheme.NewMecabAnalyzer("mecab-ipadic-neologd")

	uid, err := client.FetchUserId()
	if err != nil {
		return fmt.Errorf("fetch user id: %w", err)
	}

	chain := markov.NewChain(stateSize)
	iterator := client.FetchLatestPublicStatuses(uid, fetchStatusCount)

	for {
		statuses, hasNext, err := iterator()
		if err != nil {
			return fmt.Errorf("fetch statuses: %w", err)
		}
		if !hasNext {
			break
		}

		for _, s := range statuses {
			result, err := analyzer.Analyze(s)
			if err != nil {
				return fmt.Errorf("analyze text: %w", err)
			}
			for _, v := range result {
				chain.AddSource(v)
			}
		}
	}

	dump, err := chain.Dump()
	if err != nil {
		return fmt.Errorf("dump chain: %w", err)
	}
	if err := store.Save(dump); err != nil {
		return fmt.Errorf("save chain: %w", err)
	}

	return nil
}
