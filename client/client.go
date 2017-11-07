package client

import (
	"fmt"
	"net/url"

	"github.com/layerhq/go-client/option"
	"github.com/layerhq/go-client/transport"

	"golang.org/x/net/context"
	"time"
	"errors"
	"github.com/satori/go.uuid"
)

type NonceRequest struct {
	Nonce string `json:"nonce"`
}

type RESTClient struct {
	baseURL   *url.URL
	appID     string
	transport *transport.HTTPTransport
}

var errTimedOut = errors.New("Operation timed out.")

// NewClient creates a new Layer Client API client
func NewRESTClient(ctx context.Context, appID string, options ...option.ClientOption) (*RESTClient, error) {
	u, err := url.Parse("https://api.layer.com")
	if err != nil {
		return nil, fmt.Errorf("Error building base URL: %v", err)
	}

	return NewRESTTestClient(ctx, u, appID, options...)
}

func NewRESTTestClient(ctx context.Context, u *url.URL, appID string, options ...option.ClientOption) (*RESTClient, error) {
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

	return NewWebsocketTestClient(ctx, *wu, appID, options...)
}

func NewWebsocketTestClient(ctx context.Context, baseURL url.URL, appID string, options ...option.ClientOption) (*WebsocketClient, error) {
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

func newRequestID() string {
	return uuid.NewV1().String()
}

func getTimer(ctx context.Context) (timer *time.Timer) {
	deadline, present := ctx.Deadline()
	if present {
		timer = time.NewTimer(deadline.Sub(time.Now()))
	} else {
		// default timeout of 30 seconds
		timer = time.NewTimer(time.Duration(30) * time.Second)
	}
	return
}
