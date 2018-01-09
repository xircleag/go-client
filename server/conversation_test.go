package server

import (
	"testing"
	"context"

	"github.com/layerhq/go-client/common"
)

func createConversation(c *Server) (*Conversation, error) {
	metadata := common.Metadata{}
	metadata.Set("stuff", "blah")
	metadata.Set("things", 5)
	other := common.Metadata{}
	other.Set("one", "1")
	other.Set("two", "2")
	metadata.Set("other", other)
	return c.CreateConversation(context.Background(), Conversation{
		Participants: []BasicIdentity{
			{ID: "test"},
			{ID: "test1"},
		},
		Distinct: false,
		Metadata: metadata,
	})
}

func TestCreateConversation(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	_, err = createConversation(c)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConversation(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	convo, err := createConversation(c)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Conversation(context.Background(), convo.ID())
	if err != nil {
		t.Fatal(err)
	}
}

func TestConversation_Delete(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	var convo *Conversation
	convo, err = createConversation(c)
	if err != nil {
		t.Fatal(err)
	}

	err = convo.Delete(context.Background())
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestConversation_UpdateParticipants(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	var convo *Conversation
	convo, err = createConversation(c)
	if err != nil {
		t.Fatal(err)
	}

	err = convo.UpdateParticipants(context.Background(), []ConversationEdit{
		{Operation: Add, Property: Participants, ID: "layer:///identities/test2"},
		{Operation: Remove, Property: Participants, ID: "layer:///identities/test"},
		{Operation: Set, Property: Participants, Value: []string{"layer:///identities/test", "layer:///identities/test1"}},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestConversation_UpdateMetadata(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	var convo *Conversation
	convo, err = createConversation(c)
	if err != nil {
		t.Fatal(err)
	}

	err = convo.UpdateMetadata(context.Background(), []ConversationEdit{
		{Operation: Set, Property: MetadataProperty("other", "one"), Value: "7"},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestConversation_MarkRead(t *testing.T) {
	// TODO: implement once we can write messages
}
