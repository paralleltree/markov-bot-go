package blog

import (
	"context"

	"github.com/paralleltree/markov-bot-go/lib"
)

type BlogClient interface {
	GetPostsFetcher(ctx context.Context) lib.ChunkIteratorFunc[string]
	CreatePost(ctx context.Context, body string) error
}
