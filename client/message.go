package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/iterator"
)

const (
	MessageRecipientStatusSent      = "sent"
	MessageRecipientStatusDelivered = "delivered"
	MessageRecipientStatusRead      = "read"
)

type messageCreate struct {
	Parts        []*common.MessagePart       `json:"parts"`
	Notification *common.MessageNotification `json:"notification,omitempty"`
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (convo *Conversation) SendTextMessage(ctx context.Context, message string, notification *common.MessageNotification) (*common.Message, error) {
	msg := plaintextMessage(message)
	return convo.SendMessage(ctx, msg.Parts, notification)
}

// SendMessage sends a message on the current conversation
func (convo *Conversation) SendMessage(ctx context.Context, parts []*MessagePart, notification *common.MessageNotification) (*common.Message, error) {
	mc := &messageCreate{
		Parts:        parts,
		Notification: notification,
	}

	reqID := newRequestID()

	packet := &WebsocketPacket{
		Type: "request",
		Body: WebsocketRequest{
			Method:    WebsocketMessageCreate,
			RequestID: reqID,
			ObjectID:  convo.ID,
			Data:      mc,
		},
	}

	result := make(chan *common.Message)

	// Register a handler for the response
	unsub := convo.Client.Websocket.HandleFunc(WebsocketMessageCreate, func(w *Websocket, p *WebsocketPacket) {
		resp, ok := p.Body.(*WebsocketResponse)
		if !ok || resp.RequestID != reqID {
			return
		}

		message, _ := resp.Data.(*common.Message)
		result <- message
	})
	defer unsub.Remove()
	go convo.Client.Websocket.Listen(ctx)

	timer := getTimer(ctx)

	if err := convo.Client.Websocket.Send(ctx, packet); err != nil {
		return nil, err
	}

	var message *common.Message
	select {
	case message = <-result:
		return message, nil
	case <-timer.C:
		return nil, ErrTimedOut
	}
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (convo *Conversation) SendTextMessageREST(ctx context.Context, message string, notification *MessageNotification) (*common.Message, error) {
	msg := plaintextMessage(message)
	return convo.SendMessageREST(ctx, msg.Parts, notification)
}

// SendMessage sends a message on the current conversation
func (convo *Conversation) SendMessageREST(ctx context.Context, parts []*common.MessagePart, notification *common.MessageNotification) (*common.Message, error) {
	mc := &messageCreate{
		Parts:        parts,
		Notification: notification,
	}

	// Build the URL
	convoID := common.UUIDFromLayerURL(convo.ID)
	u, err := url.Parse(fmt.Sprintf("/conversations/%s/messages", convoID))
	if err != nil {
		return nil, fmt.Errorf("Error building conversation message URL: %v", err)
	}
	u = convo.Client.baseURL.ResolveReference(u)

	// Create the request
	query, err := json.Marshal(mc)
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation JSON: %v", err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(query))
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

	if res.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("The requested message already exists")
	}
	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	// Parse the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing response")
	}

	var message *common.Message
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("Error parsing message JSON: %v", err)
	}
	return message, nil
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
		messages, err := it.conversation.MessagesFrom(it.ctx, it.from)
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
func (convo *Conversation) MessagesFrom(ctx context.Context, from string) ([]*common.Message, error) {
	// Create the request URL
	convoID := common.UUIDFromLayerURL(convo.ID)
	u, err := url.Parse(fmt.Sprintf("/conversations/%s/messages", convoID))
	if err != nil {
		return nil, fmt.Errorf("Error building conversation message URL: %v", err)
	}
	u = convo.Client.baseURL.ResolveReference(u)

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
	q.Add("page_size", "100")
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
	err = json.NewDecoder(res.Body).Decode(&messages)
	return messages, err
}

// Messages gets all messages on a conversation
func (convo *Conversation) Messages(ctx context.Context) (*MessageIterator, error) {
	messages, err := convo.MessagesFrom(ctx, "")
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
