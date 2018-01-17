package server

import (
	"context"
	"fmt"
	"net/url"

	"github.com/layerhq/go-client/option"
	"github.com/layerhq/go-client/transport"
)

type Server struct {
	baseURL   *url.URL
	appID     string
	transport *transport.HTTPTransport
}

type updateOperation struct {
	Operation string      `json:"operation"`
	Property  string      `json:"property"`
	Value     interface{} `json:"value"`
}

// NewClient creates a new Layer Server API client
func NewClient(ctx context.Context, appID string, options ...option.ClientOption) (*Server, error) {
	u, err := url.Parse(fmt.Sprintf("https://api.layer.com/apps/%s/", appID))
	if err != nil {
		return nil, fmt.Errorf("Error building base URL: %v", err)
	}

	return NewTestClient(ctx, u, appID, options...)
}

// NewTestClient
func NewTestClient(ctx context.Context, u *url.URL, appID string, options ...option.ClientOption) (*Server, error) {
	headers := map[string][]string{
		"Accept":       {"application/vnd.layer+json; version=3.0"},
		"Content-Type": {"application/json"},
	}

	options = append(options, option.WithHeaders(headers))

	t, err := transport.NewHTTPTransport(ctx, appID, u, nil, options...)
	if err != nil {
		return nil, err
	}

	c := &Server{
		baseURL:   u,
		appID:     appID,
		transport: t,
	}

	return c, nil
}

// Return the base URL
func (c *Server) BaseURL() *url.URL {
	return c.baseURL
}
