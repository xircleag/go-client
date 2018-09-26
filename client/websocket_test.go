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

	c.Websocket.Connect()

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

	// Connect to the websocket
	err = c.Websocket.Connect()
	if err != nil {
		t.Fatalf("Error connecting to websocket: %v", err)
	}

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

func ExampleWebsocket() {
	// Register your desired handlers prior to connecting to make sure they
	// receive all events.
	c.Websocket.HandleFunc(Websocket.WebsocketMessageCreate, func(w *Websocket, p *WebsocketPacket) {
		switch p.Body.(type) {
		case *WebsocketResponse:
			res := p.Body.(WebsocketResponse)
			message := res.Data.(*common.Message)
			fmt.Println(fmt.Sprintf("%+v", message))
		}
	})

	ctx := context.Background()

	// Connect to the websocket
	c.Websocket.Connect()

	// Create a conversation
	convo, err := c.CreateConversation(ctx, []string{"recipient1", "recipient2"}, false, nil)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error creating conversation: %v", err))
		return
	}

	// Send a message to the conversation, resulting in a message creation event
	message, err := convo.SendTextMessage(ctx, "Example message", nil)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error sending message: %v", err))
		return
	}

	// Show the message contents
	fmt.Println(fmt.Sprintf("%+v", message))
}
