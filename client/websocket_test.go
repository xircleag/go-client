package client

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestWebsocketReceive(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	// Create a confirmation channel
	confirm := make(chan bool)

	// Wait for messages
	go func() {
		ctx := context.Background()
		err = c.Websocket.Receive(ctx, func(ctx context.Context, packet *WebsocketResponse) {
			if packet.RequestID == "1" {
				confirm <- true
			}
		})

		if err != nil {
			t.Fatal(err)
		}
	}()

	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	c.Websocket.Send(ctx, &WebsocketPacket{
		Type: "request",
		Body: &WebsocketRequest{
			Method:    "Counter.read",
			RequestID: "1",
		},
	})

	select {
	case <-confirm:
		return
	case <-time.After(5 * time.Second):
		t.Fatal(fmt.Errorf("Timeout waiting for reply"))
	}
}
