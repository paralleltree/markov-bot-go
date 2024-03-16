package blog

import "github.com/paralleltree/markov-bot-go/lib"

type BlogClient interface {
	GetPostsFetcher(count int) lib.IteratorFunc[string]
	CreatePost(body string) error
}
