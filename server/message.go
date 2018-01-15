package server

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
	"errors"
	"github.com/layerhq/go-client/iterator"
	"strconv"
	"google.golang.org/genproto/googleapis/storagetransfer/v1"
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

type Schedule struct {
	DelayInSeconds float64 `json:"delay_in_seconds"`
	TypingIndicatorInSeconds float64 `json:"typing_indicator_in_seconds"`
}

type MessageCreate struct {
	SenderID 	 string				   `json:"sender_id"`
	Parts        []*MessagePart       `json:"parts"`
	Notification *MessageNotification `json:"notification,omitempty"`
	Schedule     *Schedule            `json:"schedule,omitempty"`
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

// SendMessage sends a message to the server.  Note that if a schedule, is included then the returned Message
// pointer will be nil since the message will be created in the future.
func (convo *Conversation) SendMessage(ctx context.Context, mc *MessageCreate) (*Message, error) {
	if convo.apiClient == nil {
		return nil, errors.New("apiClient not set in conversation")
	}

	// Build the URL
	u, err := url.Parse(fmt.Sprintf("/conversations/%s/messages", convo.ID()))
	if err != nil {
		return nil, fmt.Errorf("error building conversation message URL: %v", err)
	}
	u = convo.apiClient.baseURL.ResolveReference(u)

	// Create the request
	query, err := json.Marshal(mc)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := convo.apiClient.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusConflict:
		return nil, fmt.Errorf("the requested message already exists")
	case res.StatusCode == http.StatusAccepted:
		return nil, nil
	case res.StatusCode != http.StatusCreated:
		return nil, fmt.Errorf("status code is %d", res.StatusCode)
	}

	var message *Message
	err = json.NewDecoder(res.Body).Decode(message)
	return message, err
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (convo *Conversation) SendTextMessage(ctx context.Context, senderID string, message string, notification *MessageNotification) (*Message, error) {
	msg := plaintextMessage(message)
	return convo.SendMessage(ctx, &MessageCreate{senderID, msg.Parts, notification, nil})
}

// SendMessage sends a message to the server.  Note that if a schedule, is included then the returned Message
// pointer will be nil since the message will be created in the future.
func (convo *Conversation) SendMessageBatch(ctx context.Context, mc []MessageCreate) error {
	if convo.apiClient == nil {
		return errors.New("apiClient not set in conversation")
	}
	if len(mc) > 12 {
		return errors.New("maximum 12 messages are supported")
	}

	for i := range mc {
		mc[i].SenderID = LayerID(IdentityType, mc[i].SenderID)
	}

	// Build the URL
	u, err := url.Parse(fmt.Sprintf("/conversations/%s/messages", convo.ID()))
	if err != nil {
		return fmt.Errorf("error building conversation message URL: %v", err)
	}
	u = convo.apiClient.baseURL.ResolveReference(u)

	// Create the request
	query, err := json.Marshal(mc)
	if err != nil {
		return fmt.Errorf("error creating conversation JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := convo.apiClient.transport.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status is %d", res.StatusCode)
	}

	return nil
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
func (convo *Conversation) MessagesFrom(ctx context.Context, from string, pageSize int) ([]*Message, error) {
	// Create the request URL
	convoID := common.UUIDFromLayerURL(convo.ID)
	u, err := url.Parse(fmt.Sprintf("/conversations/%s/messages", convoID))
	if err != nil {
		return nil, fmt.Errorf("error building conversation message URL: %v", err)
	}
	u = convo.apiClient.baseURL.ResolveReference(u)

	// Create the request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
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
	res, err := convo.apiClient.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("status code is %d", res.StatusCode)
	}

	var messages []*Message
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

func (convo *Conversation) DeleteMessage(ctx context.Context, msgID string) (error) {
	// Create the request URL
	u, err := url.Parse(fmt.Sprintf("/conversations/%s/messages/%s", common.UUIDFromLayerURL(convo.Id), common.UUIDFromLayerURL(msgID)))
	if err != nil {
		return fmt.Errorf("error building conversation message URL: %v", err)
	}
	u = convo.apiClient.baseURL.ResolveReference(u)

	// Create the request
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := convo.apiClient.transport.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code is %d", res.StatusCode)
	}

	return nil
}

type AnnouncementCreate struct {
	NotificationCreate
	SenderID     string               `json:"sender_id"`
	Parts        []*MessagePart       `json:"parts"`
}

type Announcement struct {
	ID string			 `json:"id"`
	URL string			 `json:"url"`
	SentAt string		 `json:"sent_at"`
	Recipients []string  `json:"recipients"`
	Sender BasicIdentity `json:"sender"`
	Parts []*MessagePart `json:"parts"`
}

type NotificationCreate struct {
	Recipients   []string    		  `json:"recipients"`
	Notification *MessageNotification `json:"notification,omitempty"`
}

func (s *Server) SendAnnouncement(ctx context.Context, ac *AnnouncementCreate) (*Announcement, error) {
	// Build the URL
	u, err := url.Parse("/announcements")
	if err != nil {
		return nil, fmt.Errorf("error building announcement URL: %v", err)
	}
	u = s.baseURL.ResolveReference(u)

	for i := range ac.Recipients {
		ac.Recipients[i] = LayerID(IdentityType, ac.Recipients[i])
	}

	// Create the request
	query, err := json.Marshal(ac)
	if err != nil {
		return nil, fmt.Errorf("error creating announcement JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status is %d", res.StatusCode)
	}

	var announcement *Announcement
	err = json.NewDecoder(res.Body).Decode(announcement)
	return announcement, err
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (s *Server) SendTextAnnouncement(ctx context.Context, senderID string, recipients []string, message string, notification *MessageNotification) (*Announcement, error) {
	msg := plaintextMessage(message)
	return s.SendAnnouncement(ctx, &AnnouncementCreate{NotificationCreate{recipients, notification}, senderID, msg.Parts})
}

func (s *Server) SendNotification(ctx context.Context, notification *NotificationCreate) error {
	for i := range notification.Recipients {
		notification.Recipients[i] = LayerID(IdentityType, notification.Recipients[i])
	}

	// Build the URL
	u, err := url.Parse("/notifications")
	if err != nil {
		return fmt.Errorf("error building notification URL: %v", err)
	}
	u = s.baseURL.ResolveReference(u)

	// Create the request
	query, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error creating notification JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status is %d", res.StatusCode)
	}

	return nil
}