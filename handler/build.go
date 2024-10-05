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

type buildChainConf struct {
	fetchStatusCount int
	stateSize        int
}

func WithFetchStatusCount(fetchStatusCount int) func(c *buildChainConf) {
	return func(c *buildChainConf) {
		c.fetchStatusCount = fetchStatusCount
	}
}

func WithStateSize(stateSize int) func(c *buildChainConf) {
	return func(c *buildChainConf) {
		c.stateSize = stateSize
	}
}

func BuildChain(ctx context.Context, client blog.BlogClient, analyzer morpheme.MorphemeAnalyzer, store persistence.PersistentStore, optFns ...func(*buildChainConf)) error {
	conf := &buildChainConf{
		fetchStatusCount: 100,
		stateSize:        2,
	}
	for _, f := range optFns {
		f(conf)
	}

	chain := markov.NewChain(conf.stateSize)
	iterator := lib.BuildIterator(client.GetPostsFetcher(ctx))

	for i := 0; i < conf.fetchStatusCount; i++ {
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
	if err := store.Save(ctx, dump); err != nil {
		return fmt.Errorf("save chain: %w", err)
	}

	return nil
}
