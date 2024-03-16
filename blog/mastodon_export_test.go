package blog

import "net/http"

func NewMastodonClientWithHttpClient(domain, accessToken string, postVisibility string, client *http.Client) *MastodonClient {
	return &MastodonClient{
		Domain:         domain,
		AccessToken:    accessToken,
		PostVisibility: postVisibility,
		client:         client,
	}
}
