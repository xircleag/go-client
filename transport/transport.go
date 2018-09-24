package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/option"

	"golang.org/x/net/context"
	//"golang.org/x/net/context/ctxhttp"
)

var DefaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,

	// This is used to disable HTTP/2 due to a current nginx bug
	TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
}

type HTTPTransport struct {
	Session HTTPSessionMinter
	client  *http.Client
}

func (t *HTTPTransport) Do(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}

type HTTPSessionMinter interface {
	GetNonce(context.Context) (string, error)
	Token(context.Context) (string, error)
}

type httpTransport struct {
	ctx          context.Context
	userAgent    string
	headers      map[string][]string
	base         http.RoundTripper
	dialSettings *common.DialSettings
}

func (t httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.base
	if rt == nil {
		return nil, fmt.Errorf("No transport specified")
	}
	newReq := *req
	newReq.Header = t.headers
	for k, v := range req.Header {
		newReq.Header[k] = v
	}
	newReq.Header["User-Agent"] = []string{t.userAgent}
	return rt.RoundTrip(&newReq)
}

func (t httpTransport) GetNonce(ctx context.Context) (string, error) {
	return "", fmt.Errorf("This transport does not support obtaining nonces.")
}

func (t httpTransport) Token(ctx context.Context) (string, error) {
	return "", fmt.Errorf("This transport does not support authentication")
}

func NewHTTPTransport(ctx context.Context, appID string, baseURL *url.URL, websocketURL *url.URL, opts ...option.ClientOption) (*HTTPTransport, error) {
	var o common.DialSettings
	for _, opt := range opts {
		opt.Apply(&o)
	}

	if o.UserAgent == "" {
		// TODO: Inject proper version
		o.UserAgent = fmt.Sprintf("Layer Go Client version 0.0.1")
	}

	// Credentialed client
	if o.ClientCredentials != nil {
		o.ClientCredentials.ApplicationID = appID

		t := clientCredentialTransport{
			credentials:  o.ClientCredentials,
			baseURL:      baseURL,
			websocketURL: websocketURL,
			ctx:          ctx,
			userAgent:    o.UserAgent,
			headers:      o.Headers,
			base:         DefaultTransport,
		}

		return &HTTPTransport{
			Session: t,
			client:  &http.Client{Transport: t},
		}, nil
	}

	// Bearer token transport
	if o.BearerToken != "" {
		t := bearerTokenTransport{
			token:     o.BearerToken,
			baseURL:   baseURL,
			ctx:       ctx,
			userAgent: o.UserAgent,
			headers:   o.Headers,
			base:      DefaultTransport,
		}

		return &HTTPTransport{
			client: &http.Client{Transport: t},
		}, nil
	}

	// Fallback to a plain HTTP transport
	t := httpTransport{
		ctx:       ctx,
		userAgent: o.UserAgent,
		headers:   o.Headers,
		base:      DefaultTransport,
	}

	return &HTTPTransport{
		Session: t,
		client:  &http.Client{Transport: t},
	}, nil
}
