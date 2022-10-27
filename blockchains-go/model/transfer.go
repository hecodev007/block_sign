package model

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
)

type TransferParams struct {

	//{
	//	"sign":"9c7c569508fdfa2b4acace722d7a96714c",
	//	"sfrom":"test",
	//	"outOrderId":"eos00000000001",
	//	"coinName":"eos",
	//	"amount":"0.001",
	//	"toAddress":"gonglianyun1",
	//	"tokenName":"tpt",
	//	"contractAddress":"eosiotptoken",
	//	"memo":"test",
	//	"fee":"0.001"
	//	"isForce":true
	//}

	util.ApiSignParams
	Sfrom           string          `json:"sfrom" form:"sfrom"`
	CallBack        string          `json:"call_back" form:"call_back"`             //可选参数，覆盖配置
	OutOrderId      string          `json:"outOrderId" form:"outOrderId"`           //订单号，长度在11~64位之间
	CoinName        string          `json:"coinName" form:"coinName"`               //需要转出的币种
	Amount          decimal.Decimal `json:"amount" form:"amount"`                   //转出数量
	ToAddress       string          `json:"toAddress" form:"toAddress"`             //接收地址
	TokenName       string          `json:"tokenName" form:"tokenName"`             //代币名称，只有转出代币时，才需要填写
	ContractAddress string          `json:"contractAddress" form:"contractAddress"` //代币合约，只有转出代币时，才需要填写
	Memo            string          `json:"memo" form:"memo"`                       //对个别币种有效
	Fee             decimal.Decimal `json:"fee" form:"fee"`                         //指定手续费
	IsForce         bool            `json:"isForce" form:"isForce"`
	IsPseudCustody  bool            `json:"isCustody" form:"isCustody"`           // 是否交易所那边的商户
	BanFromAddress  string          `json:"banFromAddress" form:"banFromAddress"` // 不可使用此地址出账
	// IsForce 针对某些特殊币种（目前适用于CKB），强制性修正手续费,例如ckb找零有限制61金额找零
	//默认情况下 如果扣除完手续费之后如果找零小于61，这个订单将会被拒绝，而如果此字段为true，找零如果小于61金额则会强制性把找零金额附加在手续费之中
}

func (params *TransferParams) CheckParams() error {
	if params.OutOrderId == "" || len(params.OutOrderId) < 11 || len(params.OutOrderId) > 64 {
		return fmt.Errorf("outOrderId 传入参数值:%s，订单号，长度在11~64位之间", params.OutOrderId)
	}
	if params.CoinName == "" {
		return fmt.Errorf("coinName 传入参数值:%s", params.CoinName)
	}
	if params.Amount.IsZero() {
		return fmt.Errorf("amount 传入参数值:%s", params.Amount.String())
	}
	if params.ToAddress == "" {
		return fmt.Errorf("ToAddress 传入参数值:%s", params.ToAddress)
	}
	if params.ToAddress == "" {
		return fmt.Errorf("ToAddress 传入参数值:%s", params.ToAddress)
	}
	return nil
}
