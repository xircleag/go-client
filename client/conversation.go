package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/layerhq/go-client/iterator"

	"golang.org/x/net/context"
)

type Conversation struct {
	client             *Client          `json:"-"`
	ID                 string           `json:"id,omitempty"`
	URL                string           `json:"url,omitempty"`
	MessagesURL        string           `json:"messages_url,omitempty"`
	CreatedAt          time.Time        `json:"-"`
	LastMessage        *Message         `json:"last_message,omitempty"`
	Participants       []*BasicIdentity `json:"participants"`
	Distinct           bool             `json:"distinct"`
	UnreadMessageCount json.Number      `json:"unread_message_count,omitempty"`
	Metadata           json.RawMessage  `json:"metadata,omitempty"`
}

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
	return it.conversations[it.current-1], nil
}

// Conversations gets all conversations for the user specified by the client connection, with a starting ID used for paging and iterations
func (c *Client) ConversationsFrom(ctx context.Context, sort *string, from *string) ([]*Conversation, error) {
	// Create the request URL
	u, err := url.Parse("/conversations")
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}
	u = c.baseURL.ResolveReference(u)

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
	q.Add("page_size", "5")
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
		return nil, err
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
	return nil, nil
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
	u, err := url.Parse("/conversations")
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}
	u = c.baseURL.ResolveReference(u)

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

	return conversation, nil
}

func (convo *Conversation) Delete(ctx context.Context) error {
	return nil
}

func (convo *Conversation) UpdateMetadata(ctx context.Context, metadata interface{}) error {
	return nil
}

func (convo *Conversation) AddParticipants(ctx context.Context, participants []string) error {
	return nil
}

func (convo *Conversation) RemoveParticipants(ctx context.Context, participants []string) error {
	return nil
}
