package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/layerhq/go-client/common"
)

// AnnouncementCreate defines announcement creation parameters
type AnnouncementCreate struct {
	NotificationCreate
	SenderID string                `json:"sender_id"`
	Parts    []*common.MessagePart `json:"parts"`
}

//Announcement defines an announcement
type Announcement struct {
	ID         string                `json:"id"`
	URL        string                `json:"url"`
	SentAt     string                `json:"sent_at"`
	Recipients []string              `json:"recipients"`
	Sender     common.BasicIdentity  `json:"sender"`
	Parts      []*common.MessagePart `json:"parts"`
}

type NotificationCreate struct {
	Recipients   []string                    `json:"recipients"`
	Notification *common.MessageNotification `json:"notification,omitempty"`
}

func (s *Server) buildAnnouncementURL(id string) (u *url.URL, err error) {
	u, err = url.Parse(strings.TrimSuffix("announcements/"+id, "/"))
	if err != nil {
		return
	}
	u = s.baseURL.ResolveReference(u)
	return
}

func (s *Server) buildNotificationURL(id string) (u *url.URL, err error) {
	u, err = url.Parse(strings.TrimSuffix("notifications/"+id, "/"))
	if err != nil {
		return
	}
	u = s.baseURL.ResolveReference(u)
	return
}

// SendAnnouncement sends an announcement
func (s *Server) SendAnnouncement(ctx context.Context, sender string, recipients []string, parts []*common.MessagePart, notification *common.MessageNotification) (*Announcement, error) {
	// Create the request URL
	u, err := s.buildAnnouncementURL("")
	if err != nil {
		return nil, fmt.Errorf("Error building conversation URL: %v", err)
	}

	ac := &AnnouncementCreate{
		SenderID: sender,
		Parts:    parts,
		NotificationCreate: NotificationCreate{
			Recipients:   recipients,
			Notification: notification,
		},
	}

	for i := range ac.Recipients {
		ac.Recipients[i] = common.LayerURL(common.IdentitiesName, ac.Recipients[i])
	}

	// Create the request
	query, err := json.Marshal(ac)
	if err != nil {
		return nil, fmt.Errorf("Error creating announcement JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Status is %d", res.StatusCode)
	}

	var announcement *Announcement
	err = json.NewDecoder(res.Body).Decode(&announcement)
	return announcement, err
}

// SendTextMessage is a helper function to send a single-part plaintext message
func (s *Server) SendTextAnnouncement(ctx context.Context, sender string, recipients []string, message string, notification *common.MessageNotification) (*Announcement, error) {
	msg := plaintextMessage(message)
	return s.SendAnnouncement(ctx, sender, recipients, msg.Parts, notification)
}

func (s *Server) SendNotification(ctx context.Context, notification *NotificationCreate) error {
	for i := range notification.Recipients {
		notification.Recipients[i] = common.LayerURL(common.IdentitiesName, notification.Recipients[i])
	}

	// Create the request URL
	u, err := s.buildNotificationURL("")
	if err != nil {
		return fmt.Errorf("Error building conversation URL: %v", err)
	}

	// Create the request
	query, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("Error creating notification JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(query))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Status is %d", res.StatusCode)
	}

	return nil
}
