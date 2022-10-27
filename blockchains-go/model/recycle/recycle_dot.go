package recycle

type RecycleDotReq struct {
	Coin    string `json:"coin"`
	Num     int    `json:"num"`
	AppId   int    `json:"appId"`
	Address string `json:"address"` //制定冷地址
}
