package common

import (
	"encoding/pem"
	"fmt"
)

// Key contains key data
type Key struct {
	// ID is the UUID of the key
	ID string `json:"key_id"`

	// KeyPair contains the key data
	KeyPair *KeyPair `json:"key_pair"`
}

// KeyPair contains a public and private key pair in PEM format strings
type KeyPair struct {
	// Public is the public key in PEM format
	Public string `json:"public"`

	// Private is the private key in PEM format
	Private string `json:"private"`
}

// Certificate contains Layer application credential data
type Certificate struct {
	*Key

	// Email is the email address associated with this key (not required)
	Email string `json:"email,omitempty"`

	// AccounntID is the account ID associated with this key
	AccountID string `json:"account_id"`

	// ProviderID is the Layer provider ID associated with this key
	ProviderID string `json:"provider_id"`

	// ApplicationID is the Layer application ID
	ApplicationID string `json:"application_id"`

	// APIKey is the Layer API key
	APIKey string `json:"api_key"`
}

func ValidateKey(key *Key) error {
	if err := ValidateUUID(key.ID); err != nil {
		return fmt.Errorf("Invalid key ID")
	}

	// Confirm private key
	if key.KeyPair.Private == "" {
		return fmt.Errorf("No private key data specified")
	}
	p, _ := pem.Decode([]byte(key.KeyPair.Private))
	if p == nil {
		return fmt.Errorf("Invalid PEM data in private key")
	}

	// Confirm public key if present (not required)
	if key.KeyPair.Public != "" {
		p, _ := pem.Decode([]byte(key.KeyPair.Public))
		if p == nil {
			return fmt.Errorf("Invalid PEM data in public key")
		}
	}

	return nil
}
