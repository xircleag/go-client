package client

type ClientCredentials struct {
	ApplicationID string                `json:"application_id"`
	ProviderID    string                `json:"provider_id"`
	AccountID     string                `json:"account_id"`
	APIKey        string                `json:"api_key"`
	Key           *ClientCredentialsKey `json:"key"`
}

type ClientCredentialsKey struct {
	ID      string `json:"id"`
	Private string `json:"private"`
	Public  string `json:"public"`
}
