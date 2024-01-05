package controllers

import (
	"bytes"
	"douyin-action-example/internal/actions/models"
	"douyin-action-example/internal/actions/storage"
	"douyin-action-example/internal/conf"
	"encoding/json"
	"fmt"
	"github.com/chzealot/gobase/logger"
	"github.com/chzealot/gobase/utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const getTokenUrl string = "https://open.douyin.com/oauth/access_token/"

type AuthController struct {
}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (ac *AuthController) Authorize(c *gin.Context) {
	if conf.IsDebugMode {
		utils.DumpHttpRequest(c.Request)
	}
	clientId := c.Query("client_id")
	redirectUri := c.Query("redirect_uri")
	scope := c.Query("scope")
	scope = strings.ReplaceAll(scope, ",", " ")
	scope = strings.ReplaceAll(scope, "|", " ")
	scopes := strings.Split(scope, " ")
	douYinScopes := strings.Join(scopes, ",")
	state := c.Query("state")

	oac := &models.OAuthCallback{
		State:       state,
		ClientID:    clientId,
		RedirectUri: redirectUri,
	}
	stateStr, err := oac.ToString()
	if err != nil {
		c.Error(err)
		return
	}
	thisRedirectUri := fmt.Sprintf("https://%s/auth/callback", c.Request.Host)
	douYinAuthUrl := fmt.Sprintf("https://open.douyin.com/platform/oauth/connect/?redirect_uri=%s&response_type=code&client_key=%s&scope=%s&state=%s&prompt=%s",
		thisRedirectUri, url.QueryEscape(oac.ClientID), url.QueryEscape(douYinScopes), url.QueryEscape(stateStr), "consent")
	logger.Infof("redirect to %s", douYinAuthUrl)
	c.Redirect(http.StatusFound, douYinAuthUrl)
}

func (ac *AuthController) Callback(c *gin.Context) {
	if conf.IsDebugMode {
		utils.DumpHttpRequest(c.Request)
	}
	code := c.Query("code")
	state := c.Query("state")
	oac, err := models.NewOAuthCallbackFromJson(state)
	if err != nil {
		c.Error(err)
		return
	}

	backUrl := fmt.Sprintf("%s?code=%s&state=%s",
		oac.RedirectUri,
		url.QueryEscape(code),
		url.QueryEscape(oac.State))
	logger.Infof("redirect to %s", backUrl)
	c.Redirect(http.StatusFound, backUrl)
}

func (ac *AuthController) Token(c *gin.Context) {
	getTokenRequest, err := ac.decodeGetTokenRequest(c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	douYinGetTokenRequest := ac.convertOAuth2DouYinGetTokenRequest(getTokenRequest)

	response, err := ac.sendGetTokenRequest(douYinGetTokenRequest, getTokenUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if response.StatusCode == http.StatusOK {
		douYinTokenResponse := make(map[string]interface{})
		if err := json.Unmarshal(responseBody, &douYinTokenResponse); err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		respData := douYinTokenResponse["data"].(map[string]interface{})
		errorCode := respData["error_code"].(float64)
		if errorCode == 0 {
			getTokenResponse := &models.GetTokenResponse{}
			getTokenResponse.TokenType = "bearer"
			getTokenResponse.AccessToken = respData["access_token"].(string)
			getTokenResponse.RefreshToken = respData["refresh_token"].(string)
			getTokenResponse.ExpireIn = int(respData["expires_in"].(float64))
			getTokenResponse.OpenID = respData["open_id"].(string)
			storage.OpenIdService.Save(getTokenResponse.AccessToken, getTokenResponse.OpenID)
			logger.Infof("get token succeed, response: %+v", getTokenResponse)
			c.JSON(http.StatusOK, getTokenResponse)
		} else {
			getTokenError := &models.ServiceError{}
			getTokenError.ErrorCode = errorCode
			getTokenError.ErrorDescription = respData["description"].(string)
			logger.Infof("get token returns error, response: %+v", getTokenError)
			c.JSON(http.StatusBadRequest, getTokenError)
		}

	} else {
		err = fmt.Errorf("httpResponse.StatusCode not ok, statusCode=%d", response.StatusCode)
		logger.Errorf("get token response not ok: %+v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
}

func (ac *AuthController) decodeGetTokenRequest(r *http.Request) (*models.GetTokenRequest, error) {
	defer r.Body.Close()

	var getTokenRequest models.GetTokenRequest
	err := json.NewDecoder(r.Body).Decode(&getTokenRequest)
	if err != nil {
		return nil, err
	}

	return &getTokenRequest, nil
}

// 把符合OAuth标准的获取Token请求，桥接为抖音的获取Token的请求
func (ac *AuthController) convertOAuth2DouYinGetTokenRequest(oauthTokenRequest *models.GetTokenRequest) *models.DouYinGetTokenRequest {
	return &models.DouYinGetTokenRequest{
		ClientKey:    oauthTokenRequest.ClientID,
		ClientSecret: oauthTokenRequest.ClientSecret,
		Code:         oauthTokenRequest.Code,
		GrantType:    oauthTokenRequest.GrantType,
	}
}

func (ac *AuthController) sendGetTokenRequest(request *models.DouYinGetTokenRequest, url string) (*http.Response, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpRequest, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	dyClient, err := NewDouYinClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create douyin client: %w", err)
	}

	resp, err := dyClient.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}
