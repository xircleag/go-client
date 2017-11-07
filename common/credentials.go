package common

type ClientCredentials struct {
	ApplicationID string `json:"application_id"`
	AccountID     string `json:"account_id"`
	ProviderID    string `json:"provider_id"`
	Key           *Key   `json:"key"`
	User          string `json:"-"`
	Token         string `json:"identity_token"`
}
