package recycle

type RecycleParams struct {
	Coin        string `json:"coin"`                  //币种
	FeeFloat    string `json:"feeFloat,omitempty"`    //币种手续费指定
	AmountFloat string `json:"amountFloat,omitempty"` //金额选填，默认全部
	FromAddress string `json:"fromAddress,omitempty"` // 可选参数
	AppId       int    `json:"appId"`                 //商户ID
	Model       int    `json:"model"`                 //默认0 小金额 1 大金额
}

type BtcRecycleParams struct {
	AppId int `json:"appId"` //商户ID
}
