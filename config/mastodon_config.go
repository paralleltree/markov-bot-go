package config

type MastodonConfig struct {
	Origin         string `yaml:"origin"`
	AccessToken    string `yaml:"access_token"`
	PostVisibility string `yaml:"post_visibility"`
}
