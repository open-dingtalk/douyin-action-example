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
	"strconv"
)

const getUserInfoUrl string = "https://open.douyin.com/oauth/userinfo/"
const getVideoListUrl string = "https://open.douyin.com/api/douyin/v1/video/video_list/"

type BizController struct {
}

func NewBizController() *BizController {
	return &BizController{}
}

func (bc *BizController) UserInfo(c *gin.Context) {
	if conf.IsDebugMode {
		utils.DumpHttpRequest(c.Request)
	}

	getUserInfoRequest, err := bc.buildGetUserInfoRequest(c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	response, err := bc.sendGetUserInfoRequest(getUserInfoRequest, getUserInfoUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	douYinTokenResponse := make(map[string]interface{})
	if response.StatusCode == http.StatusOK {
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
			logger.Infof("get user info succeed, response: %+v", getUserInfoResponse)
			c.JSON(http.StatusOK, getUserInfoResponse)
			return
		} else {
			getUserInfoError := &models.ServiceError{}
			getUserInfoError.ErrorCode = errorCode
			getUserInfoError.ErrorDescription = respData["description"].(string)
			logger.Infof("get user info returns error, response: %+v", getUserInfoError)
			c.JSON(http.StatusBadRequest, getUserInfoError)
			return
		}
	} else {
		err = fmt.Errorf("httpResponse.StatusCode not ok, statusCode=%d", response.StatusCode)
		logger.Errorf("get user info response not ok: %+v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
}

func (bc *BizController) GetVideoList(c *gin.Context) {
	getVideoListRequest, err := bc.buildGetVideoListRequest(c.Request, 0, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	getVideoListUrlWithParam, err := bc.generateGetVideoListUrl(getVideoListRequest, getVideoListUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	response, err := bc.sendGetVideoListRequest(getVideoListRequest, getVideoListUrlWithParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	douYinTokenResponse := make(map[string]interface{})
	if response.StatusCode == http.StatusOK {
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
					videoItem.ShareCount = int64(statistics["share_count"].(float64))
					videoItem.PlayCount = int64(statistics["play_count"].(float64))
					videoItem.CommentCount = int64(statistics["comment_count"].(float64))
					logger.Infof("videoItem=%+v", videoItem)
					getVideoListResponse.Videos = append(getVideoListResponse.Videos, videoItem)
				}
			}
			c.JSON(http.StatusOK, getVideoListResponse)
			return
		} else {
			getTokenError := &models.ServiceError{}
			getTokenError.ErrorCode = errorCode
			getTokenError.ErrorDescription = respExtra["description"].(string)
			c.JSON(http.StatusBadRequest, getTokenError)
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
}

func (bc *BizController) getOpenID(accessToken string) (string, error) {
	return storage.OpenIdService.GetOpenIdByAccessToken(accessToken)
}

func (bc *BizController) buildGetUserInfoRequest(r *http.Request) (*models.GetUserInfoRequest, error) {
	accessToken, err := GetBearerToken(r)
	if err != nil {
		return nil, err
	}

	openId, err := bc.getOpenID(accessToken)
	if err != nil {
		return nil, err
	}

	return &models.GetUserInfoRequest{
		AccessToken: accessToken,
		OpenID:      openId,
	}, nil
}

func (bc *BizController) sendGetUserInfoRequest(request *models.GetUserInfoRequest, url string) (*http.Response, error) {
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

func (bc *BizController) buildGetVideoListRequest(r *http.Request, cursor int, count int) (*models.GetVideoListRequest, error) {
	accessToken, err := GetBearerToken(r)
	if err != nil {
		return nil, err
	}

	openId, err := bc.getOpenID(accessToken)
	if err != nil {
		return nil, err
	}

	return &models.GetVideoListRequest{
		AccessToken: accessToken,
		OpenID:      openId,
		Cursor:      cursor,
		Count:       count,
	}, nil
}

func (bc *BizController) generateGetVideoListUrl(request *models.GetVideoListRequest, getVideoListUrl string) (string, error) {
	// Parse URL to make sure it is valid
	parsedURL, err := url.Parse(getVideoListUrl)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	// Create URL object to add query
	parameters := url.Values{}
	parameters.Add("open_id", request.OpenID)
	parameters.Add("cursor", strconv.Itoa(request.Cursor))
	parameters.Add("count", strconv.Itoa(request.Count))

	// Add query to url
	parsedURL.RawQuery = parameters.Encode()
	return parsedURL.String(), nil
}

func (bc *BizController) sendGetVideoListRequest(request *models.GetVideoListRequest, url string) (*http.Response, error) {
	httpRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("access-token", request.AccessToken)

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
