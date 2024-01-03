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
