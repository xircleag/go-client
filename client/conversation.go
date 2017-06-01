package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/layerhq/go-client/iterator"

	"golang.org/x/net/context"
)

type Conversation struct {
	// ID uniquely identifies the conversation.
	ID string `json:"id,omitempty"`

	// URL is the URL for accessing the conversation via the Layer REST API.
	URL string `json:"url,omitempty"`

	// MessagesURL is the URL for access the conversation messages via the Layer.
	// REST API.
	MessagesURL string `json:"messages_url,omitempty"`

	// The time at which the conversation was initially created.
	CreatedAt time.Time `json:"-"`

	// LastMessage is a message object representing the last message sent in the
	// conversation.
	LastMessage *Message `json:"last_message,omitempty"`

	// Participants is an array of BasicIdentiy objects containing information on
	// the message participants.
	Participants []*BasicIdentity `json:"participants"`

	// Distinct represents whether this is a distinct conversation with the
	// specified participant list.
	Distinct bool `json:"distinct"`

	// The number of unread messages on the conversation for the user specified
	// by the client.
	UnreadMessageCount json.Number `json:"unread_message_count,omitempty"`

	// A generic interface available to store arbitrary metadata.
	Metadata json.RawMessage `json:"metadata,omitempty"`

	// Internal reference to the client object.
	client *Client `json:"-"`
}

func (c *Client) buildConversationURL(id string) (*url.URL, error) {
	var err error
	var u *url.URL
	if id != "" {
		u, err = url.Parse(fmt.Sprintf("/conversations/%s", id))
	} else {
		u, err = url.Parse("/conversations")
	}
	if err != nil {
		return nil, err
	}
	u = c.baseURL.ResolveReference(u)
	return u, nil
}

// Internal request to create a conversation
type conversationCreate struct {
	Participants []string    `json:"participants"`
	Distinct     bool        `json:"distinct"`
	Metadata     interface{} `json:"metadata,omitempty"`
}

// ConversationIterator returns a series of conversations
type ConversationIterator struct {
	ctx           context.Context
	client        *Client
	conversations []*Conversation
	current       int
	from          *string
	sort          *string
}

// Next returns the next slice of conversations
func (it *ConversationIterator) Next() (*Conversation, error) {
	it.current++
	if it.current > len(it.conversations) {
		// First try to get a new page
		conversations, err := it.client.ConversationsFrom(it.ctx, it.sort, it.from)
		if err != nil {
			return nil, err
		}
		if len(conversations) > 0 {
			it.conversations = conversations
			from := conversations[len(conversations)-1].ID
			it.from = &from
			it.current = 0
			return it.conversations[0], nil
		}

		// No more
		return nil, iterator.Done
	}
	it.conversations[it.current-1].client = it.client
	return it.conversations[it.current-1], nil
}

// Conversations gets all conversations for the user specified by the client connection, with a starting ID used for paging and iterations
func (c *Client) ConversationsFrom(ctx context.Context, sort *string, from *string) ([]*Conversation, error) {
	// Create the request URL
	u, err := c.buildConversationURL("")
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}

	// Create the request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)
	q := req.URL.Query()
	if sort != nil {
		q.Add("sort_by", *sort)
	}
	if from != nil {
		q.Add("from_id", *from)
	}
	q.Add("page_size", "100")
	req.URL.RawQuery = q.Encode()

	// Send the request
	res, err := c.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending conversation request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("Partially matching distinct conversation")
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	// Parse the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing conversation create response")
	}

	var conversations []*Conversation
	if err := json.Unmarshal(body, &conversations); err != nil {
		return nil, fmt.Errorf("Error parsing conversation create JSON: %v", err)
	}

	return conversations, nil
}

// Conversations gets all conversations for the user specified by the client connection
func (c *Client) Conversations(ctx context.Context, sort *string) (*ConversationIterator, error) {
	conversations, err := c.ConversationsFrom(ctx, sort, nil)
	if err != nil {
		return nil, iterator.Done
	}
	if len(conversations) <= 0 {
		return nil, fmt.Errorf("No conversations")
	}
	from := conversations[len(conversations)-1].ID
	return &ConversationIterator{
		ctx:           ctx,
		client:        c,
		conversations: conversations,
		sort:          sort,
		from:          &from,
	}, nil
}

// Conversation gets a single conversation for the user specified by the client connection
func (c *Client) Conversation(ctx context.Context, id string) (*Conversation, error) {
	u, err := c.buildConversationURL(id)
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}

	// Create the request JSON
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Conversation not found")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	// Parse the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing conversation create response")
	}

	var conversation *Conversation
	if err := json.Unmarshal(body, &conversation); err != nil {
		return nil, fmt.Errorf("Error parsing conversation JSON: %v", err)
	}
	conversation.client = c
	return conversation, nil
}

// CreateConversation creates a conversation for the user specified by the client connection
func (c *Client) CreateConversation(ctx context.Context, participants []string, distinct bool, metadata interface{}) (*Conversation, error) {
	// Create the request object
	cc := &conversationCreate{
		Participants: participants,
		Distinct:     distinct,
		Metadata:     metadata,
	}

	// Create the request URL
	u, err := c.buildConversationURL("")
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}

	// Create the request JSON
	query, err := json.Marshal(cc)
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation JSON: %v", err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("Partially matching distinct conversation")
	}

	if res.StatusCode == http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("Participant blocked")
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	// Parse the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing conversation create response")
	}

	var conversation *Conversation
	if err := json.Unmarshal(body, &conversation); err != nil {
		return nil, fmt.Errorf("Error parsing conversation create JSON: %v", err)
	}
	conversation.client = c
	return conversation, nil
}

func (convo *Conversation) Delete(ctx context.Context) error {
	return errors.New("Not implemented")
}

func (convo *Conversation) UpdateMetadata(ctx context.Context, metadata interface{}) error {
	return errors.New("Not implemented")
}

func (convo *Conversation) AddParticipants(ctx context.Context, participants []string) error {
	return errors.New("Not implemented")
}

func (convo *Conversation) RemoveParticipants(ctx context.Context, participants []string) error {
	return errors.New("Not implemented")
}
