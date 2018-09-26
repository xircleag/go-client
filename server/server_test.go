package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/option"
)

var (
	testClient *Server
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

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

func createTestClient() (*Server, error) {
	// Load credentials
	path := os.Getenv("LAYER_TESTING_CREDENTIALS")
	if path == "" {
		return nil, fmt.Errorf("LAYER_TESTING_CREDENTIALS path is not set")
	}

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

	return NewClient(ctx, c.ApplicationID, option.AllowInsecure(), option.WithBearerToken(c.APIKey))
}

func TestCreateClientWithBearerToken(t *testing.T) {
	_, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}
}

func ExampleNewClient() {
	ctx := context.Background()

	// Create a new instance of the Server API
	client, err := NewClient(ctx, "APPLICATION_ID", option.WithBearerToken("API_KEY"))
	if err != nil {
		fmt.Println(fmt.Sprintf("Error creating Layer Server API client: %v", err))
	}

	// Create a new conversation
	convo, err := client.CreateConversation(ctx, []string{"recipient1", "recipient2"}, false, common.Metadata{})
	if err != nil {
		fmt.Println(fmt.Sprintf("Error creating conversation: %v", err))
	}
	fmt.Println(fmt.Sprintf("%+v", convo))
}
