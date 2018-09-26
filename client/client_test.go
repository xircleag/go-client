package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/option"

	"golang.org/x/net/context"
)

type TestingCredentials struct {
	ApplicationID string                 `json:"application_id"`
	ProviderID    string                 `json:"provider_id"`
	AccountID     string                 `json:"account_id"`
	APIKey        string                 `json:"api_key"`
	Key           *TestingCredentialsKey `json:"key"`
}

type TestingCredentialsKey struct {
	ID      string `json:"id"`
	Private string `json:"private"`
	Public  string `json:"public"`
}

type TestingIdentityRequest struct {
	Nonce string                      `json:"nonce"`
	User  *TestingIdentityRequestUser `json:"user"`
}

type TestingIdentityRequestUser struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type TestingIdentityResponse struct {
	AuthenticationToken string `json:"authentication_token"`
	Token               string `json:"layer_identity_token"`
	User                struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}
}

func createTokenTestClient() (*Client, error) {
	ctx := context.Background()
	return NewClient(
		ctx,
		common.UUIDFromLayerURL(os.Getenv("LAYER_TESTING_APPID")),
		option.WithCredentials(&common.ClientCredentials{
			User: os.Getenv("LAYER_TESTING_USER"),
		}),
		option.WithTokenFunc(func(user, nonce string) (token string, err error) {

			identity := &TestingIdentityRequest{
				Nonce: nonce,
				User: &TestingIdentityRequestUser{
					Email:    user,
					Password: os.Getenv("LAYER_TESTING_PASSWORD"),
				},
			}
			data, err := json.Marshal(identity)
			req, err := http.NewRequest("POST", "https://layer-identity-provider.herokuapp.com/users/sign_in.json", bytes.NewBuffer(data))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x_layer_app_id", os.Getenv("LAYER_TESTING_APPID"))

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer res.Body.Close()
			body, _ := ioutil.ReadAll(res.Body)

			var identityResponse TestingIdentityResponse
			err = json.Unmarshal(body, &identityResponse)
			if err != nil {
				return "", err
			}

			return identityResponse.Token, nil
		}),
	)
}

func createTestClient() (*Client, error) {
	// Load credentials
	path := os.Getenv("LAYER_TESTING_CREDENTIALS")
	if path == "" {
		return nil, fmt.Errorf("LAYER_TESTING_CREDENTIALS path is not set")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c *TestingCredentials
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return NewClient(ctx, c.ApplicationID, option.WithCredentials(&common.ClientCredentials{
		User:       "test",
		ProviderID: c.ProviderID,
		AccountID:  c.AccountID,
		Key: &common.Key{
			ID: c.Key.ID,
			KeyPair: &common.KeyPair{
				Private: c.Key.Private,
			},
		},
	}))
}

func TestCreateClientWithCredentials(t *testing.T) {
	_, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created client API object")
}

func TestCreateClientWithTokenFunc(t *testing.T) {
	_, err := createTokenTestClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created client API object")
}

func TestAuhenticatedGet(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Create the URL
	u, err := url.Parse("/ping")
	if err != nil {
		t.Fatal(err)
	}
	u = c.baseURL.ResolveReference(u)

	req, err := http.NewRequest("HEAD", u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.transport.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
}

func TestCachedToken(t *testing.T) {
	c, err := createTokenTestClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created client API object")

	var sendRequest = func() error {
		ctx := context.Background()

		// Create the URL
		u, err := url.Parse("/ping")
		if err != nil {
			return err
		}
		u = c.baseURL.ResolveReference(u)

		req, err := http.NewRequest("HEAD", u.String(), nil)
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)

		// Send the request
		res, err := c.transport.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		fmt.Println(fmt.Sprintf("TRANSPORT: %+v", c.transport))
		return nil
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sendRequest()
		wg.Done()
	}()
	go func() {
		sendRequest()
		wg.Done()
	}()
	wg.Wait()
}

func ExampleNewClient() {
	ctx := context.Background()
	return NewClient(ctx, "APP_ID", option.WithCredentials(&common.ClientCredentials{User: "USERNAME"}), option.WithTokenFunc(func(user, nonce string) (token string, err error) {
		// Make an HTTP call or perform local logic to create a signed JWT
		// with your private key.
		//
		// More details on this process can be found at:
		//   https://docs.layer.com/reference/client_api/authentication.out
		return
	}))
}
