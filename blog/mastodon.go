package blog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	StatusPublic   = "public"
	StatusUnlisted = "unlisted"
	StatusPrivate  = "private"
	StatusDirect   = "direct"
)

type MastodonClient struct {
	Domain      string
	AccessToken string
}

func NewMastodonClient(domain, accessToken string) *MastodonClient {
	return &MastodonClient{
		Domain:      domain,
		AccessToken: accessToken,
	}
}

func (c *MastodonClient) FetchLatestPublicStatuses(userId string, count int) ([]string, error) {
	res := []string{}
	maxId := ""
	for count > 0 {
		chunkSize := 100
		if count < chunkSize {
			chunkSize = count
		}
		statuses, nextMaxId, err := c.fetchPublicStatusesChunk(userId, chunkSize, maxId)
		if err != nil {
			return nil, err
		}
		res = append(res, statuses...)
		maxId = nextMaxId
		count -= len(statuses)
	}
	return res, nil
}

// Returns status slice and minimum status id to fetch next older statuses.
// This function may returns statuses lesser than specified count because this exlcludes private and direct visibility statuses.
func (c *MastodonClient) fetchPublicStatusesChunk(userId string, count int, maxId string) ([]string, string, error) {
	url := c.buildUrl(fmt.Sprintf("/api/v1/accounts/%s/statuses?limit=%d&exclude_reblogs=1&exclude_replies=1", userId, count))
	if maxId != "" {
		url = fmt.Sprintf("%s&max_id=%s", url, maxId)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("get statuses: %v", err)
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}

	statuses := []struct {
		Id         string `json:"id"`
		Content    string `json:"content"`
		Visibility string `json:"visibility"`
	}{}
	if err := json.Unmarshal(bytes, &statuses); err != nil {
		return nil, "", fmt.Errorf("unmarshal response: %v(%s)", err, bytes)
	}

	tagPattern := regexp.MustCompile(`<[^>]*?>`)
	result := make([]string, 0, len(statuses))
	for _, v := range statuses {
		if v.Visibility == StatusPrivate || v.Visibility == StatusDirect {
			continue
		}
		// remove tags
		result = append(result, tagPattern.ReplaceAllLiteralString(v.Content, ""))
	}
	return result, statuses[len(statuses)-1].Id, nil
}

func (c *MastodonClient) FetchUserId() (string, error) {
	req, err := http.NewRequest("GET", c.buildUrl("/api/v1/accounts/verify_credentials"), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get account details: %v", err)
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
		return "", fmt.Errorf("unmarshal response: %v", err)
	}
	return account.Id, nil
}

// Posts toot and returns created status id.
func (c *MastodonClient) CreateStatus(payload string, visibility string) (string, error) {
	form := url.Values{}
	form.Add("status", payload)
	form.Add("visibility", visibility)
	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", c.buildUrl("/api/v1/statuses"), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("post status: %v", err)
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	status := &struct {
		Id string `json:"id"`
	}{}
	if err := json.Unmarshal(bytes, status); err != nil {
		return "", fmt.Errorf("unmarshal response: %v", err)
	}
	return status.Id, nil
}

func (c *MastodonClient) buildUrl(path string) string {
	return fmt.Sprintf("https://%s%s", c.Domain, path)
}
