package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/option"

	"golang.org/x/net/context"
)

type TestingCredentials struct {
	ApplicationID string                 `json:"application_id"`
	ProviderID    string                 `json:"provider_id"`
	AccountID     string                 `json:"account_id"`
	Key           *TestingCredentialsKey `json:"key"`
}

type TestingCredentialsKey struct {
	ID      string `json:"id"`
	Private string `json:"private"`
	Public  string `json:"public"`
}

func createTestClient() (*Client, error) {
	// Load credentials
	path := os.Getenv("TESTING_CREDENTIALS")
	if path == "" {
		return nil, fmt.Errorf("TESTING_CREDENTIALS path is not set")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c *TestingCredentials
	json.Unmarshal(data, &c)

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
	fmt.Println(fmt.Sprintf("REQUEST: %+v", req))

	// Send the request
	fmt.Println(fmt.Sprintf("%+v", req))
	res, err := c.transport.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	fmt.Println(fmt.Sprintf("RESULT: %+v", res))

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("BODY: %+v", body))
}
