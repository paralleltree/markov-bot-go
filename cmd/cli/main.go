package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/paralleltree/markov-bot-go/config"
	"github.com/paralleltree/markov-bot-go/handler"
	"github.com/paralleltree/markov-bot-go/morpheme"
	"github.com/paralleltree/markov-bot-go/persistence"
	"github.com/urfave/cli/v2"
)

const (
	ConfigFileKey       = "config-file"
	ModelFileKey        = "model-file"
	StateSizeKey        = "state-size"
	FetchStatusCountKey = "fetch-status-count"
	MinWordsCount       = "min-words-count"
	ExpiresInKey        = "expires-in"
	DryRunKey           = "dry-run"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	configFileFlag := &cli.StringFlag{
		Name:  ConfigFileKey,
		Usage: "Load configuration from `FILE`. If command-line arguments or environment variables are set, they override the configuration file.",
	}
	modelFileFlag := &cli.StringFlag{
		Name:     ModelFileKey,
		Usage:    "Load model from `FILE`.",
		Required: true,
	}

	buildingFlags := []cli.Flag{
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
		configFileFlag,
		modelFileFlag,
	}

	postingFlags := []cli.Flag{
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
		configFileFlag,
		modelFileFlag,
	}

	analyzer := morpheme.NewMecabAnalyzer("mecab-ipadic-neologd")

	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "Builds chain model and save it",
				Flags: buildingFlags,
				Action: func(c *cli.Context) error {
					store := persistence.NewFileStore(c.String(ModelFileKey))
					conf, err := LoadBotConfigFromFile(c.String(ConfigFileKey))
					if err != nil {
						return fmt.Errorf("load config: %w", err)
					}
					overrideChainConfigFromCli(&conf.ChainConfig, c)
					return handler.BuildChain(conf.FetchClient, analyzer, conf.FetchStatusCount, conf.StateSize, store)
				},
			},
			{
				Name:  "post",
				Usage: "Posts new text from built chain",
				Flags: postingFlags,
				Action: func(c *cli.Context) error {
					store := persistence.NewFileStore(c.String(ModelFileKey))
					conf, err := LoadBotConfigFromFile(c.String(ConfigFileKey))
					if err != nil {
						return fmt.Errorf("load config: %w", err)
					}
					overrideChainConfigFromCli(&conf.ChainConfig, c)
					return handler.GenerateAndPost(conf.PostClient, store, conf.MinWordsCount, c.Bool(DryRunKey))
				},
			},
			{
				Name:  "run",
				Usage: "Posts new text after building chain if it expired",
				Flags: append(append([]cli.Flag{}, buildingFlags...), postingFlags...),
				Action: func(c *cli.Context) error {
					store := persistence.NewFileStore(c.String(ModelFileKey))
					conf, err := LoadBotConfigFromFile(c.String(ConfigFileKey))
					if err != nil {
						return fmt.Errorf("load config: %w", err)
					}
					overrideChainConfigFromCli(&conf.ChainConfig, c)
					mod, ok, err := store.ModTime()
					if err != nil {
						return fmt.Errorf("get modtime: %w", err)
					}
					if !ok || float64(conf.ExpiresIn) < time.Since(mod).Seconds() {
						if err := handler.BuildChain(conf.FetchClient, analyzer, conf.FetchStatusCount, conf.StateSize, store); err != nil {
							return err
						}
					}
					return handler.GenerateAndPost(conf.PostClient, store, conf.MinWordsCount, c.Bool(DryRunKey))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func overrideChainConfigFromCli(conf *config.ChainConfig, c *cli.Context) {
	if c.IsSet(StateSizeKey) {
		conf.StateSize = c.Int(StateSizeKey)
	}
	if c.IsSet(FetchStatusCountKey) {
		conf.FetchStatusCount = c.Int(FetchStatusCountKey)
	}
	if c.IsSet(ExpiresInKey) {
		conf.ExpiresIn = c.Int(ExpiresInKey)
	}
	if c.IsSet(MinWordsCount) {
		conf.MinWordsCount = c.Int(MinWordsCount)
	}
}

func LoadBotConfigFromFile(path string) (*config.BotConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	confBody, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	return config.LoadBotConfig(confBody)
}
