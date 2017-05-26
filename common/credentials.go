package common

type ClientCredentials struct {
	ApplicationID string `json:"app_id"`
	AccountID     string `json:"-"`
	ProviderID    string `json:"-"`
	Key           *Key   `json:"-"`
	User          string `json:"-"`
	Token         string `json:"identity_token"`
}
