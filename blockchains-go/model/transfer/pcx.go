package transfer

type PcxOrderRequest struct {
	ApplyId int64 `json:"applyid,omitempty"`

	OuterOrderNo string `json:"outerorderno,omitempty"`
	OrderNo      string `json:"orderno,omitempty"`
	MchName      string `json:"mchname,omitempty"`
	CoinName     string `json:"coinname,omitempty"`
	FromAddress  string `json:"fromAddress"`
	ToAddress    string `json:"toaddress"` // '接收者地址'
	Amount       string `json:"toamount"`
	Memo         string `json:"memo"`
	//MchId        int64  `json:"mch_id,omitempty"`
	//Worker       string `json:"worker,omitempty"` //指定机器运行
	//ApplyCoinId  int64  `json:"apply_coin_id,omitempty"`
}
