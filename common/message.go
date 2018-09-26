package common

import (
	"encoding/json"
	"time"
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

	// Channel is the channel URL
	Channel string `json:"-"`
}

type MessagePart struct {
	// The message part ID
	ID string `json:"id,omitempty"`

	// The message part URL
	URL string `json:"url,omitempty"`

	// The message text
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
