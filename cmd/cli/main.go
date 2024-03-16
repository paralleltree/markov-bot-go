package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/handler"
	"github.com/paralleltree/markov-bot-go/persistence"
	"github.com/urfave/cli/v2"
)

const (
	SourceDomainKey      = "source-domain"
	SourceAccessTokenKey = "source-access-token"
	PostDomainKey        = "post-domain"
	PostAccessTokenKey   = "post-access-token"
	PostVisibility       = "post-visibility"
	StateSizeKey         = "state-size"
	FetchStatusCountKey  = "fetch-status-count"
	MinWordsCount        = "min-words-count"
	ExpiresInKey         = "expires-in"
	DryRunKey            = "dry-run"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	store := persistence.NewFileStore(".cache/model")

	buildingFlags := []cli.Flag{
		&cli.StringFlag{
			Name:     SourceDomainKey,
			Usage:    "mastodon domain of source account",
			Required: true,
			EnvVars:  []string{"SOURCE_DOMAIN"},
		},
		&cli.StringFlag{
			Name:     SourceAccessTokenKey,
			Usage:    "mastodon access token of source account",
			Required: true,
			EnvVars:  []string{"SOURCE_ACCESS_TOKEN"},
		},
		&cli.IntFlag{
			Name:    StateSizeKey,
			Usage:   "The state size of markov chain",
			Value:   3,
			EnvVars: []string{"STATE_SIZE"},
		},
		&cli.IntFlag{
			Name:    FetchStatusCountKey,
			Usage:   "The number of statuses to fetch from source account.",
			Value:   300,
			EnvVars: []string{"FETCH_STATUS_COUNT"},
		},
	}

	postingFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    PostDomainKey,
			Usage:   "mastodon domain of posting account",
			EnvVars: []string{"POST_DOMAIN"},
		},
		&cli.StringFlag{
			Name:    PostAccessTokenKey,
			Usage:   "mastodon access token of posting account",
			EnvVars: []string{"POST_ACCESS_TOKEN"},
		},
		&cli.StringFlag{
			Name:    PostVisibility,
			Usage:   "specifies the visibility of post.",
			EnvVars: []string{"POST_VISIBILITY"},
			Value:   "unlisted",
		},
		&cli.BoolFlag{
			Name:    DryRunKey,
			Usage:   "switches the output of generated text",
			EnvVars: []string{"DRY_RUN"},
		},
		&cli.IntFlag{
			Name:    MinWordsCount,
			Usage:   "specifies the minimum number of words",
			EnvVars: []string{"MIN_WORDS_COUNT"},
			Value:   1,
		},
		&cli.IntFlag{
			Name:    ExpiresInKey,
			Usage:   "specifies the duration to expire the model in seconds.",
			EnvVars: []string{"EXPIRES_IN"},
			Value:   60 * 60 * 24,
		},
	}

	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "Builds chain model and save it",
				Flags: buildingFlags,
				Action: func(c *cli.Context) error {
					client := blog.NewMastodonClient(c.String(SourceDomainKey), c.String(SourceAccessTokenKey), "")
					return handler.BuildChain(client, c.Int(FetchStatusCountKey), c.Int(StateSizeKey), store)
				},
			},
			{
				Name:  "post",
				Usage: "Posts new text from built chain",
				Flags: postingFlags,
				Action: func(c *cli.Context) error {
					client := blog.NewMastodonClient(c.String(PostDomainKey), c.String(PostAccessTokenKey), c.String(PostVisibility))
					return handler.GenerateAndPost(client, store, c.Int(MinWordsCount), c.Bool(DryRunKey))
				},
			},
			{
				Name:  "run",
				Usage: "Posts new text after building chain if it expired",
				Flags: append(append([]cli.Flag{}, buildingFlags...), postingFlags...),
				Action: func(c *cli.Context) error {
					srcClient := blog.NewMastodonClient(c.String(SourceDomainKey), c.String(SourceAccessTokenKey), "")
					postClient := blog.NewMastodonClient(c.String(PostDomainKey), c.String(PostAccessTokenKey), c.String(PostVisibility))
					mod, ok, err := store.ModTime()
					if err != nil {
						return fmt.Errorf("get modtime: %w", err)
					}
					if !ok || float64(c.Int(ExpiresInKey)) < time.Since(mod).Seconds() {
						if err := handler.BuildChain(srcClient, c.Int(FetchStatusCountKey), c.Int(StateSizeKey), store); err != nil {
							return err
						}
					}
					return handler.GenerateAndPost(postClient, store, c.Int(MinWordsCount), c.Bool(DryRunKey))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
