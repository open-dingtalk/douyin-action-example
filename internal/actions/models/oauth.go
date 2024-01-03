package models

/**
 * @Author simu.nn
 * @Date   2024/1/3 4:17 PM
 **/

type GetTokenRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Code         string `json:"code"`
	GrantType    string `json:"grantType"`
}

type DouYinGetTokenRequest struct {
	ClientKey    string `json:"client_key"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	GrantType    string `json:"grant_type"`
}

type GetTokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpireIn     int    `json:"expires_in"`
	OpenID       string `json:"open_id"`
}

type DouYinError struct {
	ErrorCode        float64 `json:"error_code"`
	ErrorDescription string  `json:"error_description"`
}
