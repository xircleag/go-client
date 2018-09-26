package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/iterator"
)

// MessageSchedule defines a delayed message schedule
type MessageSchedule struct {
	DelayInSeconds           float64 `json:"delay_in_seconds"`
	TypingIndicatorInSeconds float64 `json:"typing_indicator_in_seconds"`
}

// MessageCreate contains detail on a new message
type MessageCreate struct {
	SenderID     string                      `json:"sender_id"`
	Parts        []*common.MessagePart       `json:"parts"`
	Notification *common.MessageNotification `json:"notification,omitempty"`
	Schedule     *MessageSchedule            `json:"schedule,omitempty"`
}

// plaintextMessage is a helper function that returns a Message with a single "text/plain" message part
func plaintextMessage(content string) *common.Message {
	return &common.Message{
		Parts: []*common.MessagePart{
			&common.MessagePart{
				Body:     content,
				MimeType: "text/plain",
			},
		},
	}
}

func (convo *Conversation) buildMessageURL(id string) (u *url.URL, err error) {
	u, err = url.Parse(strings.TrimSuffix(fmt.Sprintf("conversations/%s/messages/%s", convo.UUID(), id), "/"))
	if err != nil {
		return
	}
	u = convo.Client.baseURL.ResolveReference(u)
	return
}

// SendMessage sends a message to the server
func (convo *Conversation) SendMessage(ctx context.Context, sender string, parts []*common.MessagePart, notification *common.MessageNotification, schedule *MessageSchedule) (*common.Message, error) {
	if convo.Client == nil {
		return nil, errors.New("Client not set in conversation")
	}

	// Build the URL
	u, err := convo.buildMessageURL("")
	if err != nil {
		return nil, fmt.Errorf("Error building message URL: %v", err)
	}

	mc := &MessageCreate{
		SenderID:     sender,
		Parts:        parts,
		Notification: notification,
		Schedule:     schedule,
	}

	// Create the request
	query, err := json.Marshal(mc)
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := convo.Client.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %v", err)
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusConflict:
		return nil, fmt.Errorf("Requested message already exists")
	case res.StatusCode == http.StatusAccepted:
		return nil, nil
	case res.StatusCode != http.StatusCreated:
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	var message *common.Message
	err = json.NewDecoder(res.Body).Decode(&message)
	return message, err
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (convo *Conversation) SendTextMessage(ctx context.Context, sender string, message string, notification *common.MessageNotification) (*common.Message, error) {
	msg := plaintextMessage(message)
	return convo.SendMessage(ctx, sender, msg.Parts, notification, nil)
}

// SendMessage sends a message batch
func (convo *Conversation) SendMessageBatch(ctx context.Context, mc []MessageCreate) error {
	if convo.Client == nil {
		return errors.New("Client not set in conversation")
	}
	if len(mc) > 12 {
		return errors.New("A maximum of 12 messages are supported")
	}

	for i := range mc {
		mc[i].SenderID = common.LayerURL(common.IdentitiesName, mc[i].SenderID)
	}

	// Build the URL
	u, err := convo.buildMessageURL("")
	if err != nil {
		return fmt.Errorf("Error building message URL: %v", err)
	}

	// Create the request
	query, err := json.Marshal(mc)
	if err != nil {
		return fmt.Errorf("Error creating conversation JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := convo.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Status is %d", res.StatusCode)
	}

	return nil
}

// MessageIterator returns a series of messages
type MessageIterator struct {
	ctx          context.Context
	conversation *Conversation
	messages     []*common.Message
	current      int
	from         string
	sort         string
}

// Next returns the next slice of messages
func (it *MessageIterator) Next() (*common.Message, error) {
	it.current++
	if it.current > len(it.messages) {
		// First try to get a new page
		messages, err := it.conversation.MessagesFrom(it.ctx, it.from, 0)
		if err != nil {
			return nil, err
		}
		if len(messages) > 0 {
			it.messages = messages
			from := messages[len(messages)-1].ID
			it.from = from
			it.current = 0
			return it.messages[0], nil
		}

		// No more
		return nil, iterator.Done
	}
	return it.messages[it.current-1], nil
}

// MessagesFrom gets all messages on a conversation from the specified offset
func (convo *Conversation) MessagesFrom(ctx context.Context, from string, pageSize int) ([]*common.Message, error) {
	// Build the URL
	u, err := convo.buildMessageURL("")
	if err != nil {
		return nil, fmt.Errorf("Error building message URL: %v", err)
	}

	// Create the request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)
	q := req.URL.Query()
	if from != "" {
		q.Add("from_id", from)
	}
	pSize := "100"
	if pageSize != 0 {
		pSize = strconv.Itoa(pageSize)
	}
	q.Add("page_size", pSize)

	req.URL.RawQuery = q.Encode()

	// Send the request
	res, err := convo.Client.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	var messages []*common.Message
	err = json.NewDecoder(res.Body).Decode(messages)
	return messages, err
}

// Messages gets all messages on a conversation
func (convo *Conversation) Messages(ctx context.Context) (*MessageIterator, error) {
	messages, err := convo.MessagesFrom(ctx, "", 0)
	if err != nil {
		return nil, err
	}
	if len(messages) <= 0 {
		return nil, iterator.Done
	}
	from := messages[len(messages)-1].ID
	return &MessageIterator{
		ctx:          ctx,
		conversation: convo,
		messages:     messages,
		from:         from,
	}, nil
}

// DeleteMessage deletes a message
func (convo *Conversation) DeleteMessage(ctx context.Context, messageID string) error {
	// Build the URL
	u, err := convo.buildMessageURL(messageID)
	if err != nil {
		return fmt.Errorf("Error building message URL: %v", err)
	}

	// Create the request
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := convo.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Status code is %d", res.StatusCode)
	}

	return nil
}
