package config

import (
	"fmt"
	"strings"

	"github.com/paralleltree/markov-bot-go/blog"
	"gopkg.in/yaml.v3"
)

type BotConfig struct {
	FetchClient blog.BlogClient
	PostClient  blog.BlogClient
	ChainConfig
}

type ConfigFile struct {
	Input       map[string]interface{} `yaml:"input"`
	Output      map[string]interface{} `yaml:"output"`
	ChainConfig `yaml:",inline"`
}

func LoadBotConfig(body []byte) (*BotConfig, error) {
	conf := ConfigFile{
		ChainConfig: DefaultChainConfig(),
	}
	if err := yaml.Unmarshal(body, &conf); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	fetchClient, err := resolveBlogClient(conf.Input)
	if err != nil {
		return nil, fmt.Errorf("resolve fetch client: %w", err)
	}
	postClient, err := resolveBlogClient(conf.Output)
	if err != nil {
		return nil, fmt.Errorf("resolve post client: %w", err)
	}

	return &BotConfig{
		FetchClient: fetchClient,
		PostClient:  postClient,
		ChainConfig: conf.ChainConfig,
	}, nil
}

func resolveBlogClient(conf map[string]interface{}) (blog.BlogClient, error) {
	platform, ok := conf["platform"].(string)
	if !ok {
		return nil, fmt.Errorf("platform is not specified")
	}

	switch strings.ToLower(platform) {
	case "mastodon":
		origin := resolveMapValue[string](conf, "origin")
		accessToken := resolveMapValue[string](conf, "access_token")
		PostVisibility := resolveMapValue[string](conf, "post_visibility")
		return blog.NewMastodonClient(origin, accessToken, PostVisibility), nil

	case "stdio":
		return blog.NewStdIOClient(), nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// Finds a value from a map and returns it as a specified type.
// If the value is not found or the type is not matched, it returns the zero value of the specified type.
func resolveMapValue[T any](conf map[string]interface{}, key string) T {
	var empty T
	raw, ok := conf[key]
	if !ok {
		return empty
	}
	value, ok := raw.(T)
	if !ok {
		return empty
	}
	return value
}
