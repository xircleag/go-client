package client

import (
	"encoding/json"
	"time"
	//"golang.org/x/net/context"
	//"github.com/layerhq/go-client/common"
)

const (
	MessageRecipientStatusSent      = "sent"
	MessageRecipientStatusDelivered = "delivered"
	MessageRecipientStatusRead      = "read"
)

type Message struct {
	ID              string            `json:"id,omitempty"`
	URL             string            `json:"url,omitempty"`
	ReceiptsURL     string            `json:"receipts_url,omitempty"`
	Position        json.Number       `json:"-"`
	Conversation    *Conversation     `json:"conversation,omitempty"`
	Parts           []*MessagePart    `json:"parts,omitempty"`
	SentAt          time.Time         `json:"-"`
	Sender          *MessageSender    `json:"sender,omitempty"`
	Unread          bool              `json:"is_unread,omitempty"`
	RecipientStatus map[string]string `json:"recipient_status,omitempty"`
}

type MessageSender struct {
	UserID string `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
}

type MessagePart struct {
	Body     string              `json:"body"`
	MimeType string              `json:"mime_type"`
	Encoding string              `json:"encoding,omitempty"`
	Content  *MessagePartContent `json:"content,omitempty"`
}

type MessagePartContent struct {
	ID         string
	MimeType   string `json:"mime_type"`
	Expiration time.Time
	RefreshURL string `json:"refresh_url,omitempty"`
	Size       json.Number
}
