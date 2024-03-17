package config

type MastodonConfig struct {
	Domain         string `yaml:"domain"`
	AccessToken    string `yaml:"access_token"`
	PostVisibility string `yaml:"post_visibility"`
}
