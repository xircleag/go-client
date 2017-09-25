package client

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/layerhq/go-client/iterator"

	"golang.org/x/net/context"
)

func TestPlaintextMessage(t *testing.T) {
	msg := plaintextMessage("Test")
	if len(msg.Parts) != 1 {
		t.Fatal(errors.New("Invalid message parts length"))
	}
	if msg.Parts[0].MimeType != "text/plain" {
		t.Fatal(errors.New("Invalid message part content type - expected text/plain"))
	}
	if msg.Parts[0].Body != "Test" {
		t.Fatal("Invalid message part body")
	}
}

func TestSendTextMessage(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	convo, err := c.CreateConversation(ctx, []string{"123"}, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := convo.SendTextMessage(ctx, "Testing", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Created message with ID %s", msg.ID)
}

func TestSendTextMessageExistingConversation(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	convo, err := c.Conversation(ctx, "2ab5971a-bb34-45b9-bada-ed86f25c6327")
	if err != nil {
		t.Fatal(err)
	}
	msg, err := convo.SendTextMessage(ctx, "Testing", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Created message with ID %s", msg.ID)
}

func TestGetMessages(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	convos, err := c.Conversations(ctx, "")
	convo, err := convos.Next()
	if err != nil {
		t.Fatal(err)
	}
	messages, err := convo.Messages(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for {
		message, err := messages.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, part := range message.Parts {
			if part.MimeType == "text/plain" {
				fmt.Println(fmt.Sprintf("%+s %+v: %s",
					message.SentAt.Format("2006-01-02T15:04:05Z07:00"),
					message.Sender,
					strings.TrimSpace(part.Body)))
			}
		}
	}
}
