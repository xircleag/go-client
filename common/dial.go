package common

import (
	"net/http"
	"net/url"
)

type DialSettings struct {
	BaseURL *url.URL
	//Endpoint           string
	UserAgent          string
	BearerToken        string
	SessionToken       string
	Headers            map[string][]string
	ClientCredentials  *ClientCredentials
	Key                *Key
	AuthenticationFunc *func(bool, error)
	HTTPClient         *http.Client
}
