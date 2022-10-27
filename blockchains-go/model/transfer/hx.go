package transfer

type HxOrderRequest struct {
	OrderRequestHead
	FromAddress  string           `json:"from_address"`  //发送地址
	ToAddress    string           `json:"to_address"`    //接收地址
	Amount       int64            `json:"amount"`        //金额，数量为 amount*10的八位
	IsRetry      bool             `json:"is_retry"`      //是否重试，默认false，不是必填
	OrderAddress []HxOrderAddress `json:"order_address"` //订单数组
	Memo         string           `json:"memo"`
}
type HxOrderAddress struct {
	Dir     int    `json:"dir"`
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}
