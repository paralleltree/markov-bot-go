package blog_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
)

type RewriteToHttpTransport struct {
	Transport http.RoundTripper
}

func (t *RewriteToHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

func newTestServer() (*http.Client, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	teardown := func() {
		server.Close()
	}

	transport := &RewriteToHttpTransport{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}
	client := &http.Client{Transport: transport}
	return client, mux, teardown
}
