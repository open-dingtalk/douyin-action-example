package controllers

import (
	"bytes"
	"douyin-action-example/internal/actions/models"
	"douyin-action-example/internal/actions/storage"
	"douyin-action-example/internal/conf"
	"encoding/json"
	"github.com/chzealot/gobase/logger"
	"github.com/chzealot/gobase/utils"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

/**
 * @Author simu.nn
 * @Date   2024/1/3 5:47 PM
 **/

const getUserInfoUrl string = "https://open.douyin.com/oauth/userinfo/"
const getVideoListUrl string = "https://open.douyin.com/api/douyin/v1/video/video_list/"

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
		logger.Errorf("NewDouYinClient failed, err=%s", err.Error())
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if conf.IsDebugMode {
		utils.DumpHttpRequest(c.Request)
	}

	getUserInfoRequest := &models.GetUserInfoRequest{}
	accessToken := getBearerToken(c.Request)
	getUserInfoRequest.AccessToken = accessToken
	getUserInfoRequest.OpenID, err = getOpenID(accessToken)
	if err != nil {
		logger.Errorf("getOpenID failed, err=%s, accessToken=%s", err.Error(), accessToken)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	jsonData, err := json.Marshal(getUserInfoRequest)
	if err != nil {
		logger.Errorf("json.Marshal(getUserInfoRequest) failed, err=%s", err.Error())
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	httpRequest, err := http.NewRequest("POST", getUserInfoUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("http.NewRequest failed, err=%s", err.Error())
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := dyClient.httpClient.Do(httpRequest)
	if err != nil {
		logger.Errorf("dyClient.httpClient.Do failed, err=%s", err.Error())
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer httpResponse.Body.Close()
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		logger.Errorf("io.ReadAll failed, err=%s", err.Error())
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	douYinTokenResponse := make(map[string]interface{})
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, &douYinTokenResponse); err != nil {
			logger.Errorf("json.Unmarshal failed, err=%s", err.Error())
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
			logger.Infof("get user info: %+v", getUserInfoResponse)
			c.JSON(http.StatusOK, getUserInfoResponse)
			return
		} else {
			getTokenError := &models.DouYinError{}
			getTokenError.ErrorCode = errorCode
			getTokenError.ErrorDescription = respData["description"].(string)
			c.JSON(http.StatusBadRequest, getTokenError)
		}
	} else {
		logger.Errorf("httpResponse.StatusCode not ok")
		c.JSON(http.StatusInternalServerError, err)
		return
	}
}

func (bc *BizController) GetVideoList(c *gin.Context) {
	dyClient, err := NewDouYinClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	accessToken := getBearerToken(c.Request)
	getVideoListUrlWithParam := generateGetVideoListUrl(accessToken, 0, 10)

	httpRequest, err := http.NewRequest("GET", getVideoListUrlWithParam, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	//httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("access-token", accessToken)

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
		respExtra := douYinTokenResponse["extra"].(map[string]interface{})
		errorCode := respExtra["error_code"].(float64)
		if errorCode == 0 {
			getVideoListResponse := &models.GetVideoListResponse{}
			videoList := respData["list"].([]interface{})
			for _, rawVideo := range videoList {
				// 类型断言，将interface{}转换为map[string]interface{}
				video, ok := rawVideo.(map[string]interface{})
				if !ok {
					c.JSON(http.StatusInternalServerError, err)
					return
				}

				title := video["title"].(string)
				if title != "" {
					videoItem := &models.VideoItem{}
					videoItem.Title = title
					statistics := video["statistics"].(map[string]interface{})
					videoItem.DiggCount = int64(statistics["digg_count"].(float64))
					videoItem.DownloadCount = int64(statistics["download_count"].(float64))
					videoItem.ShareCount = int64(statistics["share_count"].(float64))
					videoItem.ForwardCount = int64(statistics["forward_count"].(float64))
					videoItem.PlayCount = int64(statistics["play_count"].(float64))
					getVideoListResponse.Videos = append(getVideoListResponse.Videos, videoItem)
				}
			}

			debugInfo, _ := json.Marshal(getVideoListResponse)
			logger.Infof("status ok, response=%+v", debugInfo)
			c.JSON(http.StatusOK, getVideoListResponse)
			return
		} else {
			getTokenError := &models.DouYinError{}
			getTokenError.ErrorCode = errorCode
			getTokenError.ErrorDescription = respExtra["description"].(string)
			c.JSON(http.StatusBadRequest, getTokenError)
		}

	} else {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
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

func generateGetVideoListUrl(accessToken string, cursor int, count int) string {
	// 解析URL，并确保没有错误发生
	parsedURL, err := url.Parse(getVideoListUrl)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}

	// 创建一个URL值对象，用于添加query参数
	parameters := url.Values{}
	openId, _ := getOpenID(accessToken)
	parameters.Add("open_id", openId)
	parameters.Add("cursor", strconv.Itoa(cursor))
	parameters.Add("count", strconv.Itoa(count))

	// 将query参数添加到URL中
	parsedURL.RawQuery = parameters.Encode()
	return parsedURL.String()
}
