package transport

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/option"

	"golang.org/x/net/context"
	//"golang.org/x/net/context/ctxhttp"
)

type HTTPTransport struct {
	Session HTTPSessionMinter
	client  *http.Client
}

func (t *HTTPTransport) Do(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}

type HTTPSessionMinter interface {
	Token() (string, error)
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

func (t httpTransport) Token() (string, error) {
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

	// See if we have supplied client credentials
	if o.ClientCredentials != nil {
		o.ClientCredentials.ApplicationID = appID

		t := clientCredentialTransport{
			credentials:  o.ClientCredentials,
			baseURL:      baseURL,
			websocketURL: websocketURL,
			ctx:          ctx,
			userAgent:    o.UserAgent,
			headers:      o.Headers,
			base:         http.DefaultTransport,
		}

		return &HTTPTransport{
			Session: t,
			client:  &http.Client{Transport: t},
		}, nil
	}

	// Fallback to a plain HTTP transport
	t := httpTransport{
		ctx:       ctx,
		userAgent: o.UserAgent,
		headers:   o.Headers,
		base:      http.DefaultTransport,
	}

	return &HTTPTransport{
		Session: t,
		client:  &http.Client{Transport: t},
	}, nil
}
