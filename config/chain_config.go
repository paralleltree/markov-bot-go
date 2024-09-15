package config

type ChainConfig struct {
	StateSize        int `yaml:"state_size"`
	FetchStatusCount int `yaml:"fetch_status_count"`
	MinWordsCount    int `yaml:"min_words_count"`
}

func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		StateSize:        3,
		FetchStatusCount: 200,
		MinWordsCount:    1,
	}
}
