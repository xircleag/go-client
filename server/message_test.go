package server

import (
	"context"
	"testing"
)

func TestCreateTextMessage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	conversation, err := createConversation(c)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = conversation.SendTextMessage(ctx, "layer:///identities/test", "Test Message", nil)
	if err != nil {
		t.Fatal(err)
	}
}
