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

func UUIDFromLayerURL(url string) string {
	if strings.HasPrefix(url, "layer") {
		parts := strings.Split(url, "/")
		url = parts[len(parts)-1]
	}

	return url
}
