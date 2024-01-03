package models

/**
 * @Author simu.nn
 * @Date   2024/1/3 4:17 PM
 **/

type GetUserInfoRequest struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"open_id"`
}

type GetUserInfoResponse struct {
	AvatarUrl string `json:"avatarUrl"`
	Nick      string `json:"nick"`
	OpenID    string `json:"openId"`
	UnionID   string `json:"unionId"`
}

type GetVideoListResponse struct {
	Videos []*VideoItem `json:"videos"`
}

type VideoItem struct {
	// 视频标题
	Title string `json:"title"`
	// 点赞数
	DiggCount int64 `json:"diggCount"`
	// 下载数
	DownloadCount int64 `json:"downloadCount"`
	// 播放数，只有作者本人可见。公开视频设为私密后，播放数也会返回0
	PlayCount int64 `json:"playCount"`
	// 分享数
	ShareCount int64 `json:"shareCount"`
	// 转发数
	ForwardCount int64 `json:"forwardCount"`
	// 评论数
	CommentCount int64 `json:"commentCount"`
}
