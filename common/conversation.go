package common

import (
	"encoding/json"
	"time"
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
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// LastMessage is a message object representing the last message sent in the
	// conversation.
	LastMessage *Message `json:"last_message,omitempty"`

	// Participants is an array of BasicIdentiy objects containing information on
	// the message participants.
	Participants []*BasicIdentity `json:"participants,omitempty"`

	// Distinct represents whether this is a distinct conversation with the
	// specified participant list.
	Distinct bool `json:"distinct"`

	// The number of unread messages on the conversation for the user specified
	// by the Client.
	UnreadMessageCount json.Number `json:"unread_message_count,omitempty"`

	// A generic interface available to store arbitrary metadata.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}
