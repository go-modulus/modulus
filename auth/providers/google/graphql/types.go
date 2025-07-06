package graphql

type RegisterViaGoogleInput struct {
	Code        string                 `json:"code"`
	Verifier    string                 `json:"verifier"`
	RedirectURL *string                `json:"redirectUrl"`
	UserInfo    map[string]interface{} `json:"userInfo"`
}
