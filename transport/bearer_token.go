package transport

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
)

type bearerTokenTransport struct {
	baseURL   *url.URL
	ctx       context.Context
	token     string
	userAgent string
	headers   map[string][]string
	base      http.RoundTripper
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.base
	if rt == nil {
		return nil, fmt.Errorf("No transport specified")
	}

	// Build the new request
	newReq := *req
	newReq.WithContext(t.ctx)
	newReq.Header = t.headers
	for k, v := range req.Header {
		newReq.Header[k] = v
	}
	newReq.Header["User-Agent"] = []string{t.userAgent}
	if t.token != "" {
		newReq.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", t.token)}
	}
	if newReq.Method == "PATCH" {
		newReq.Header["Content-Type"] = []string{fmt.Sprintf("application/vnd.layer-patch+json")}
	}

	return rt.RoundTrip(&newReq)
}
