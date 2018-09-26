package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/layerhq/go-client/common"
)

// Conversation represents a Server API conversation
type Conversation struct {
	common.Conversation
	Client *Server `json:"-"`
}

const ConversationIDPrefix = "layer:///conversations/"

func (c *Conversation) UUID() string {
	return common.UUIDFromLayerURL(c.ID)
}

func (c *Conversation) LayerURL() string {
	return common.LayerURL(common.ConversationsName, c.ID)
}

func (s *Server) buildConversationURL(id string) (u *url.URL, err error) {
	u, err = url.Parse(strings.TrimSuffix("conversations/"+id, "/"))
	if err != nil {
		return
	}
	u = s.baseURL.ResolveReference(u)
	return
}

func (s *Server) CreateConversation(ctx context.Context, participants []string, distinct bool, metadata common.Metadata) (*Conversation, error) {
	// Create the request URL
	u, err := s.buildConversationURL("")
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}

	reqBody := map[string]interface{}{
		"participants": participants,
		"distinct":     distinct,
		"metadata":     metadata,
	}

	obj, err := json.Marshal(reqBody)
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(obj))
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	c := &Conversation{}
	err = json.NewDecoder(res.Body).Decode(&c)
	c.Client = s
	return c, err
}

func (s *Server) Conversation(ctx context.Context, id string) (*Conversation, error) {
	// Create the request URL
	u, err := s.buildConversationURL(id)
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("Error creating delete conversations request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error deleting conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	c := &Conversation{}
	err = json.NewDecoder(res.Body).Decode(&c)
	c.Client = s
	return c, err
}

func (c *Conversation) Delete(ctx context.Context) error {
	if c.Client == nil {
		return errors.New("Client not set in conversation")
	}
	// Create the request URL
	u, err := c.Client.buildConversationURL(c.UUID())
	if err != nil {
		return fmt.Errorf("Error building conversation URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodDelete, u.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("Error creating delete conversations request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("Error deleting conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Status code is %d", res.StatusCode)
	}
	return nil
}

type EditOperation string

const (
	Add    EditOperation = "add"
	Remove EditOperation = "remove"
	Set    EditOperation = "set"
)

type EditProperty string

const (
	Participants EditProperty = "participants"
)

func MetadataProperty(keypath ...string) EditProperty {
	return EditProperty(strings.Join(append([]string{"metadata"}, keypath...), "."))
}

type ConversationEdit struct {
	Operation EditOperation `json:"operation"`
	Property  EditProperty  `json:"property"`
	Value     interface{}   `json:"value,omitempty"`
	ID        string        `json:"id,omitempty"`
}

func (c *Conversation) UpdateParticipants(ctx context.Context, edits []ConversationEdit) error {
	if c.Client == nil {
		return errors.New("Client not set in conversation")
	}
	// Create the request URL
	u, err := c.Client.buildConversationURL(c.UUID())
	if err != nil {
		return fmt.Errorf("Error building conversation URL: %v", err)
	}

	obj, err := json.Marshal(edits)
	req, err := http.NewRequest(http.MethodPatch, u.String(), bytes.NewBuffer(obj))
	if err != nil {
		return fmt.Errorf("Error creating update participants request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("Error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Status code is %d", res.StatusCode)
	}
	return nil
}

func (c *Conversation) UpdateMetadata(ctx context.Context, edits []ConversationEdit) error {
	if c.Client == nil {
		return errors.New("Client not set in conversation")
	}

	// Create the request URL
	u, err := c.Client.buildConversationURL(c.UUID())
	if err != nil {
		return fmt.Errorf("Error building conversation URL: %v", err)
	}

	obj, err := json.Marshal(edits)
	req, err := http.NewRequest(http.MethodPatch, u.String(), bytes.NewBuffer(obj))
	if err != nil {
		return fmt.Errorf("Error creating update participants request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("Error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Status code is %d", res.StatusCode)
	}
	return nil
}

func (c *Conversation) MarkRead(ctx context.Context, userID string, msgIndex *uint32) (uint32, error) {
	if c.Client == nil {
		return 0, errors.New("Client not set in conversation")
	}

	u, err := url.Parse(strings.TrimSuffix("users/"+userID+"/conversations/"+c.UUID(), "/"))
	if err != nil {
		return 0, err
	}
	u = c.Client.baseURL.ResolveReference(u)

	reqBody := []byte{}
	if msgIndex != nil {
		reqBody, err = json.Marshal(map[string]uint32{"position": *msgIndex})
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, fmt.Errorf("Error creating delete conversations request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Error deleting conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return 0, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	resp := map[string]uint32{}
	err = json.NewDecoder(res.Body).Decode(&c)
	return resp["position"], err
}
