package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/handler"
	"github.com/paralleltree/markov-bot-go/persistence"
	"gopkg.in/yaml.v3"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	lambda.Start(requestHandler)
}

type PostEvent struct {
	S3Region     string `json:"s3Region"`
	S3BucketName string `json:"s3BucketName"`
	S3KeyPrefix  string `json:"s3KeyPrefix"`
}

type Config struct {
	SourceDomain      string `yaml:"source_domain"`
	SourceAccessToken string `yaml:"source_access_token"`
	PostDomain        string `yaml:"post_domain"`
	PostAccessToken   string `yaml:"post_access_token"`
	StateSize         int    `yaml:"state_size"`
	FetchStatusCount  int    `yaml:"fetch_status_count"`
	ExpiresIn         int    `yaml:"expires_in"`
	MinWordsCount     int    `yaml:"min_words_count"`
}

func requestHandler(e PostEvent) error {
	confStore, err := persistence.NewS3Store(e.S3Region, e.S3BucketName, fmt.Sprintf("%s/config.yml", e.S3KeyPrefix))
	if err != nil {
		return err
	}
	conf, err := loadConfig(confStore)
	if err != nil {
		return err
	}

	modelStore, err := persistence.NewS3Store(e.S3Region, e.S3BucketName, fmt.Sprintf("%s/model", e.S3KeyPrefix))
	if err != nil {
		return err
	}

	postVisibility := blog.MastodonStatusUnlisted
	srcClient := blog.NewMastodonClient(conf.SourceDomain, conf.SourceAccessToken, "")
	postClient := blog.NewMastodonClient(conf.PostDomain, conf.PostAccessToken, postVisibility)

	mod, ok, err := modelStore.ModTime()
	if err != nil {
		return fmt.Errorf("get modtime: %w", err)
	}

	if !ok || float64(conf.ExpiresIn) < time.Since(mod).Seconds() {
		if err := handler.BuildChain(srcClient, conf.FetchStatusCount, conf.StateSize, modelStore); err != nil {
			return fmt.Errorf("build chain: %w", err)
		}
	}

	if err := handler.GenerateAndPost(postClient, modelStore, conf.MinWordsCount, false); err != nil {
		return fmt.Errorf("generate and post: %w", err)
	}

	return nil
}

func loadConfig(store persistence.PersistentStore) (Config, error) {
	res := Config{
		FetchStatusCount: 200,
		StateSize:        3,
		ExpiresIn:        60 * 60 * 24,
		MinWordsCount:    1,
	}

	data, err := store.Load()
	if err != nil {
		return res, fmt.Errorf("load config: %w", err)
	}

	if err := yaml.Unmarshal(data, &res); err != nil {
		return res, fmt.Errorf("unmarshal config: %w", err)
	}

	return res, nil
}
