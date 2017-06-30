package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"golang.org/x/net/context"
)

type Identity struct {
	ID           string            `json:"id,omitempty"`
	URL          string            `json:"url,omitempty"`
	UserID       string            `json:"user_id,omitempty"`
	DisplayName  string            `json:"display_name,omitempty"`
	AvatarURL    string            `json:"avatar_url,omitempty"`
	FirstName    string            `json:"first_name,omitempty"`
	LastName     string            `json:"last_name,omitempty"`
	PhoneNumber  string            `json:"phone_number,omitempty"`
	EmailAddress string            `json:"email_address,omitempty"`
	IdentityType string            `json:"identity_type,omitempty"`
	PublicKey    string            `json:"public_key,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type BasicIdentity struct {
	ID          string `json:id`
	URL         string `json:url`
	UserID      string `json:user_id`
	DisplayName string `json:display_name,omitempty`
	AvatarURL   string `json:avatar_url,omitempty`
}

func (s *Server) buildIdentityURL(id string) (*url.URL, error) {
	var err error
	var u *url.URL
	if id != "" {
		u, err = url.Parse(fmt.Sprintf("users/%s/identity", id))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("No user ID specified")
	}
	u = s.baseURL.ResolveReference(u)
	return u, nil
}

func (s *Server) Identity(ctx context.Context, userID string) (*Identity, error) {
	// Create the request URL
	u, err := s.buildIdentityURL(userID)
	if err != nil {
		return nil, fmt.Errorf("Error building identity URL: %v", err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating identity request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error getting identity: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("No identity exists for the specified user")
	}

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	// Parse the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing conversation create response")
	}

	var identity *Identity
	if err := json.Unmarshal(body, &identity); err != nil {
		return nil, fmt.Errorf("Error parsing identity JSON: %v", err)
	}
	return identity, nil
}

func (s *Server) CreateIdentity(ctx context.Context, identity *Identity) (*Identity, error) {
	// Create the request URL
	u, err := s.buildIdentityURL(identity.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error building identity URL: %v", err)
	}

	// Create the request JSON
	query, err := json.Marshal(identity)
	if err != nil {
		return nil, fmt.Errorf("Error creating identity JSON: %v", err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("Error creating identity request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error creating identity: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("An identity for the specified user already exists")
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	return identity, nil
}

func (s *Server) DeleteIdentity(ctx context.Context, userID string) error {
	// Create the request URL
	u, err := s.buildIdentityURL(userID)
	if err != nil {
		return fmt.Errorf("Error building identity URL: %v", err)
	}

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return fmt.Errorf("Error creating identity request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return fmt.Errorf("Error deleting identity: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("No identity exists for the specified user")
	}

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return fmt.Errorf("Status code is %d", res.StatusCode)
	}

	return nil
}

func (s *Server) UpdateIdentity(ctx context.Context, identity *Identity, upsert bool) (*Identity, error) {
	if identity.UserID == "" {
		return nil, fmt.Errorf("UserID must be set on the Identity object")
	}

	// Get the existing identity
	_, err := s.Identity(ctx, identity.UserID)
	if err != nil {
		if !upsert {
			return nil, fmt.Errorf("Identity does not exist, cannot update")
		}

		_, err := s.CreateIdentity(ctx, identity)
		if err != nil {
			return nil, fmt.Errorf("Error creating identity")
		}
	}

	needUpdate := false

	// Parse identity structure into a map with tags
	updates := make(map[string]string)
	v := reflect.ValueOf(identity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Interface().(type) {
		case string:
			tag := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]
			value := v.Field(i).Interface().(string)
			if tag != "user_id" && value != "" {
				updates[tag] = value
				needUpdate = true
			}
		}
	}

	if !needUpdate && len(identity.Metadata) > 0 {
		return nil, fmt.Errorf("Nothing to update")
	}

	var data []*updateOperation
	if len(identity.Metadata) > 0 {
		metadataJSON, err := json.Marshal(identity.Metadata)
		if err == nil {
			data = append(data, &updateOperation{
				Operation: "set",
				Property:  "metadata",
				Value:     metadataJSON,
			})
		}
	}
	for key, value := range updates {
		data = append(data, &updateOperation{
			Operation: "set",
			Property:  key,
			Value:     value,
		})
	}

	// Create the request URL
	u, err := s.buildIdentityURL(identity.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error building identity URL: %v", err)
	}

	// Create the request JSON
	query, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Error creating update operation JSON: %v", err)
	}

	req, err := http.NewRequest("PATCH", u.String(), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("Error creating identity update request: %v", err)
	}
	req = req.WithContext(ctx)

	// Send the request
	res, err := s.transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error updating identity: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}

	updatedIdentity, err := s.Identity(ctx, identity.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error getting identity after set operations")
	}
	return updatedIdentity, nil
}
