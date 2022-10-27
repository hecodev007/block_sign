package model

import (
	"encoding/hex"
	"github.com/bsc-sign/conf"
	"github.com/bsc-sign/util"
	"strings"
)

type TransferParams struct {
	OuterOrderNo    string `json:"outer_order_no,omitempty"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	Token           string `json:"token"`
	ContractAddress string `json:"contract_address"`
	Mac             string `json:"mac"` // 消息认证码，保证消息来源和完整性

	// 对应finance数据库orderHot表的主键`id`
	// 因为该表会存在多条`outerOrderNo`相同的记录
	// 所以需要借助`id`来准确查询
	// 这里只会把该值存起来，等签名完毕再和`outerOrderNo`一起回传过去，不会依赖它做任何逻辑
	OrderHotId int `json:"order_hot_id"`

	// IsCollect       int    `json:"is_collect"`
	// Timestamp       uint64 `json:"timestamp"` // 时间戳，秒

}

type SignParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	ContractAddress string `json:"contract_address"`
	Nonce           int64  `json:"nonce"`
	GasPrice        int64  `json:"gas_price"`
	GasLimit        int64  `json:"gas_limit"`
}

// 转换为消息认证码
// 哈希算法使用SHA3-256
func (param *TransferParams) ToHMAC() string {
	var sb strings.Builder
	sb.WriteString("outer_order_no=")
	sb.WriteString(param.OuterOrderNo)
	sb.WriteString("&")

	sb.WriteString("from_address=")
	sb.WriteString(param.FromAddress)
	sb.WriteString("&")

	sb.WriteString("to_address=")
	sb.WriteString(param.ToAddress)
	sb.WriteString("&")

	sb.WriteString("amount=")
	sb.WriteString(param.Amount)
	sb.WriteString("&")

	sb.WriteString("token=")
	sb.WriteString(param.Token)
	sb.WriteString("&")

	sb.WriteString("contract_address=")
	sb.WriteString(param.ContractAddress)
	sb.WriteString("&")

	sb.WriteString("salt=")
	sb.WriteString(conf.Config.Secret.Salt)

	return hex.EncodeToString(util.Sha256([]byte(sb.String())))
}
