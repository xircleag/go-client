package client

import (
	"github.com/layerhq/go-client/common"

	"golang.org/x/net/context"
)

// Identity fetches the identity with the given ID
func (c *Client) Identity(id string) (*common.Identity, error) {
	return nil, nil
}

// Followed returns a list of Identity objects that represent users following the client user
func (c *Client) Followed(ctx context.Context) ([]*common.Identity, error) {
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
