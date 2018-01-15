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

type Message struct {
	// ID uniquely identifies the message.
	ID string `json:"id,omitempty"`

	// URL is the URL for accessing the conversation via the Layer REST API.
	URL string `json:"url,omitempty"`

	// The URL for the message receipt status.
	ReceiptsURL string `json:"receipts_url,omitempty"`

	// Per-Client ordering of the message in the conversation.
	Position json.Number `json:"-"`

	// Conversation that the message is part of.
	Conversation *Conversation `json:"conversation,omitempty"`

	// An array of message parts.
	Parts []*MessagePart `json:"parts,omitempty"`

	// The time at which the message was sent.
	SentAt time.Time `json:"sent_at"`

	// The identity of the message sender.
	Sender *BasicIdentity `json:"sender,omitempty"`

	// Indicates if the user has read the message.
	Unread bool `json:"is_unread,omitempty"`

	// A map of identity URLs and message status (sent, delivered, read).
	RecipientStatus map[string]string `json:"recipient_status,omitempty"`
}

type MessagePart struct {
	// The message text.
	Body string `json:"body"`

	// The MIME type of the part ("text/plain", "image/png", etc.).
	MimeType string `json:"mime_type"`

	// "base64" if the Body is Base64 encoded
	Encoding string `json:"encoding,omitempty"`

	// Content is set if the message part contains over 2KB of data, and
	// contains data on the external data.
	Content *MessagePartContent `json:"content,omitempty"`
}

type MessagePartContent struct {
	// ID uniquely identifies the message part external content.
	ID string `json:"id"`

	// The URL at which the external content data can be access.
	DownloadURL string `json:"download_url"`

	// The date and time at which the DownloadURL expires.
	Expiration time.Time

	// URL to call to refresh the DownloadURL upon expiration.
	RefreshURL string `json:"refresh_url,omitempty"`

	// The size in bytes of the content payload.
	Size json.Number
}

type MessageNotification struct {
	// The title of the notification that will be presented with the notification.
	Title string `json:"title,omitempty"`

	// The text body that will be presented with the notification.
	Text string `json:"text,omitempty"`

	// The optional sound that will be played with the notification.
	Sound string `json:"sound,omitempty"`
}

type messageCreate struct {
	Parts        []*MessagePart       `json:"parts"`
	Notification *MessageNotification `json:"notification,omitempty"`
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (convo *Conversation) SendTextMessage(ctx context.Context, message string, notification *MessageNotification) (*Message, error) {
	msg := plaintextMessage(message)
	return convo.SendMessage(ctx, msg.Parts, notification)
}

// SendMessage sends a message on the current conversation
func (convo *Conversation) SendMessage(ctx context.Context, parts []*MessagePart, notification *MessageNotification) (*Message, error) {
	mc := &messageCreate{
		Parts:        parts,
		Notification: notification,
	}

	reqID := newRequestID()

	packet := &WebsocketPacket{
		Type: "request",
		Body: WebsocketRequest{
			Method:    WebsocketChangeMessageCreate,
			RequestID: reqID,
			ObjectID:  convo.ID,
			Data:      mc,
		},
	}

	result := make(chan *Message)

	// register a handler for the response
	unsub := convo.Client.Websocket.HandleFunc(WebsocketChangeMessageCreate, func(w *Websocket, p *WebsocketPacket) {
		resp, ok := p.Body.(*WebsocketResponse)
		if !ok || resp.RequestID != reqID {
			return
		}

		message, _ := resp.Data.(*Message)
		result <- message
	})
	defer unsub.Remove()
	go convo.Client.Websocket.Listen(ctx)

	timer := getTimer(ctx)

	if err := convo.Client.Websocket.Send(ctx, packet); err != nil {
		return nil, err
	}

	var message *Message
	select {
	case message = <-result:
		return message, nil
	case <-timer.C:
		return nil, ErrTimedOut
	}
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (convo *Conversation) SendTextMessageREST(ctx context.Context, message string, notification *MessageNotification) (*Message, error) {
	msg := plaintextMessage(message)
	return convo.SendMessageREST(ctx, msg.Parts, notification)
}

// SendMessage sends a message on the current conversation
func (convo *Conversation) SendMessageREST(ctx context.Context, parts []*MessagePart, notification *MessageNotification) (*Message, error) {
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

	var message *Message
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("Error parsing message JSON: %v", err)
	}
	return message, nil
}

// plaintextMessage is a helper function that returns a Message with a single "text/plain" message part
func plaintextMessage(content string) *Message {
	return &Message{
		Parts: []*MessagePart{
			&MessagePart{
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
	messages     []*Message
	current      int
	from         string
	sort         string
}

// Next returns the next slice of messages
func (it *MessageIterator) Next() (*Message, error) {
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
func (convo *Conversation) MessagesFrom(ctx context.Context, from string) ([]*Message, error) {
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

	var messages []*Message
	err = json.NewDecoder(res.Body).Decode(messages)
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