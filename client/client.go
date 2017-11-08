package client

import (
	"fmt"
	"net/url"

	"github.com/layerhq/go-client/option"
	"github.com/layerhq/go-client/transport"

	"golang.org/x/net/context"
)

type Client struct {
	Websocket    *Websocket
	baseURL      *url.URL
	websocketURL *url.URL
	appID        string
	transport    *transport.HTTPTransport
}

type NonceRequest struct {
	Nonce string `json:"nonce"`
}

// NewClient creates a new Layer Client API Client
func NewClient(ctx context.Context, appID string, options ...option.ClientOption) (*Client, error) {
	u, err := url.Parse("https://api.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building base URL: %v", err)
	}

	wu, err := url.Parse("wss://websockets.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building websocket URL: %v", err)
	}

	return NewTestClient(u, wu, ctx, appID, options...)
}

func NewTestClient(u *url.URL, wu *url.URL, ctx context.Context, appID string, options ...option.ClientOption) (*Client, error) {

	headers := map[string][]string{
		"Accept":       {"application/vnd.layer+json; version=2.0"},
		"Content-Type": {"application/json"},
	}

	options = append(options, option.WithHeaders(headers))

	t, err := transport.NewHTTPTransport(ctx, appID, u, wu, options...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		baseURL:      u,
		websocketURL: wu,
		appID:        appID,
		transport:    t,
	}
	c.Websocket = &Websocket{client: c}

	return c, nil
}

// Return the base URL
func (c *Client) BaseURL() *url.URL {
	return c.baseURL
}

// Return the websocket URL
func (c *Client) WebsocketURL() *url.URL {
	return c.websocketURL
}
