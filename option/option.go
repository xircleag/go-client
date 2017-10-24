// A package to manage client options
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
	return overrideURL{ u }
}

type overrideURL struct { baseURL *url.URL}

func (o overrideURL) Apply(s *common.DialSettings) {
	s.BaseURL = o.baseURL
}

// WithHeaders returns a ClientOption that specified a header map
func OverrideURL(baseURL *url.URL) ClientOption {
	return overrideURL{baseURL: baseURL}
}

type overrideURL struct{ baseURL *url.URL }

func (o overrideURL) Apply(s *common.DialSettings) {
	s.BaseURL = o.baseURL
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

// WithAuthenticationFunc returns a ClientOption that accepts an
// authentication function.
func WithAuthenticationFunc(f func(bool, error)) ClientOption {
	return withAuthenticationFunc{f}
}

type withAuthenticationFunc struct{ f func(bool, error) }

func (w withAuthenticationFunc) Apply(s *common.DialSettings) {
	s.AuthenticationFunc = &w.f
}

// WithCredentials returns a ClientOption that specifies explicit client
// credentials to be used for authenticaion.
func WithCredentials(c *common.ClientCredentials) ClientOption {
	return withCredentials{c}
}

type withCredentials struct{ credentials *common.ClientCredentials }

func (w withCredentials) Apply(s *common.DialSettings) {
	s.ClientCredentials = w.credentials
}
