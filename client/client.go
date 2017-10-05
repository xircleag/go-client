package client

import (
	"fmt"
	"net/url"

	"github.com/layerhq/go-client/option"
	"github.com/layerhq/go-client/transport"

	"golang.org/x/net/context"
)

type NonceRequest struct {
	Nonce string `json:"nonce"`
}

type RESTClient struct {
	baseURL   *url.URL
	appID     string
	transport *transport.HTTPTransport
}

// NewClient creates a new Layer Client API client
func NewRESTClient(ctx context.Context, appID string, options ...option.ClientOption) (*RESTClient, error) {
	u, err := url.Parse("https://api.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building base URL: %v", err)
	}

	headers := map[string][]string{
		"Accept":       {"application/vnd.layer+json; version=3.0"},
		"Content-Type": {"application/json"},
	}

	options = append(options, option.WithHeaders(headers))

	t, err := transport.NewHTTPTransport(ctx, appID, u, options...)
	if err != nil {
		return nil, err
	}

	c := &RESTClient{
		baseURL:   u,
		appID:     appID,
		transport: t,
	}

	return c, nil
}

type WebsocketClient struct {
	Websocket *Websocket
	baseURL   *url.URL
	appID     string
	transport *transport.HTTPTransport
}

func NewWebsocketClient(ctx context.Context, appID string, options ...option.ClientOption) (*WebsocketClient, error) {
	wu, err := url.Parse("wss://websockets.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building websocket URL: %v", err)
	}

	return NewWebsocketTestClient(ctx, appID, *wu, options...)
}

func NewWebsocketTestClient(ctx context.Context, appID string, baseURL url.URL, options ...option.ClientOption) (*WebsocketClient, error) {
	u, err := url.Parse("https://api.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building base URL: %v", err)
	}

	headers := map[string][]string{
		"Accept":       {"application/vnd.layer+json; version=3.0"},
		"Content-Type": {"application/json"},
	}

	options = append(options, option.WithHeaders(headers))

	t, err := transport.NewHTTPTransport(ctx, appID, u, options...)
	if err != nil {
		return nil, err
	}

	c := &WebsocketClient{
		baseURL:   &baseURL,
		appID:     appID,
		transport: t,
	}
	c.Websocket = &Websocket{client: c}

	return c, nil
}

// Return the base URL
func (c *RESTClient) BaseURL() *url.URL {
	return c.baseURL
}

// Return the websocket URL
func (c *WebsocketClient) BaseURL() *url.URL {
	return c.baseURL
}
