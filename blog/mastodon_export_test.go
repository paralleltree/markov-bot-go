package blog

import (
	"fmt"
	"net/http"
)

func NewMastodonClientWithHttpClient(domain, accessToken string, postVisibility string, client *http.Client) *MastodonClient {
	return &MastodonClient{
		Origin:         fmt.Sprintf("https://%s", domain),
		AccessToken:    accessToken,
		PostVisibility: postVisibility,
		client:         client,
	}
}
