package controllers

import (
	"bytes"
	"douyin-action-example/internal/actions/models"
	"encoding/json"
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
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	getUserInfoRequest := &models.GetUserInfoRequest{}
	accessToken := getBearerToken(c.Request)
	getUserInfoRequest.AccessToken = accessToken
	getUserInfoRequest.OpenID = getOpenID(accessToken)

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
				videoItem := &models.VideoItem{}
				videoItem.Title = video["title"].(string)

				statistics := video["statistics"].(map[string]interface{})
				videoItem.DiggCount = statistics["digg_count"].(float64)
				videoItem.DownloadCount = statistics["download_count"].(float64)
				videoItem.ShareCount = statistics["share_count"].(float64)
				videoItem.ForwardCount = statistics["forward_count"].(float64)
				videoItem.PlayCount = statistics["play_count"].(float64)

				getVideoListResponse.Videos = append(getVideoListResponse.Videos, videoItem)
			}
			c.JSON(http.StatusOK, getVideoListResponse)
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

func getOpenID(accessToken string) string {
	return "_000mWp_xZXvu3GStyvGh3QV4q3gM4dK8DPw"
}

func generateGetVideoListUrl(accessToken string, cursor int, count int) string {
	// 解析URL，并确保没有错误发生
	parsedURL, err := url.Parse(getVideoListUrl)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}

	// 创建一个URL值对象，用于添加query参数
	parameters := url.Values{}
	parameters.Add("open_id", getOpenID(accessToken))
	parameters.Add("cursor", strconv.Itoa(cursor))
	parameters.Add("count", strconv.Itoa(count))

	// 将query参数添加到URL中
	parsedURL.RawQuery = parameters.Encode()
	return parsedURL.String()
}
