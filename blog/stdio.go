package blog

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/paralleltree/markov-bot-go/lib"
)

type stdIOClient struct {
}

func NewStdIOClient() BlogClient {
	return &stdIOClient{}
}

func (c *stdIOClient) GetPostsFetcher(ctx context.Context) lib.ChunkIteratorFunc[string] {
	stdin := bufio.NewScanner(os.Stdin)
	return func() ([]string, bool, error) {
		hasNext := stdin.Scan()
		if !hasNext {
			return nil, false, nil
		}

		if err := stdin.Err(); err != nil {
			return nil, false, err
		}

		return []string{stdin.Text()}, true, nil
	}
}

func (c *stdIOClient) CreatePost(ctx context.Context, body string) error {
	fmt.Println(body)
	return nil
}
