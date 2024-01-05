package models

// GetTokenRequest 定义了符合OAuth标准的获取Token的请求格式
// OAuth标准定义详见: https://datatracker.ietf.org/doc/html/rfc6749#section-4.1.3
type GetTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	GrantType    string `json:"grant_type"`
}

// GetTokenResponse 定义了符合OAuth标准的获取Token的响应格式
type GetTokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpireIn     int    `json:"expires_in"`
	OpenID       string `json:"open_id"`
}

// ServiceError 定义了本服务的错误响应格式
type ServiceError struct {
	ErrorCode        float64 `json:"error_code"`
	ErrorDescription string  `json:"error_description"`
}

// DouYinGetTokenRequest 抖音定义的获取Token的请求格式
type DouYinGetTokenRequest struct {
	ClientKey    string `json:"client_key"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	GrantType    string `json:"grant_type"`
}
