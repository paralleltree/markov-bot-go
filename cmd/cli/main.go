package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/paralleltree/markov-bot-go/blog"
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
	MinWordsCountKey    = "min-words-count"
	DryRunKey           = "dry-run"
)

func main() {
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
	}

	postingFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    DryRunKey,
			Usage:   "switches the output of generated text",
			EnvVars: []string{"DRY_RUN"},
		},
		&cli.IntFlag{
			Name:    MinWordsCountKey,
			Usage:   "specifies the minimum number of words",
			EnvVars: []string{"MIN_WORDS_COUNT"},
			Value:   1,
		},
	}

	commonFlags := []cli.Flag{
		configFileFlag,
		modelFileFlag,
	}

	analyzer := morpheme.NewMecabAnalyzer("mecab-ipadic-neologd")

	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "Builds chain model and save it",
				Flags: append(append([]cli.Flag{}, commonFlags...), buildingFlags...),
				Action: func(c *cli.Context) error {
					store := persistence.NewCompressedStore(persistence.NewFileStore(c.String(ModelFileKey)))
					conf, err := LoadBotConfigFromFile(c.String(ConfigFileKey))
					if err != nil {
						return fmt.Errorf("load config: %w", err)
					}
					overrideChainConfigFromCli(&conf.ChainConfig, c)
					return handler.BuildChain(c.Context, conf.FetchClient, analyzer, store, handler.WithFetchStatusCount(conf.FetchStatusCount), handler.WithStateSize(conf.StateSize))
				},
			},
			{
				Name:  "post",
				Usage: "Posts new text from built chain",
				Flags: append(append([]cli.Flag{}, commonFlags...), postingFlags...),
				Action: func(c *cli.Context) error {
					store := persistence.NewCompressedStore(persistence.NewFileStore(c.String(ModelFileKey)))
					conf, err := LoadBotConfigFromFile(c.String(ConfigFileKey))
					if err != nil {
						return fmt.Errorf("load config: %w", err)
					}
					overrideChainConfigFromCli(&conf.ChainConfig, c)
					if c.Bool(DryRunKey) {
						conf.PostClient = blog.NewStdIOClient()
					}
					return handler.GenerateAndPost(c.Context, conf.PostClient, store, handler.WithMinWordsCount(conf.MinWordsCount))
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
	if c.IsSet(MinWordsCountKey) {
		conf.MinWordsCount = c.Int(MinWordsCountKey)
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
