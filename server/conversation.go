package server

import (
	"net/url"
	"net/http"
	"fmt"
	"encoding/json"
	"context"
	"bytes"
	"strings"
	"errors"

	"github.com/layerhq/go-client/common"
)

type Conversation struct {
	Id string `json:"id,omitempty"`
	URL string `json:"url,omitempty"`
	Participants []BasicIdentity `json:"participants"`
	CreatedAt string `json:"created_at"`
	Distinct bool `json:"distinct"`
	Metadata common.Metadata `json:"metadata"`

	// this feels so wrong
	Client *Server
}

const ConversationIDPrefix = "layer:///conversations/"

func (c *Conversation) ID() string {
	return common.UUIDFromLayerURL(c.Id)
}

func (c *Conversation) LayerID() string {
	return common.LayerID(common.ConversationsName, c.Id)
}

func (s *Server) buildConversationURL(id string) (u *url.URL, err error) {
	u, err = url.Parse(strings.TrimSuffix("conversations/" + id, "/"))
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
		return nil, fmt.Errorf("error building conversation URL: %v", err)
	}

	for p := range participants {
		participants[p] = common.LayerID(common.ConversationsName, participants[p])
	}
	reqBody := map[string]interface{}{
		"participants": participants,
		"distinct": distinct,
		"metadata": metadata,
	}

	obj, err := json.Marshal(reqBody)
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(obj))
	if err != nil {
		return nil, fmt.Errorf("error creating conversation request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status code is %d", res.StatusCode)
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
		return nil, fmt.Errorf("error building conversation URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("error creating delete conversations request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error deleting conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code is %d", res.StatusCode)
	}

	c := &Conversation{}
	err = json.NewDecoder(res.Body).Decode(&c)
	c.Client = s
	return c, err
}

func (c *Conversation) Delete(ctx context.Context) error {
	if c.Client == nil {
		return errors.New("apiClient not set in conversation")
	}
	// Create the request URL
	u, err := c.Client.buildConversationURL(c.ID())
	if err != nil {
		return fmt.Errorf("error building conversation URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodDelete, u.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("error creating delete conversations request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code is %d", res.StatusCode)
	}
	return nil
}

type EditOperation string
const (
	Add EditOperation = "add"
	Remove EditOperation = "remove"
	Set EditOperation = "set"
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
	Property EditProperty `json:"property"`
	Value interface{} `json:"value,omitempty"`
	//ParticipantsValue []string `json:"value,omitempty"`
	//MetadataValue common.Metadata `json:"value,omitempty"`

	// ID must be Layer ID (prefixed with layer:///identitities
	ID string `json:"id,omitempty"`
}

func (c *Conversation) UpdateParticipants(ctx context.Context, edits []ConversationEdit) error {
	if c.Client == nil {
		return errors.New("apiClient not set in conversation")
	}
	// Create the request URL
	u, err := c.Client.buildConversationURL(c.ID())
	if err != nil {
		return fmt.Errorf("error building conversation URL: %v", err)
	}

	obj, err := json.Marshal(edits)
	req, err := http.NewRequest(http.MethodPatch, u.String(), bytes.NewBuffer(obj))
	if err != nil {
		return fmt.Errorf("error creating update participants request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code is %d", res.StatusCode)
	}
	return nil
}

func (c *Conversation) UpdateMetadata(ctx context.Context, edits []ConversationEdit) error {
	if c.Client == nil {
		return errors.New("apiClient not set in conversation")
	}

	// Create the request URL
	u, err := c.Client.buildConversationURL(c.ID())
	if err != nil {
		return fmt.Errorf("error building conversation URL: %v", err)
	}

	obj, err := json.Marshal(edits)
	req, err := http.NewRequest(http.MethodPatch, u.String(), bytes.NewBuffer(obj))
	if err != nil {
		return fmt.Errorf("error creating update participants request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return fmt.Errorf("error creating conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code is %d", res.StatusCode)
	}
	return nil
}

func (c *Conversation) MarkRead(ctx context.Context, userID string, msgIndex *uint32) (uint32, error) {
	if c.Client == nil {
		return 0, errors.New("apiClient not set in conversation")
	}

	u, err := url.Parse(strings.TrimSuffix("users/" + userID + "/conversations/" + c.ID(), "/"))
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
		return 0, fmt.Errorf("error creating delete conversations request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := c.Client.transport.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error deleting conversation: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return 0, fmt.Errorf("status code is %d", res.StatusCode)
	}

	resp := map[string]uint32{}
	err = json.NewDecoder(res.Body).Decode(&c)
	return resp["position"], err
}
