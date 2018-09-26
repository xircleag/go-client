// Package common contains various helper functions and definitions.
package common

import (
	"fmt"
	"regexp"
	"strings"
)

var uuidRE *regexp.Regexp = regexp.MustCompile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")

// ValidateUUID confirms a given ID is in a valid UUID format
func ValidateUUID(id string) error {
	if !uuidRE.MatchString(id) {
		return fmt.Errorf("Invalid UUID %s", id)
	}

	return nil
}

const (
	IdentitiesName    = "identities"
	ConversationsName = "conversations"
)

// LayerID creates a full Layer URL from a type and UUID
func LayerURL(typeName, id string) string {
	prefix := "layer:///" + typeName + "/"
	if strings.HasPrefix(id, prefix) {
		return id
	}
	return prefix + id
}

// UUIDFromLayerURL extracts the UUID portion of a Layer URL
func UUIDFromLayerURL(url string) string {
	if strings.HasPrefix(url, "layer") {
		parts := strings.Split(url, "/")
		url = parts[len(parts)-1]
	}

	return url
}
