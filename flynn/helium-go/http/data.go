package http

import (
	"github.com/JFJun/helium-go/utils"
	"math/big"
)

var (
	PaymentV1 string = "payment_v1"
	PaymentV2 string = "payment_v2"
)

type RespBlock struct {
	TransactionCount int    `json:"transaction_count"`
	Time             int64  `json:"time"`
	PrevHash         string `json:"prev_hash"`
	Height           int64  `json:"height"`
	Hash             string `json:"hash"`
}

type RespTransaction struct {
	Data   []TransactionData `json:"data"`
	Cursor string            `json:"cursor"`
}

type TransactionData struct {
	Type string `json:"type"`
	Txid string `json:"hash"`
}

type RespPaymentTransaction struct {
	Type      string    `json:"type"`
	Time      int64     `json:"time"`
	Signature string    `json:"signature"`
	Payer     string    `json:"payer"`
	Payee     string    `json:"payee"`
	Nonce     int64     `json:"nonce"`
	Height    int64     `json:"height"`
	Txid      string    `json:"hash"`
	Fee       uint64    `json:"fee"`
	Amount    uint64    `json:"amount"`
	Payments  []Payment `json:"payments"`
}

func (rpt *RespPaymentTransaction) ParsePayment() map[string]float64 {
	maps := make(map[string]float64)
	if rpt.Type == PaymentV1 {
		maps[rpt.Payee] = utils.CoinToFloat(new(big.Int).SetUint64(rpt.Amount), int32(8))
		return maps
	}
	for _, payment := range rpt.Payments {
		maps[payment.Payee] = utils.MathAdd(maps[payment.Payee], utils.CoinToFloat(new(big.Int).SetUint64(payment.Amount),
			int32(8)), int32(8))
	}
	return maps
}

type Payment struct {
	Payee  string `json:"payee"`
	Amount uint64 `json:"amount"`
}

type RespPendingStatus struct {
	UpdatedAt    string `json:"updated_at"`
	Status       string `json:"status"`
	Hash         string `json:"hash"`
	FailedReason string `json:"failed_reason"`
	CreatedAt    string `json:"created_at"`
}

type RespBroadcastTx struct {
	Hash string `json:"hash"`
}

type RespAccount struct {
	SpeculativeNonce int    `json:"speculative_nonce"`
	SecNonce         int    `json:"sec_nonce"`
	SecBalance       uint64 `json:"sec_balance"`
	Nonce            int    `json:"nonce"`
	DcNonce          int    `json:"dc_nonce"`
	DcBalance        uint64 `json:"dc_balance"`
	Block            int64  `json:"block"`
	Balance          uint64 `json:"balance"`
	Address          string `json:"address"`
}

type RespVars struct {
	DcPayloadSize    int64 `json:"dc_payload_size"`
	TxnFeeMultiplier int64 `json:"txn_fee_multiplier"`
	//todo add another field
}
type RespCurrentPrices struct {
	Price int64 `json:"price"`
	Block int64 `json:"block"`
}
