package xrp

import (
	"time"
)

type LedgerCurrentResponse struct {
	XrpError
	LedgerCurrentIndex int64  `json:"ledger_current_index"`
	Status             string `json:"status"`
}

type XrpError struct {
	ErrorSt      string `json:"error"`
	ErrorCode    int64  `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func (e XrpError) Error() string {
	return e.ErrorMessage
}

type Transaction struct {
	XrpError
	Status             string      `json:"status"`
	Account            string      `json:"Account"`
	Amount             interface{} `json:"Amount"`
	Destination        string      `json:"Destination"`
	DestinationTag     int64       `json:"DestinationTag"`
	Sequence           int64       `json:"Sequence"`
	TakerGets          interface{} `json:"TakerGets"`
	Date               int64       `json:"date"`
	Meta               Meta        `json:"meta"`
	Fee                string      `json:"Fee"`
	SigningPubKey      string      `json:"SigningPubKey"`
	TransactionType    string      `json:"TransactionType"` //Payment
	InLedger           int64       `json:"inLedger"`
	LedgerIndex        int64       `json:"ledger_index"`
	LastLedgerSequence int64       `json:"LastLedgerSequence"`
	TakerPays          interface{} `json:"TakerPays"`
	TxnSignature       string      `json:"TxnSignature"`
	Validated          bool        `json:"validated"`
	Flags              int64       `json:"Flags"` //2147483648
	OfferSequence      int64       `json:"OfferSequence"`
	Hash               string      `json:"hash"`
	Memo               string      `json:"memo"`
}
type Meta struct {
	AffectedNodes     interface{} `json:"AffectedNodes"`
	TransactionIndex  float64     `json:"TransactionIndex"`
	TransactionResult string      `json:"TransactionResult"` //tesSUCCESS
	DeliveredAmount   interface{} `json:"delivered_amount"`
}
type TakerPays struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
	Issuer   string `json:"issuer"`
}

//type ModifiedNode struct {
//	FinalFields FinalFields `json:"FinalFields"`
//	LedgerEntryType string `json:"LedgerEntryType"`
//	LedgerIndex string `json:"LedgerIndex"`
//}
//type FinalFields struct {
//	TakerPaysIssuer string `json:"TakerPaysIssuer"`
//	ExchangeRate string `json:"ExchangeRate"`
//	Flags float64 `json:"Flags"`
//	RootIndex string `json:"RootIndex"`
//	TakerGetsCurrency string `json:"TakerGetsCurrency"`
//	TakerGetsIssuer string `json:"TakerGetsIssuer"`
//	TakerPaysCurrency string `json:"TakerPaysCurrency"`
//	OwnerCount float64 `json:"OwnerCount"`
//	Sequence float64 `json:"Sequence"`
//	Account string `json:"Account"`
//	Balance string `json:"Balance"`
//}
//
type GetBlockResult struct {
	Block *Block `json:"ledger"`
	XrpError
	BlockHeight int64  `json:"ledger_index"`
	Status      string `json:"status"`
}

type Block struct {
	Height         int64     `json:"height"`
	Hash           string    `json:"hash"`
	Time           time.Time `json:"time"`
	CloseTimeHuman string    `json:"close_time_human"`
	ParentHash     string    `json:"parent_hash"`
	Transactions   []string  `json:"transactions"`
}

type FullBlock struct {
	Block
	Transacitons []*Transaction
}

type BalanceResponse struct {
	XrpError
	AccountData        AccountData `json:"account_data"`
	LedgerCurrentIndex int64       `json:"ledger_current_index"`
	Status             string      `json:"status"`
}
type AccountData struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}
