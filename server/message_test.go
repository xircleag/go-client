package server

import (
	"testing"
	"context"
)

func createTextMessage(c *Server) (*Message, error) {
	c, err := createTestClient()
	if err != nil {
		return nil, err
	}

	convo, err := createConversation(c)
	if err != nil {
		return nil, err
	}

	return convo.SendTextMessage(context.Background(), "test", "Hello there, how are you?", &MessageNotification{})
}
