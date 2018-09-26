package common

type BasicIdentity struct {
	// ID is the Layer URL of this user identity
	ID string `json:"id,omitempty"`

	// URL is the API URL of this user identity
	URL string `json:"url,omitempty"`

	// UserID is the user ID
	UserID string `json:"user_id,omitempty"`

	// DisplayName is the name to be displayed when referencing this identity
	DisplayName string `json:"display_name,omitempty"`

	// AvatarURL Is the URL of the user avatar
	AvatarURL string `json:"avatar_url,omitempty"`

	// IdentityType signifies what type of user this identity references
	IdentityType string `json:"identity_type,omitempty"`
}

type Identity struct {
	// ID is the Layer URL of this user identity
	ID string `json:"id,omitempty"`

	// URL is the API URL of this user identity
	URL string `json:"url,omitempty"`

	// UserID is the user ID
	UserID string `json:"user_id,omitempty"`

	// DisplayName is the name to be displayed when referencing this identity
	DisplayName string `json:"display_name,omitempty"`

	// AvatarURL Is the URL of the user avatar
	AvatarURL string `json:"avatar_url,omitempty"`

	// IdentityType signifies what type of user this identity references
	IdentityType string `json:"identity_type,omitempty"`

	// FirstName contains the first name of the user
	FirstName string `json:"first_name,omitempty"`

	// LastName contains the last name of the user
	LastName string `json:"last_name,omitempty"`

	// PhoneNumber contains the phone number of the user
	PhoneNumber string `json:"phone_number,omitempty"`

	// EmailAddress contains the email address of THE USER
	EmailAddress string `json:"email_address,omitempty"`

	// PublicKey contains the public key of the user, used for facilitating
	// end-to-end encryption and key exhcange
	PublicKey string `json:"public_key,omitempty"`

	// Metadata allows setting of arbitrary data on the identity
	Metadata map[string]string `json:"metadata,omitempty"`
}
