package blog

import "github.com/paralleltree/markov-bot-go/lib"

type BlogClient interface {
	GetPostsFetcher(count int) lib.ChunkIteratorFunc[string]
	CreatePost(body string) error
}
