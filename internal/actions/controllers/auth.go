package controllers

import (
	"bytes"
	"context"
	"douyin-action-example/internal/actions/models"
	"douyin-action-example/internal/actions/storage"
	"douyin-action-example/internal/conf"
	"encoding/json"
	"fmt"
	"github.com/chzealot/gobase/logger"
	"github.com/chzealot/gobase/utils"
	"github.com/gin-gonic/gin"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const getTokenUrl string = "https://open.douyin.com/oauth/access_token/"

type AuthController struct {
	mu         sync.Mutex
	httpClient *http.Client
}

func NewAuthController() *AuthController {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "tcp", addr)
			},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 60 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}

	return &AuthController{
		httpClient: httpClient,
	}
}

func (h *AuthController) Authorize(c *gin.Context) {
	if conf.IsDebugMode {
		utils.DumpHttpRequest(c.Request)
	}
	// responseType := c.Query("response_type")
	clientId := c.Query("client_id")
	redirectUri := c.Query("redirect_uri")
	scope := c.Query("scope")
	scope = strings.ReplaceAll(scope, ",", " ")
	scope = strings.ReplaceAll(scope, "|", " ")
	scopes := strings.Split(scope, " ")
	dingtalkScope := strings.Join(scopes, " ")
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
	dingtalkUrl := fmt.Sprintf("https://login.dingtalk.com/oauth2/auth?redirect_uri=%s&response_type=code&client_id=%s&scope=%s&state=%s&prompt=%s",
		url.QueryEscape(thisRedirectUri), url.QueryEscape(oac.ClientID), url.QueryEscape(dingtalkScope), url.QueryEscape(stateStr), "consent")
	logger.Infof("redirect to %s", dingtalkUrl)
	c.Redirect(http.StatusFound, dingtalkUrl)
}

func (h *AuthController) Callback(c *gin.Context) {
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

func (h *AuthController) Token(c *gin.Context) {
	var getTokenRequest models.GetTokenRequest
	err := json.NewDecoder(c.Request.Body).Decode(&getTokenRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	douYinTokenRequest := &models.DouYinGetTokenRequest{
		ClientKey:    getTokenRequest.ClientID,
		ClientSecret: getTokenRequest.ClientSecret,
		Code:         getTokenRequest.Code,
		GrantType:    getTokenRequest.GrantType,
	}
	jsonData, err := json.Marshal(douYinTokenRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	httpRequest, err := http.NewRequest("POST", getTokenUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := h.httpClient.Do(httpRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer httpResponse.Body.Close()
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	douYinTokenResponse := make(map[string]interface{})
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, &douYinTokenResponse); err != nil {
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
			c.JSON(http.StatusOK, getTokenResponse)
		} else {
			getTokenError := &models.DouYinError{}
			getTokenError.ErrorCode = errorCode
			getTokenError.ErrorDescription = respData["description"].(string)
			c.JSON(http.StatusBadRequest, getTokenError)
		}

	} else {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
}
