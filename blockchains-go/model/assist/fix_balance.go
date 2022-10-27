package assist

//{"mch":"HOOtest","coin":"eth","token_name":"dnat","contract":"0x6a39d7e18cbcd49e7d2de8ceef392c3757427607","address":"0x002f9d554cc781948eee90324d7da4a282b5fba4"}

type FixBalanceParams struct {
	Mch       string `json:"mch"`        //商户ID
	Coin      string `json:"coin"`       //主链
	TokenName string `json:"token_name"` //代币名称
	Contract  string `json:"contract"`   //代币合约
	Address   string `json:"address"`    //地址
}
