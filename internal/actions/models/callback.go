package models

import "encoding/json"

type OAuthCallback struct {
	State       string `json:"state"`
	ClientID    string `json:"clientId"`
	RedirectUri string `json:"redirectUri"`
}

func NewOAuthCallbackFromJson(s string) (*OAuthCallback, error) {
	oac := &OAuthCallback{}
	if err := json.Unmarshal([]byte(s), oac); err != nil {
		return nil, err
	}
	return oac, nil
}

func (oac *OAuthCallback) ToString() (string, error) {
	b, err := json.Marshal(oac)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
