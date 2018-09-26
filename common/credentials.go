package common

type ClientCredentials struct {
	// ApplicationID is the UUID of the application
	ApplicationID string `json:"app_id"`

	// AccountID is the UUID of the account
	AccountID string `json:"-"`

	// ProviderID is the UUID of the identity provider
	ProviderID string `json:"-"`

	// Key is the encoded key data
	Key *Key `json:"-"`

	// User is the username or identifier
	User string `json:"-"`

	// Token is the identity token
	Token string `json:"identity_token"`
}
