package client

import (
	"golang.org/x/net/context"
)

type Identity struct {
	ID           string            `json:id,omitempty`
	URL          string            `json:url,omitempty`
	UserID       string            `json:user_id,omitempty`
	DisplayName  string            `json:display_name,omitempty`
	AvatarURL    string            `json:avatar_url,omitempty`
	FirstName    string            `json:first_name,omitempty`
	LastName     string            `json:last_name,omitempty`
	PhoneNumber  string            `json:phone_number,omitempty`
	EmailAddress string            `json:email_address,omitempty`
	IdentityType string            `json:identity_type,omitempty`
	PublicKey    string            `json:public_key,omitempty`
	Metadata     map[string]string `json:metadata,omitempty`
}

type BasicIdentity struct {
	ID          string `json:id,omitempty`
	URL         string `json:url,omitempty`
	UserID      string `json:user_id`
	DisplayName string `json:display_name,omitempty`
	AvatarURL   string `json:avatar_url,omitempty`
}

// GetIdentity fetches the identity with the given ID
func (c *Client) GetIdentity(id string) (*Identity, error) {
	return nil, nil
}

// CreateBasicIdentity creates a basic identity
func (c *Client) CreateBasicIdentity(ctx context.Context, i *BasicIdentity) (*BasicIdentity, error) {
	return nil, nil
}

// CreateIdentity creates a new identity
func (c *Client) CreateIdentity(ctx context.Context, i *Identity) (*Identity, error) {
	return nil, nil
}
