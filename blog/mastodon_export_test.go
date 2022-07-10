package blog

import "net/http"

func NewMastodonClientWithHttpClient(domain, accessToken string, client *http.Client) *MastodonClient {
	return &MastodonClient{
		Domain:      domain,
		AccessToken: accessToken,
		client:      client,
	}
}
