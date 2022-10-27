package transfer

import "encoding/json"

type EthOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
	Token       string `json:"token,omitempty"`
	Fee         int64  `json:"fee,omitempty"`
}

type EthCollectReq struct {
	OrderRequestHead
	FromAddrs    []string `json:"from_addrs"`
	ToAddr       string   `json:"to_addr"`
	ContractAddr string   `json:"contract_addr,omitempty"`
	Decimal      int      `json:"decimal"`
	Amount       string   `json:"amount"` // 为了ETH内部转账新增的字段

}

type EthTransferFeeReq struct {
	OrderRequestHead
	FromAddr string   `json:"from_addr"`
	ToAddrs  []string `json:"to_addrs"`
	NeedFee  string   `json:"need_fee"`
}

type EthBalanceReq struct {
	Address      string `json:"address"`
	ContractAddr string `json:"contract_addr,omitempty"`
	Decimal      int    `json:"decimal,omitempty"`
}

/*
date: 2020-09-28
*/
// 钉钉打手续费
type EthTransferFee struct {
	To       string `json:"to"`
	MchId    int64  `json:"mchId"`
	FeeFloat string `json:"feeFloat"`
}

type EthListAmount struct {
	Coin  string `json:"coin"`  // 币种名字
	MchId int64  `json:"mchId"` // 商户Id
	Num   int    `json:"num"`   // 查看数量  	可选（=10）
}

type EthCollectToken struct {
	Coin  string   `json:"coin"`  // 币种名字
	MchId int64    `json:"mchId"` // 商户Id
	From  []string `json:"from"`
}

type EthInternal struct {
	From   string `json:"from"`
	To     string `json:"to"`
	MchId  int64  `json:"mchId"`
	Amount string `json:"amount"`
}

type Collect struct {
	Coin  string   `json:"coin"`  // 币种名字
	MchId int64    `json:"mchId"` // 商户Id
	From  []string `json:"from"`
}

type SameWayBack struct {
	TxId string `json:"txId"` // 交易id
}

type EthResetNonce struct {
	Address string `json:"addr"`
	Nonce   int    `json:"nonce"`
}

type XRPSupplemental struct {
	From   string `json:"from"`
	Amount string `json:"amount"`
}

type ETHOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
	//  token 转账 需要参数
	ContractAddress string `json:"contract_address"`
	Token           string `json:"token"`  // 代币的名字，主链转账不传这个值
	Latest          bool   `json:"latest"` // 使用latest获取nonce，默认为false
}

func DecodeETHTransferResp(data []byte) map[string]interface{} {
	var result map[string]interface{}
	if len(data) != 0 {
		err := json.Unmarshal(data, &result)
		if err == nil {
			return result
		} else {
			return nil
		}
	}
	return nil
}
