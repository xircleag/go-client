package server

import (
	"context"
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/url"
	//"os"
	"testing"

	//"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/option"
	"github.com/layerhq/go-client/common"
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

func createTestClient() (*Server, error) {
	// Load credentials
	/*
	path := os.Getenv("TESTING_CREDENTIALS")
	if path == "" {
		return nil, fmt.Errorf("TESTING_CREDENTIALS path is not set")
	}
	*/
	path := "/Users/adam/workspace/test/staging1-creds.json"

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c *common.Certificate
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	u, _ := url.Parse("https://staging-api.layer.com/apps/" + c.AppID + "/")
	return NewTestClient(ctx, u, c.AppID, option.WithBearerToken(c.APIKey))
}

func TestCreateClientWithBearerToken(t *testing.T) {
	_, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created client API object")
}
