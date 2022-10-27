package model

import "github.com/group-coldwallet/blockchains-go/pkg/util"

type NotifyOrderToMch struct {
	Sfrom             string   `json:"sfrom"`
	OutOrderId        string   `json:"outOrderId"`
	OuterOrderNo      string   `json:"outer_order_no"`
	CoinType          string   `json:"coin_type"`
	Chain             string   `json:"coin"` // V2新增
	Txid              string   `json:"txid"`
	TransactionId     string   `json:"transaction_id"`
	Msg               string   `json:"msg"`
	Memo              string   `json:"memo"`
	MemoEncrypt       string   `json:"memo_encrypt"`
	Contract          string   `json:"contract_address"`
	ContractAddress   string   `json:"contractAddress"`
	IsIn              IsInType `json:"is_in"` //1接收，2发送
	Amount            string   `json:"amount"`
	OrderSplitTxCount int      `json:"order_split_tx_count"` // V2新增 订单分隔交易的数量
	Txs               string   `json:"txs"`
	BlockHeight       int64    `json:"block_height"`
	Confirmations     int64    `json:"confirmations"`
	ConfirmTime       int64    `json:"confirm_time"`
	Fee               string   `json:"fee"`
	FromAddress       string   `json:"from_address"`
	ToAddress         string   `json:"to_address"`

	util.ApiSignParams

	// 兼容旧版本
	CoinName  string `json:"coin_name"`
	TokenName string `json:"tokenName"`

	// 新增全局ID的版本
	CoinUnionId     int `json:"coin_union_id"`
	CoinTypeUnionId int `json:"coin_type_union_id"`
}

type TxPushStatus int

const (
	TxPushNormal  = 1
	TxPushFailure = 2
)

type TxPushInner struct {
	SeqNo         string `json:"seq_no"`
	Status        int    `json:"status"` //交易状态；1：正常，2：失败交易
	BlockHeight   int64  `json:"block_height"`
	Amount        string `json:"amount"`
	ConfirmTime   int64  `json:"confirm_time"`
	Confirmations int64  `json:"confirmations"`
	Fee           string `json:"fee"`
	FromAddress   string `json:"from_address"`
	ToAddress     string `json:"to_address"`
	Memo          string `json:"memo"`
	Timestamp     int64  `json:"timestamp"`
	TransactionId string `json:"transaction_id"`
	TrxN          int    `json:"trx_n"`
}

//额外参数补充
type NotifyOtherParamsToMch struct {
	Txid    string `json:"txid"`
	Message string `json:"message"`
}

type IsInType int

const IsInType_Receive IsInType = 1
const IsInType_Send IsInType = 2
