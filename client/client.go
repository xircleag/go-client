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

// NewClient creates a new Layer Client API client
func NewClient(ctx context.Context, appID string, options ...option.ClientOption) (*Client, error) {
	u, err := url.Parse("https://api.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building base URL: %v", err)
	}

	wu, err := url.Parse("wss://websockets.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building websocket URL: %v", err)
	}

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
