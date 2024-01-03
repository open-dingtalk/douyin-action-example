package controllers

import (
	"douyin-action-example/internal/actions/models"
	"douyin-action-example/internal/conf"
	"encoding/json"
	"fmt"
	"github.com/chzealot/gobase/logger"
	"github.com/chzealot/gobase/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const getTokenUrl string = "https://open.douyin.com/oauth/access_token"

type AuthController struct {
	mu sync.Mutex
}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (h AuthController) Authorize(c *gin.Context) {
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
	douYinClient, err := NewDouYinClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// Declare a new Person struct.
	var tokenRequest models.GetTokenRequest

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(c.Request.Body).Decode(&tokenRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	fmt.Printf("tokenRequest: %+v\n", tokenRequest)

	getTokenResponse := &models.GetTokenResponse{}
	if err := douYinClient.Post(c, getTokenUrl, tokenRequest, getTokenResponse); err != nil {
		c.Error(err)
		return
	}

	resp := make(map[string]any)
	resp["access_token"] = getTokenResponse.AccessToken
	resp["token_type"] = "bearer"
	resp["refresh_token"] = getTokenResponse.RefreshToken
	resp["expires_in"] = getTokenResponse.ExpireIn
	c.JSON(http.StatusOK, resp)
}
