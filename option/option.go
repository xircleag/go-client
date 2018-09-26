// Package option manages client and transport options.
package option

import (
	"net/url"

	"github.com/layerhq/go-client/common"
)

// A ClientOption is an option for Layer API client
type ClientOption interface {
	Apply(*common.DialSettings)
}

func OverrideURL(u *url.URL) ClientOption {
	return overrideURL{u}
}

type overrideURL struct{ baseURL *url.URL }

func (o overrideURL) Apply(s *common.DialSettings) {
	s.BaseURL = o.baseURL
}

// AllowInsecure skips TLS verification (this is very likely only useful
// during testing)
func AllowInsecure() ClientOption {
	return allowInsecure{}
}

type allowInsecure struct{}

func (o allowInsecure) Apply(s *common.DialSettings) {
	s.AllowInsecure = true
}

// WithHeaders returns a ClientOption that specified a header map
func WithHeaders(headers map[string][]string) ClientOption {
	return withHeaders{headers}
}

type withHeaders struct{ headers map[string][]string }

func (w withHeaders) Apply(s *common.DialSettings) {
	s.Headers = w.headers
}

// WithBearerToken returns a ClientOption that specifies a bearer token
// string to be used for authentication.
func WithBearerToken(token string) ClientOption {
	return withBearerToken{token}
}

type withBearerToken struct{ token string }

func (w withBearerToken) Apply(s *common.DialSettings) {
	s.BearerToken = w.token
}

// WithSessionToken returns a ClientOption that specifies a session token
// string to be used for authentication.
func WithSessionToken(token string) ClientOption {
	return withSessionToken{token}
}

type withSessionToken struct{ token string }

func (w withSessionToken) Apply(s *common.DialSettings) {
	s.SessionToken = w.token
}

// WithTokenFunc returns a ClientOption that specifies a token minting
// function.  This would normally be used to call an external identity
// service on a remote server.
func WithTokenFunc(tokenFunc func(string, string) (string, error)) ClientOption {
	return withTokenFunc{tokenFunc}
}

type withTokenFunc struct {
	tokenFunc func(user, nonce string) (token string, err error)
}

func (w withTokenFunc) Apply(s *common.DialSettings) {
	s.TokenFunc = w.tokenFunc
}

// WithCredentials returns a ClientOption that specifies explicit client
// credentials to be used for authentication.
func WithCredentials(c *common.ClientCredentials) ClientOption {
	return withCredentials{c}
}

type withCredentials struct{ credentials *common.ClientCredentials }

func (w withCredentials) Apply(s *common.DialSettings) {
	s.ClientCredentials = w.credentials
}
