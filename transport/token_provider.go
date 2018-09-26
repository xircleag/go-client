package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/layerhq/go-client/common"

	"golang.org/x/net/context"
)

type ClientNonceRequest struct {
	Nonce string `json:"nonce"`
}

type TokenFactory func(user, nonce string) (token string, err error)

type tokenProviderTransport struct {
	tokenFactory TokenFactory
	tokenTimeout time.Duration
	token        string
	tokenMu      *sync.Mutex
	credentials  *common.ClientCredentials
	baseURL      *url.URL
	websocketURL *url.URL
	ctx          context.Context
	userAgent    string
	headers      map[string][]string
	base         http.RoundTripper
}

func (t *tokenProviderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.base
	if rt == nil {
		return nil, fmt.Errorf("No transport specified")
	}

	// See if need to generate a token
	if t.token == "" && req.URL.Path != "/nonces" && req.URL.Path != "/sessions" {
		_, err := t.Token(context.Background())
		if err != nil {
			return nil, err
		}
	}

	// Build the new request
	newReq := req
	newReq.WithContext(t.ctx)
	for k, v := range t.headers {
		newReq.Header.Del(k)
		for _, val := range v {
			newReq.Header.Add(k, val)
		}
	}
	for k, v := range req.Header {
		newReq.Header.Del(k)
		for _, val := range v {
			newReq.Header.Add(k, val)
		}
	}
	newReq.Header.Del("User-Agent")
	newReq.Header.Add("User-Agent", t.userAgent)
	if t.token != "" {
		newReq.Header.Del("Authorization")
		newReq.Header.Add("Authorization", fmt.Sprintf("Layer session-token=\"%s\"", t.token))
	}

	return rt.RoundTrip(newReq)
}

func (t *tokenProviderTransport) GetNonce(ctx context.Context) (string, error) {
	// Create the URL
	u, err := url.Parse("/nonces")
	if err != nil {
		return "", fmt.Errorf("Error building nonce URL: %v", err)
	}
	u = t.baseURL.ResolveReference(u)

	// Create the request
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("Error creating nonce request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := t.RoundTrip(req)
	if err != nil {
		return "", fmt.Errorf("Error getting nonce: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Error getting nonce: status code is %d", res.StatusCode)
	}

	// Parse the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Error parsing nonce response")
	}

	var data ClientNonceRequest
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("Error parsing nonce JSON")
	}
	if data.Nonce != "" {
		return data.Nonce, nil
	}

	return "", fmt.Errorf("Error parsing nonce JSON")
}

func (t *tokenProviderTransport) Token(ctx context.Context) (string, error) {
	t.tokenMu.Lock()
	defer t.tokenMu.Unlock()

	// Check if we have an existing valid token
	if t.token == "" {
		var err error
		t.token, err = t.getToken(ctx)
		if err != nil {
			return "", err
		}
	}
	return t.token, nil
}

func (t *tokenProviderTransport) getToken(ctx context.Context) (string, error) {
	var err error

	// Get a nonce
	nonce, err := t.GetNonce(context.Background())
	if err != nil {
		return "", err
	}

	// Get the signed token
	tokenCh := make(chan string)
	errCh := make(chan error)

	if t.credentials == nil {
		return "", fmt.Errorf("No username credentials have been specified")
	}

	go func(nonce string, user string, factory TokenFactory) {
		token, err := factory(user, nonce)
		if err != nil {
			errCh <- err
		}
		tokenCh <- token
	}(nonce, t.credentials.User, t.tokenFactory)

	select {
	case token := <-tokenCh:
		t.credentials.Token = token
	case err := <-errCh:
		return "", err
	case <-time.After(t.tokenTimeout):
		return "", fmt.Errorf("Timeout in token factory")
	}

	// Build the session request
	u, err := url.Parse("/sessions")
	if err != nil {
		return "", fmt.Errorf("Error building nonce URL: %v", err)
	}
	u = t.baseURL.ResolveReference(u)

	// Create the request
	query, err := json.Marshal(t.credentials)
	if err != nil {
		return "", fmt.Errorf("Error creating session JSON - %v", err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(query))
	if err != nil {
		return "", fmt.Errorf("Error creating session request - %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := t.RoundTrip(req)
	if err != nil {
		return "", fmt.Errorf("Error getting session - %v", err)
	}
	defer res.Body.Close()

	// Parse the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Error parsing session response")
	}

	if res.StatusCode != http.StatusCreated {
		var resError common.RequestError
		err := json.Unmarshal(body, &resError)
		if err == nil {
			return "", resError
		}
		return "", fmt.Errorf("Error getting session - status code is %d", res.StatusCode)
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("Error parsing session JSON")
	}

	if session, ok := data.(map[string]interface{})["session_token"]; ok {
		return session.(string), nil
	}

	return "", fmt.Errorf("Error parsing session JSON")
}
