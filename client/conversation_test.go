package client

import (
	"fmt"
	"testing"

	"github.com/layerhq/go-client/iterator"

	"golang.org/x/net/context"
)

func TestCreateConversation(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	convo, err := c.CreateConversation(ctx, []string{"123"}, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Created conversation with ID %s", convo.ID)
}

func TestGetConversations(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	convos, err := c.Conversations(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for {
		convo, err := convos.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(fmt.Sprintf("%+v", convo))
	}
}
