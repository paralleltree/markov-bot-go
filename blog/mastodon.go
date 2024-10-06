package blog

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/paralleltree/markov-bot-go/lib"
)

type MastodonStatusVisibility string

const (
	MastodonStatusPublic   = "public"
	MastodonStatusUnlisted = "unlisted"
	MastodonStatusPrivate  = "private"
	MastodonStatusDirect   = "direct"
)

type MastodonClient struct {
	Origin         string
	AccessToken    string
	PostVisibility string
	client         *http.Client
}

func NewMastodonClient(origin, accessToken string, postVisibility string) BlogClient {
	return &MastodonClient{
		Origin:         origin,
		AccessToken:    accessToken,
		PostVisibility: MastodonStatusUnlisted,
		client:         &http.Client{},
	}
}

func (c *MastodonClient) GetPostsFetcher(ctx context.Context) lib.ChunkIteratorFunc[string] {
	userId := ""
	maxId := ""
	return func() ([]string, bool, error) {
		if userId == "" {
			gotUserId, err := c.FetchUserId(ctx)
			if err != nil {
				return nil, false, fmt.Errorf("fetch user id: %w", err)
			}
			userId = gotUserId
		}

		chunkSize := 100
		statuses, hasNext, nextMaxId, err := c.fetchPublicStatusesChunk(ctx, userId, chunkSize, maxId)
		if err != nil {
			return nil, false, fmt.Errorf("fetch public statuses: %w", err)
		}
		maxId = nextMaxId
		return statuses, hasNext, nil
	}
}

// Returns status slice and minimum status id to fetch next older statuses.
// This function may returns statuses lesser than specified count because this exlcludes private and direct visibility statuses.
func (c *MastodonClient) fetchPublicStatusesChunk(ctx context.Context, userId string, count int, maxId string) ([]string, bool, string, error) {
	url := c.buildUrl(fmt.Sprintf("/api/v1/accounts/%s/statuses?limit=%d&exclude_reblogs=1&exclude_replies=1", userId, count))
	if maxId != "" {
		url = fmt.Sprintf("%s&max_id=%s", url, maxId)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	res, err := c.client.Do(req)
	if err != nil {
		return nil, false, "", fmt.Errorf("get statuses: %w", err)
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, false, "", fmt.Errorf("read response: %w", err)
	}

	statuses := []struct {
		Id         string `json:"id"`
		Content    string `json:"content"`
		Visibility string `json:"visibility"`
	}{}
	if err := json.Unmarshal(bytes, &statuses); err != nil {
		return nil, false, "", fmt.Errorf("unmarshal response: %w(%s)", err, bytes)
	}

	if len(statuses) == 0 {
		return nil, false, "", nil
	}

	tagPattern := regexp.MustCompile(`<[^>]*?>`)
	result := make([]string, 0, len(statuses))
	for _, v := range statuses {
		if v.Visibility == MastodonStatusPrivate || v.Visibility == MastodonStatusDirect {
			continue
		}
		// remove tags
		body := html.UnescapeString(tagPattern.ReplaceAllLiteralString(v.Content, ""))
		result = append(result, body)
	}
	return result, true, statuses[len(statuses)-1].Id, nil
}

func (c *MastodonClient) FetchUserId(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.buildUrl("/api/v1/accounts/verify_credentials"), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	res, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get account details: %w", err)
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}

	account := &struct {
		Id       string `json:"id"`
		UserName string `json:"username"`
	}{}
	if err := json.Unmarshal(bytes, account); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	return account.Id, nil
}

// Posts toot and returns created status id.
func (c *MastodonClient) CreatePost(ctx context.Context, payload string) error {
	form := url.Values{}
	form.Add("status", payload)
	form.Add("visibility", c.PostVisibility)
	body := strings.NewReader(form.Encode())

	req, err := http.NewRequestWithContext(ctx, "POST", c.buildUrl("/api/v1/statuses"), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("post status: %w", err)
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	status := &struct {
		Id string `json:"id"`
	}{}
	if err := json.Unmarshal(bytes, status); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	return nil
}

func (c *MastodonClient) buildUrl(path string) string {
	return fmt.Sprintf("%s%s", c.Origin, path)
}
