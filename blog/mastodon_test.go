package blog_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/paralleltree/markov-bot-go/blog"
	"github.com/paralleltree/markov-bot-go/lib"
)

type mastodonStatus struct {
	Id         string `json:"id"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
}

func TestMastodonClient_CreateStatus(t *testing.T) {
	httpClient, mux, teardown := newTestServer()
	defer teardown()

	ctx := context.Background()
	wantHost := "foo.net"
	wantContentType := "application/x-www-form-urlencoded"
	wantAccessToken := "token"
	wantBody := "body"
	wantVisibility := "unlisted"
	wantId := "1"

	wantAuthorizationHeader := fmt.Sprintf("Bearer %s", wantAccessToken)

	mux.HandleFunc("/api/v1/statuses", func(w http.ResponseWriter, r *http.Request) {
		// check headers
		gotContentType := r.Header.Get("Content-Type")
		if wantContentType != gotContentType {
			t.Fatalf("unexpected Content-Type: expected %v, but got %v", wantContentType, gotContentType)
		}
		gotAuthorizationHeader := r.Header.Get("Authorization")
		if wantAuthorizationHeader != gotAuthorizationHeader {
			t.Fatalf("unexpected Authorization: expected %v, but got %v", wantAuthorizationHeader, gotAuthorizationHeader)
		}

		// check form values
		gotVisibility := r.FormValue("visibility")
		if wantVisibility != gotVisibility {
			t.Fatalf("unexpected visibility: want %s, but got %s", wantVisibility, gotVisibility)
		}

		gotBody := r.FormValue("status")
		if wantBody != gotBody {
			t.Fatalf("unexpected body: want %s, but got %s", wantBody, gotBody)
		}

		w.WriteHeader(http.StatusOK)
		res := mastodonStatus{
			Id:         wantId,
			Content:    wantBody,
			Visibility: wantVisibility,
		}
		body, err := json.Marshal(res)
		if err != nil {
			t.Fatalf("marshal response: %v", err)
		}
		w.Write(body)
	})

	client := blog.NewMastodonClientWithHttpClient(wantHost, wantAccessToken, wantVisibility, httpClient)
	err := client.CreatePost(ctx, wantBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMastodonClient_FetchLatestPublicStatuses_RequestsSpecifiedCount(t *testing.T) {
	httpClient, mux, teardown := newTestServer()
	defer teardown()

	ctx := context.Background()
	wantHost := "foo.jp"
	wantAccountId := "100"
	wantAccessToken := "token"
	oldestStatusId := "1"
	wantAuthorizationHeader := fmt.Sprintf("Bearer %s", wantAccessToken)

	directStatus := mastodonStatus{
		Id:         "4",
		Content:    "4",
		Visibility: "direct",
	}
	privateStatus := mastodonStatus{
		Id:         "3",
		Content:    "3",
		Visibility: "private",
	}
	unlistedStatus := mastodonStatus{
		Id:         "2",
		Content:    "2",
		Visibility: "unlisted",
	}
	publicStatus := mastodonStatus{
		Id:         oldestStatusId,
		Content:    "1",
		Visibility: "public",
	}

	responseStatuses := []mastodonStatus{
		directStatus,
		privateStatus,
		unlistedStatus,
		publicStatus,
	}
	wantStatuses := []mastodonStatus{
		unlistedStatus,
		publicStatus,
	}

	mux.HandleFunc(fmt.Sprintf("/api/v1/accounts/%s/statuses", wantAccountId), func(w http.ResponseWriter, r *http.Request) {
		maxId := r.URL.Query().Get("max_id")

		w.WriteHeader(http.StatusOK)

		if maxId == oldestStatusId {
			w.Write([]byte("[]"))
			return
		}

		body, err := json.Marshal(responseStatuses)
		if err != nil {
			t.Fatalf("marshal server response: %v", err)
		}
		w.Write(body)
	})

	inflateVerifyCredentialsHandler(t, mux, wantHost, wantAuthorizationHeader, wantAccountId)

	client := blog.NewMastodonClientWithHttpClient(wantHost, wantAccessToken, "", httpClient)
	iterator := client.GetPostsFetcher(ctx)
	gotStatuses := consumeIterator(t, iterator, len(wantStatuses))

	if len(wantStatuses) != len(gotStatuses) {
		t.Fatalf("unexpected result: expected %v items, but got %v items", len(wantStatuses), len(gotStatuses))
	}

	for i, wantStatus := range wantStatuses {
		gotStatus := gotStatuses[i]
		if wantStatus.Content != gotStatus {
			t.Fatalf("unexpected status: expected %v, but got %v", wantStatus.Content, gotStatus)
		}
	}
}

func TestMastodonClient_FetchLatestPublicStatuses_ReturnsPublicStatusesWithPaging(t *testing.T) {
	httpClient, mux, teardown := newTestServer()
	defer teardown()

	ctx := context.Background()
	wantHost := "foo.jp"
	wantAccountId := "100"
	wantAccessToken := "token"
	wantAuthorizationHeader := fmt.Sprintf("Bearer %s", wantAccessToken)

	statusesCount := 120
	allStatuses := []mastodonStatus{}
	for i := 0; i < statusesCount; i++ {
		id := fmt.Sprintf("%d", statusesCount-i)
		s := mastodonStatus{
			Id:         id,
			Content:    id,
			Visibility: "unlisted",
		}
		allStatuses = append(allStatuses, s)
	}

	// max_id to response map
	pages := map[string]struct {
		statuses  []mastodonStatus
		wantLimit int
	}{
		// first page
		"": {
			statuses:  allStatuses[0:100],
			wantLimit: 100,
		},
		// second(last) page
		allStatuses[99].Id: {
			statuses:  allStatuses[100:],
			wantLimit: 100,
		},
	}

	mux.HandleFunc(fmt.Sprintf("/api/v1/accounts/%s/statuses", wantAccountId), func(w http.ResponseWriter, r *http.Request) {
		gotMaxId := r.URL.Query().Get("max_id")

		gotLimitStr := r.URL.Query().Get("limit")
		gotLimit, err := strconv.ParseInt(gotLimitStr, 10, 64)
		if err != nil {
			t.Fatalf("parse limit: %v", err)
		}
		wantLimit := pages[gotMaxId].wantLimit
		if wantLimit != int(gotLimit) {
			t.Fatalf("unexpected limit: expected %v, but got %v", wantLimit, gotLimit)
		}

		w.WriteHeader(http.StatusOK)
		body, err := json.Marshal(pages[gotMaxId].statuses)
		if err != nil {
			t.Fatalf("marshal server response: %v", err)
		}
		w.Write(body)
	})

	inflateVerifyCredentialsHandler(t, mux, wantHost, wantAuthorizationHeader, wantAccountId)

	client := blog.NewMastodonClientWithHttpClient(wantHost, wantAccessToken, "", httpClient)
	iterator := client.GetPostsFetcher(ctx)
	gotStatuses := consumeIterator(t, iterator, statusesCount)

	if len(allStatuses) != len(gotStatuses) {
		t.Fatalf("unexpected result: expected %v items, but got %v items", len(allStatuses), len(gotStatuses))
	}

	for i, wantStatus := range allStatuses {
		gotStatus := gotStatuses[i]
		if wantStatus.Content != gotStatus {
			t.Fatalf("unexpected status: expected %v, but got %v", wantStatus.Content, gotStatus)
		}
	}
}

func TestMastodonClient_FetchLatestPublicStatuses_RequestsWithRequiredParameters(t *testing.T) {
	httpClient, mux, teardown := newTestServer()
	defer teardown()

	ctx := context.Background()
	wantHost := "foo.net"
	wantAccessToken := "token"
	wantAuthorizationHeader := fmt.Sprintf("Bearer %s", wantAccessToken)
	wantAccountId := "1"
	wantExcludeReblogs := "1"
	wantExcludeReplies := "1"

	mux.HandleFunc(fmt.Sprintf("/api/v1/accounts/%s/statuses", wantAccountId), func(w http.ResponseWriter, r *http.Request) {
		gotHost := r.URL.Host
		if wantHost != gotHost {
			t.Fatalf("unexpected host: expected %v, but got %v", wantHost, gotHost)
		}
		gotAuthorizationHeader := r.Header.Get("Authorization")
		if wantAuthorizationHeader != gotAuthorizationHeader {
			t.Fatalf("unexpected authorziation header: expected %v, but got %v", wantAuthorizationHeader, gotAuthorizationHeader)
		}

		gotExcludeReblogs := r.URL.Query().Get("exclude_reblogs")
		if wantExcludeReblogs != gotExcludeReblogs {
			t.Fatalf("unexpected exclude_reblogs: expected %v, but got %v", wantExcludeReblogs, gotExcludeReblogs)
		}
		gotExcludeReplies := r.URL.Query().Get("exclude_replies")
		if wantExcludeReplies != gotExcludeReplies {
			t.Fatalf("unexpected exclude_replies: expected %v, but got %v", wantExcludeReplies, gotExcludeReplies)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]")) // empty response
	})

	inflateVerifyCredentialsHandler(t, mux, wantHost, wantAuthorizationHeader, wantAccountId)

	client := blog.NewMastodonClientWithHttpClient(wantHost, wantAccessToken, "", httpClient)
	iter := client.GetPostsFetcher(ctx)
	consumeIterator(t, iter, 1)
}

func TestMastodonClient_FetchUserId_ReturnsUserId(t *testing.T) {
	httpClient, mux, teardown := newTestServer()
	defer teardown()

	ctx := context.Background()
	wantHost := "foo.net"
	wantAccessToken := "token"
	wantId := "123"
	wantAuthorizationHeader := fmt.Sprintf("Bearer %s", wantAccessToken)

	inflateVerifyCredentialsHandler(t, mux, wantHost, wantAuthorizationHeader, wantId)

	client := blog.NewMastodonClientWithHttpClient(wantHost, wantAccessToken, "", httpClient)
	gotId, err := client.FetchUserId(ctx)
	if err != nil {
		t.Fatalf("unexpected error while fetching user id: %v", err)
	}
	if wantId != gotId {
		t.Fatalf("unexpected id: expected %v, but got %v", wantId, gotId)
	}
}

func inflateVerifyCredentialsHandler(t *testing.T, mux *http.ServeMux, wantHost, wantAuthorizationHeader, wantId string) {
	mux.HandleFunc("/api/v1/accounts/verify_credentials", func(w http.ResponseWriter, r *http.Request) {
		gotHost := r.URL.Host
		if wantHost != gotHost {
			t.Fatalf("unexpected host: expected %v, but got %v", wantHost, gotHost)
		}

		gotAuthorizationHeader := r.Header.Get("Authorization")
		if wantAuthorizationHeader != gotAuthorizationHeader {
			t.Fatalf("unexpected authorziation header: expected %v, but got %v", wantAuthorizationHeader, gotAuthorizationHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"id": "%s", "username": "test"}`, wantId))) // empty response
	})
}

func consumeIterator[T any](t *testing.T, chunkIterator lib.ChunkIteratorFunc[T], count int) []T {
	t.Helper()
	iterator := lib.BuildIterator[T](chunkIterator)
	res := []T{}
	for {
		if count <= len(res) {
			break
		}
		item, ok, err := iterator()
		if err != nil {
			t.Fatalf("iterate result: %v", err)
		}
		if !ok {
			break
		}
		res = append(res, item)
	}
	return res
}
