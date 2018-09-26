package server

import (
	"context"
	"testing"

	"github.com/layerhq/go-client/common"
)

func TestCreateTextAnnouncement(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	_, err = c.CreateIdentity(ctx, &common.Identity{
		UserID:       "bot",
		DisplayName:  "Bot User",
		IdentityType: "bot",
	})

	_, err = c.SendTextAnnouncement(ctx, "layer:///identities/bot", []string{"test"}, "Test Announcement", nil)
	if err != nil {
		t.Error(err)
	}
}
