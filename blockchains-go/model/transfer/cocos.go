package transfer

import "github.com/shopspring/decimal"

// cocos订单请求
type CocosOrderRequest struct {
	ApplyId      int64           `json:"applyid"`      //商户ID
	OuterOrderNo string          `json:"outerorderno"` //外部订单号
	OrderNo      string          `json:"orderno"`      //内部订单号
	MchName      string          `json:"mchname"`      //商户名称
	CoinName     string          `json:"coinname"`     //币种名称
	FromAddress  string          `json:"fromaddress"`  //发送地址
	ToAddress    string          `json:"toaddress"`    //接收地址
	ToAmount     decimal.Decimal `json:"toamount"`     //接收金额
	Memo         string          `json:"memo"`         //memo
	// write by flynn 2020-10-14
	AssetSymbol  string `json:"asset_symbol"`
	AssetId      string `json:"asset_id"` // asset_id
	AssetDecimal int32  `json:"asset_decimal"`
}
