package transfer

//交易接口，返回orderId
type WalletServerRespOrder struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (ws *WalletServerRespOrder) Success() bool {
	return ws.Code == 0
}

//BTC计价返回
type WalletServerRespEstBtc struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *BtcOrderRequest `json:"data"`
}

//Ltc计价返回
type WalletServerRespEstLtc struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *BtcOrderRequest `json:"data"`
}
