package blog

import (
	"context"

	"github.com/paralleltree/markov-bot-go/lib"
)

type recordableBlogClient struct {
	contents        []string
	contentsFetched bool
	PostedContents  []string
}

func NewRecordableBlogClient(contents []string) *recordableBlogClient {
	return &recordableBlogClient{
		contents: contents,
	}
}

func (f *recordableBlogClient) GetPostsFetcher(ctx context.Context) lib.ChunkIteratorFunc[string] {
	return func() ([]string, bool, error) {
		if f.contentsFetched {
			return nil, false, nil
		}
		f.contentsFetched = true
		return f.contents, false, nil
	}
}

func (f *recordableBlogClient) CreatePost(ctx context.Context, body string) error {
	f.PostedContents = append(f.PostedContents, body)
	return nil
}
