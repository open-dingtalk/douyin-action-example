package controllers

import (
	"bytes"
	"douyin-action-example/internal/actions/models"
	"douyin-action-example/internal/actions/storage"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"sync"
)

/**
 * @Author simu.nn
 * @Date   2024/1/3 5:47 PM
 **/

const getUserInfoUrl string = "https://open.douyin.com/oauth/userinfo/"

type BizController struct {
	mu         sync.Mutex
	httpClient *http.Client
}

func NewBizController() *BizController {
	return &BizController{}
}

func (bc *BizController) UserInfo(c *gin.Context) {
	dyClient, err := NewDouYinClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	getUserInfoRequest := &models.GetUserInfoRequest{}
	accessToken := getBearerToken(c.Request)
	getUserInfoRequest.AccessToken = accessToken
	getUserInfoRequest.OpenID, err = getOpenID(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	jsonData, err := json.Marshal(getUserInfoRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	httpRequest, err := http.NewRequest("POST", getUserInfoUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := dyClient.httpClient.Do(httpRequest)
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
			getUserInfoResponse := &models.GetUserInfoResponse{}
			getUserInfoResponse.AvatarUrl = respData["avatar"].(string)
			getUserInfoResponse.Nick = respData["nickname"].(string)
			getUserInfoResponse.OpenID = respData["open_id"].(string)
			getUserInfoResponse.UnionID = respData["union_id"].(string)
			c.JSON(http.StatusOK, getUserInfoResponse)
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

func (bc *BizController) GetVideoList(c *gin.Context) {

}

func (bc *BizController) GetVideoBase(c *gin.Context) {

}

func getBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Split the header value by space.
	// Should be in the form of ["Bearer", "token"]
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func getOpenID(accessToken string) (string, error) {
	return storage.OpenIdService.GetOpenIdByAccessToken(accessToken)
}
