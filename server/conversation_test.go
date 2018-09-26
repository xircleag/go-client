package server

import (
	"context"
	"testing"

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
	return c.CreateConversation(context.Background(), []string{"test", "test1"}, false, metadata)
}

func TestCreateConversation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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
	if testing.Short() {
		t.Skip()
	}

	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	convo, err := createConversation(c)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Conversation(context.Background(), convo.UUID())
	if err != nil {
		t.Fatal(err)
	}
}

func TestConversationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

func TestConversationUpdateParticipants(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

func TestConversationUpdateMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

func TestConversationMarkRead(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// TODO: implement once we can write messages
}
