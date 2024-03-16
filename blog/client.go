package blog

import "github.com/paralleltree/markov-bot-go/lib"

type BlogClient interface {
	GetPostsFetcher() lib.ChunkIteratorFunc[string]
	CreatePost(body string) error
}
