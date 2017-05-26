package common

import (
	"encoding/pem"
	"fmt"
)

type Key struct {
	ID      string   `json:"key_id"`
	KeyPair *KeyPair `json:"key_pair"`
}

type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

type Certificate struct {
	*Key
	Email      string `json:"email"`
	Password   string `json:"password"`
	AccountID  string `json:"account_id"`
	ProviderID string `json:"provider_id"`
	AppID      string `json:"app_id"`
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
