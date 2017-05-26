// Package common contains various helper functions and definitions.
package common

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateUUID confirms a given ID is in a valid UUID format
func ValidateUUID(id string) error {
	format := "^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$"
	re := regexp.MustCompile(format)
	if !re.MatchString(id) {
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
