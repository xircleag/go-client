package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/layerhq/go-client/common"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

type ClientNonceRequest struct {
	Nonce string `json:"nonce"`
}

type clientCredentialTransport struct {
	credentials  *common.ClientCredentials
	baseURL      *url.URL
	websocketURL *url.URL
	ctx          context.Context
	token        string
	userAgent    string
	headers      map[string][]string
	base         http.RoundTripper
}

func (t clientCredentialTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.base
	if rt == nil {
		return nil, fmt.Errorf("No transport specified")
	}

	// See if need to generate a token
	if t.token == "" && req.URL.Path != "/nonces" && req.URL.Path != "/sessions" {
		token, err := t.Token()
		if err != nil {
			return nil, err
		}
		t.token = token
	}

	// Build the new request
	newReq := *req
	newReq.WithContext(t.ctx)
	newReq.Header = t.headers
	for k, v := range req.Header {
		newReq.Header[k] = v
	}
	newReq.Header["User-Agent"] = []string{t.userAgent}
	if t.token != "" {
		newReq.Header["Authorization"] = []string{fmt.Sprintf("Layer session-token=\"%s\"", t.token)}
	}

	// XXX
	//fmt.Println(fmt.Sprintf("newReq: %+v", newReq))

	return rt.RoundTrip(&newReq)
}

func (t clientCredentialTransport) getNonce() (string, error) {
	// Create the URL
	u, err := url.Parse("/nonces")
	if err != nil {
		return "", fmt.Errorf("Error building nonce URL: %v", err)
	}
	u = t.baseURL.ResolveReference(u)

	// Create the request
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("Error creating nonce request: %v", err)
	}
	req = req.WithContext(t.ctx)

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

func (t clientCredentialTransport) generateToken(nonce string) (string, error) {
	// Set claims
	claims := jwt.MapClaims{}
	claims["iss"] = t.credentials.ProviderID
	claims["prn"] = t.credentials.User
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	claims["nce"] = nonce

	// Create a token
	token := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)

	// Set header values
	token.Header["typ"] = "JWT"
	token.Header["alg"] = "RS256"
	token.Header["cty"] = "layer-eit;v=1"
	token.Header["kid"] = t.credentials.Key.ID

	// Build the keypair from the key
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(t.credentials.Key.KeyPair.Private))
	if err != nil {
		return "", fmt.Errorf("Error getting private key from keypair - %v", err)
	}

	// Sign and get the complete encoded token as a string
	ts, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("Error signing token - %v", err)
	}

	return ts, nil
}

func (t clientCredentialTransport) Token() (string, error) {
	var err error

	// TODO: Handle re-using an existing token

	// Get a nonce
	nonce, err := t.getNonce()
	if err != nil {
		return "", err
	}

	// Generate signed token
	if t.credentials.Token, err = t.generateToken(nonce); err != nil {
		return "", fmt.Errorf("Error getting token - %v", err)
	}

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