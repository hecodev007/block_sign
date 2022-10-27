package merge

type MergeParams struct {
	Froms       []string `json:"froms"`                 //来源地址
	To          string   `json:"to"`                    //接收地址
	Coin        string   `json:"coin"`                  //币种
	Token       string   `json:"token"`                 //token
	AmountFloat string   `json:"amountFloat,omitempty"` //金额选填，默认全部
	AppId       int      `json:"appId"`                 //商户ID
}

type BtcMergeParams struct {
	AppId int `json:"appId"` //商户ID
}

type OutCollectParams struct {
	Status int `json:"status"`
}
