package server

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"
)

func TestDeleteIdentity(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	err = c.DeleteIdentity(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateIdentity(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = c.CreateIdentity(ctx, &Identity{
		UserID:      "test",
		DisplayName: "Test User",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetIdentity(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	i, err := c.Identity(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("Got identity: %+v", i))
}

func TestUpdateIdentity(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = c.UpdateIdentity(ctx, &Identity{
		UserID:      "test",
		DisplayName: "Test User Updated",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
}
