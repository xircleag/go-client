package server

import (
	"testing"

	"github.com/layerhq/go-client/common"

	"golang.org/x/net/context"
)

func TestDeleteIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	c.DeleteIdentity(ctx, "test")
	_, err = c.CreateIdentity(ctx, &common.Identity{
		UserID:      "test",
		DisplayName: "Test User",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.DeleteIdentity(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = c.CreateIdentity(ctx, &common.Identity{
		UserID:      "test",
		DisplayName: "Test User",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = c.Identity(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = c.UpdateIdentity(ctx, &common.Identity{
		UserID:      "test",
		DisplayName: "Test User Updated",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
}
