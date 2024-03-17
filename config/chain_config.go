package config

type ChainConfig struct {
	StateSize        int `yaml:"state_size"`
	FetchStatusCount int `yaml:"fetch_status_count"`
	ExpiresIn        int `yaml:"expires_in"`
	MinWordsCount    int `yaml:"min_words_count"`
}

func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		StateSize:        3,
		FetchStatusCount: 200,
		ExpiresIn:        60 * 60 * 24,
		MinWordsCount:    1,
	}
}
