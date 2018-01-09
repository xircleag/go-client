// Package common contains various helper functions and definitions.
package common

import (
	"fmt"
	"regexp"
	"strings"
	"errors"
	"encoding/json"
)

var uuidRE *regexp.Regexp = regexp.MustCompile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")

// ValidateUUID confirms a given ID is in a valid UUID format
func ValidateUUID(id string) error {
	if !uuidRE.MatchString(id) {
		return fmt.Errorf("Invalid UUID %s", id)
	}

	return nil
}

func UUIDFromLayerURL(url string) string {
	if strings.HasPrefix(url, "layer") {
		parts := strings.Split(url, "/")
		url = parts[len(parts)-1]
	}

	return url
}

// Metadata enforces the types allowed as values in Layer metadata
type Metadata struct {
	data map[string]interface{}
}

// Set adds data to your Metadata so long as the value is a string or Metadata.
func (m *Metadata) Set(key string, value interface{}) error {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	_, isString := value.(string)
	_, isMetadata := value.(Metadata)
	if !isString && !isMetadata {
		return errors.New("value must be string or Metadata")
	}
	m.data[key] = value
	return nil
}

// Get retrieves the value in the map and returns nil if the key is not found.
func (m *Metadata) Get(key string) interface{} {
	if m.data == nil {
		return nil
	}
	return m.data[key]
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}

func (m *Metadata) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.data)
}
