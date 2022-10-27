package bo

type CreateAddrBO struct {
	OrderId    string `json:"orderId" binding:"required"`    //订单ID
	MchId      string `json:"mchId" binding:"required"`      //商家名
	CoinName   string `json:"coinName" binding:"required"`   //币种
	Num        uint   `json:"num" binding:"required,min=10"` //数量
	CreatePath string `json:"createPath"`                    //生成位置
}
