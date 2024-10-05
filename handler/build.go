package handler

import (
	"context"
	"fmt"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/lib"
	"github.com/paralleltree/markov-bot-go/markov"
	"github.com/paralleltree/markov-bot-go/morpheme"
	"github.com/paralleltree/markov-bot-go/persistence"
)

func BuildChain(ctx context.Context, client blog.BlogClient, analyzer morpheme.MorphemeAnalyzer, fetchStatusCount int, stateSize int, store persistence.PersistentStore) error {
	chain := markov.NewChain(stateSize)
	iterator := lib.BuildIterator(client.GetPostsFetcher(ctx))

	for i := 0; i < fetchStatusCount; i++ {
		status, hasNext, err := iterator()
		if err != nil {
			return fmt.Errorf("fetch statuses: %w", err)
		}
		if !hasNext {
			break
		}

		result, err := analyzer.Analyze(status)
		if err != nil {
			return fmt.Errorf("analyze text: %w", err)
		}
		for _, v := range result {
			chain.AddSource(v)
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
