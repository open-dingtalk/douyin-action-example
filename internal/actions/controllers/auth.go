package controllers

import (
	"douyin-action-example/internal/actions/models"
	"douyin-action-example/internal/conf"
	"fmt"
	"github.com/chzealot/gobase/logger"
	"github.com/chzealot/gobase/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

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
	authCode := c.Query("authCode")
	state := c.Query("state")
	oac, err := models.NewOAuthCallbackFromJson(state)
	if err != nil {
		c.Error(err)
		return
	}

	backUrl := fmt.Sprintf("%s?code=%s&state=%s",
		oac.RedirectUri,
		url.QueryEscape(authCode),
		url.QueryEscape(oac.State))
	logger.Infof("redirect to %s", backUrl)
	c.Redirect(http.StatusFound, backUrl)
}

func (h *AuthController) Token(c *gin.Context) {
	//f := &TokenRequestForm{}
	//if err := c.ShouldBind(f); err != nil {
	//	c.Error(err)
	//	return
	//}
	//if len(f.ClientID) == 0 {
	//	auth := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
	//	if len(auth) != 2 || auth[0] != "Basic" {
	//		c.Error(errors.New("invalid Authorization"))
	//		return
	//	}
	//	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	//	pair := strings.SplitN(string(payload), ":", 2)
	//	if len(pair) != 2 {
	//		c.Error(errors.New("invalid Authorization"))
	//		return
	//	}
	//	f.ClientID = pair[0]
	//	f.ClientSecret = pair[1]
	//}
	//
	//if conf.IsDebugMode {
	//	utils.DumpHttpRequest(c.Request)
	//}
	//logger.Infof("token request, path=%s, grantType=%s", c.Request.URL.String(), f.GrantType)
	//
	//client, ok := h.GetDingTalkClient(f.ClientID)
	//if !ok {
	//	client = dingtalk.NewDingTalkClient(f.ClientID, f.ClientSecret)
	//	h.SaveDingTalkClient(f.ClientID, client)
	//}
	//
	//token, err := client.GetUserAccessToken(f.Code)
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	//resp := make(map[string]any)
	//resp["access_token"] = token.AccessToken
	//resp["token_type"] = "bearer"
	//resp["refresh_token"] = token.RefreshToken
	//resp["expires_in"] = token.ExpireIn
	//c.JSON(http.StatusOK, resp)
}
