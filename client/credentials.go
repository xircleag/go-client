package client

type ClientCredentials struct {
	// ApplicationID is UUID of the application
	ApplicationID string `json:"application_id"`

	// ProviderID is he UUID of the provider
	ProviderID string `json:"provider_id"`

	// AccountID is the UUID of the account
	AccountID string `json:"account_id"`

	// APIKey is te API key
	APIKey string `json:"api_key"`

	// Key is the PEM encoded key data
	Key *ClientCredentialsKey `json:"key"`
}

type ClientCredentialsKey struct {
	// ID is the key UUID
	ID string `json:"id"`

	// Private is the private key data in PEM format
	Private string `json:"private"`

	// Public is the public key data in PEM format
	Public string `json:"public"`
}
