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
func (c *Client) Identity(id string) (*Identity, error) {
	return nil, nil
}

// Followed returns a list of Identity objects that represent users following the client user
func (c *Client) Followed(ctx context.Context) ([]*Identity, error) {
	return nil, nil
}

// FollowedUsers returns a list of user ID strings that represent users following the client user
func (c *Client) FollowedUsers(ctx context.Context) ([]string, error) {
	return nil, nil
}

// IsFollowed returns true if the specified user Layer user ID is followed
func (c *Client) IsFollowed(ctx context.Context, id string) (bool, error) {
	return false, nil
}

// Follow follows the provided Layer user IDs
func (c *Client) Follow(ctx context.Context, ids []string) error {
	return nil
}
