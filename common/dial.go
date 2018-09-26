package common

import (
	"net/url"
)

type DialSettings struct {
	BaseURL           *url.URL
	UserAgent         string
	BearerToken       string
	SessionToken      string
	Headers           map[string][]string
	ClientCredentials *ClientCredentials
	Key               *Key
	TokenFunc         func(user, nonce string) (token string, err error)
	AllowInsecure     bool
}
