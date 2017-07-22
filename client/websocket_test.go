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
		err = c.Websocket.Receive(ctx, func(ctx context.Context, p *WebsocketPacket) {
			switch p.Body.(type) {
			case *WebsocketResponse:
				if p.Body.(*WebsocketResponse).RequestID == "1" {
					confirm <- true
				}
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
			Method:    WebsocketMethodCounterRead,
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

func TestWebsocketEventHandler(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	// Create a confirmation channel
	confirm := make(chan bool)

	// Setup event handlers
	c.Websocket.HandleFunc(WebsocketMethodCounterRead, func(w *Websocket, p *WebsocketPacket) {
		switch p.Body.(type) {
		case *WebsocketResponse:
			if p.Body.(*WebsocketResponse).RequestID == "1" {
				confirm <- true
			}
		}
	})

	// Start listening for events
	ctx := context.Background()
	go c.Websocket.Listen(ctx)

	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	c.Websocket.Send(ctx, &WebsocketPacket{
		Type: "request",
		Body: &WebsocketRequest{
			Method:    WebsocketMethodCounterRead,
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

func TestWebsocketMultipleEventHandler(t *testing.T) {
	c, err := createTestClient()
	if err != nil {
		t.Fatal(err)
	}

	// Create a confirmation channel
	confirm := make(chan bool, 2)

	// Setup event handlers
	c.Websocket.HandleFunc(WebsocketMethodCounterRead, func(w *Websocket, p *WebsocketPacket) {
		switch p.Body.(type) {
		case *WebsocketResponse:
			if p.Body.(*WebsocketResponse).RequestID == "1" {
				confirm <- true
			}
		}
	})

	c.Websocket.HandleFunc(WebsocketMethodCounterRead, func(w *Websocket, p *WebsocketPacket) {
		switch p.Body.(type) {
		case *WebsocketResponse:
			if p.Body.(*WebsocketResponse).RequestID == "1" {
				confirm <- true
			}
		}
	})

	// Start listening for events
	ctx := context.Background()
	go c.Websocket.Listen(ctx)

	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	c.Websocket.Send(ctx, &WebsocketPacket{
		Type: "request",
		Body: &WebsocketRequest{
			Method:    WebsocketMethodCounterRead,
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
