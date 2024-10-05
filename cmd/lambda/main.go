package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/paralleltree/markov-bot-go/config"
	"github.com/paralleltree/markov-bot-go/handler"
	"github.com/paralleltree/markov-bot-go/morpheme"
	"github.com/paralleltree/markov-bot-go/persistence"
)

func main() {
	lambda.Start(requestHandler)
}

type PostEvent struct {
	S3Region     string `json:"s3Region"`
	S3BucketName string `json:"s3BucketName"`
	S3KeyPrefix  string `json:"s3KeyPrefix"`
}

func requestHandler(ctx context.Context, e PostEvent) error {
	confStore, err := persistence.NewS3Store(e.S3Region, e.S3BucketName, fmt.Sprintf("%s/config.yml", e.S3KeyPrefix))
	if err != nil {
		return fmt.Errorf("new s3 store: %w", err)
	}

	conf, err := loadConfig(ctx, confStore)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	s3Store, err := persistence.NewS3Store(e.S3Region, e.S3BucketName, fmt.Sprintf("%s/model", e.S3KeyPrefix))
	if err != nil {
		return fmt.Errorf("new s3 store: %w", err)
	}
	modelStore := persistence.NewCompressedStore(s3Store)

	if err := run(ctx, conf, modelStore); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}

func run(ctx context.Context, conf *config.BotConfig, modelStore persistence.PersistentStore) error {
	analyzer := morpheme.NewMecabAnalyzer("mecab-ipadic-neologd")

	mod, ok, err := modelStore.ModTime(ctx)
	if err != nil {
		return fmt.Errorf("get modtime: %w", err)
	}

	buildChain := func() error {
		return handler.BuildChain(ctx, conf.FetchClient, analyzer, modelStore, handler.WithFetchStatusCount(conf.FetchStatusCount), handler.WithStateSize(conf.StateSize))
	}

	if !ok {
		// return an error if initial build fails
		if err := buildChain(); err != nil {
			return fmt.Errorf("build chain: %w", err)
		}
	}

	if float64(conf.ExpiresIn) < time.Since(mod).Seconds() {
		// attempt to build chain if expired
		// when building chain fails, it will use the existing chain
		if err := buildChain(); err != nil {
			fmt.Fprintf(os.Stderr, "build chain: %v\n", err)
		}
	}

	if err := handler.GenerateAndPost(ctx, conf.PostClient, modelStore, handler.WithMinWordsCount(conf.MinWordsCount)); err != nil {
		return fmt.Errorf("generate and post: %w", err)
	}

	return nil
}

func loadConfig(ctx context.Context, store persistence.PersistentStore) (*config.BotConfig, error) {
	data, err := store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	conf, err := config.LoadBotConfig(data)
	if err != nil {
		return nil, fmt.Errorf("load bot config: %w", err)
	}

	return conf, nil
}
