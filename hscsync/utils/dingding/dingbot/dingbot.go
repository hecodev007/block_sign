package dingbot

import (
	"github.com/dghubble/sling"
)

const (
	DINGTALK_API_BASE_URL  = "https://oapi.dingtalk.com"
	DINGTALK_API_BASE_PATH = "/robot/send"
)

type baseMessage struct {
	MsgType string `json:"msgtype"`
}

type ResponseError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type defaultParams struct {
	AccessToken string `url:"access_token"`
}

func (msg baseMessage) GetClient(accessToken string) *sling.Sling {
	return sling.New().
		Post(DINGTALK_API_BASE_URL).
		Path(DINGTALK_API_BASE_PATH).
		QueryStruct(&defaultParams{AccessToken: accessToken}).
		BodyJSON(msg)
}
