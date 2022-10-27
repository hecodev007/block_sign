package vet

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
)

const WEI = 18
const VTHOContract = "0x0000000000000000000000000000456e65726779"
const VTHOInit = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

type Block struct {
	Height     int64    `json:"number"`
	Hash       string   `json:"id"`
	ParentHash string   `json:"parentID"`
	Timestamp  int64    `json:"timestamp"`
	IsTrunk    bool     `json:"isTrunk"`
	Txs        []string `json:"transactions"`
}

// Clause for json marshal
type Clause struct {
	To    string `json:"to"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

type TxMeta struct {
	BlockID        string `json:"blockID"`
	BlockNumber    uint32 `json:"blockNumber"`
	BlockTimestamp uint64 `json:"blockTimestamp"`
}

type ReceiptMeta struct {
	BlockID        string `json:"blockID"`
	BlockNumber    uint32 `json:"blockNumber"`
	BlockTimestamp uint64 `json:"blockTimestamp"`
	TxID           string `json:"txID"`
	TxOrigin       string `json:"txOrigin"`
}

// Output output of clause execution.
type Output struct {
	ContractAddress string      `json:"contractAddress"`
	Events          []*Event    `json:"events"`
	Transfers       []*Transfer `json:"transfers"`
}

// Event event.
type Event struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

// Transfer transfer log.
type Transfer struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
}

//Transaction transaction
type Transaction struct {
	Txid         string    `json:"id"`
	ChainTag     int32     `json:"chainTag"`
	BlockRef     string    `json:"blockRef"`
	Expiration   int64     `json:"expiration"`
	Clauses      []*Clause `json:"clauses"`
	GasPriceCoef uint8     `json:"gasPriceCoef"`
	Gas          uint64    `json:"gas"`
	Origin       string    `json:"origin"`
	Delegator    string    `json:"delegator"`
	Nonce        string    `json:"nonce"`
	DependsOn    string    `json:"dependsOn"`
	Size         uint32    `json:"size"`
	Meta         TxMeta    `json:"meta"`
}

//Receipt for json marshal
type TransactionReceipt struct {
	GasUsed  uint64      `json:"gasUsed"`
	GasPayer string      `json:"gasPayer"`
	Paid     string      `json:"paid"`
	Reward   string      `json:"reward"`
	Reverted bool        `json:"reverted"`
	Meta     ReceiptMeta `json:"meta"`
	Outputs  []*Output   `json:"outputs"`
}

func (e *Event) GetSender() (string, error) {
	if len(e.Topics[1]) != 66 {
		return "", fmt.Errorf("%s len is %d", e.Topics[1], len(e.Topics[1]))
	}
	addr := e.Topics[1][26:]
	return fmt.Sprintf("0x%s", addr), nil
}

func (e *Event) GetRecipient() (string, error) {
	if len(e.Topics[2]) != 66 {
		return "", fmt.Errorf("%s len is %d", e.Topics[2], len(e.Topics[2]))
	}
	addr := e.Topics[2][26:]
	return fmt.Sprintf("0x%s", addr), nil
}

func (e *Event) GetAmount() (decimal.Decimal, error) {
	if e.Data == "" {
		return decimal.Zero, fmt.Errorf("event data is empty")
	}
	amount := new(big.Int)
	amount.SetString(e.Data, 0)
	if amount.Sign() < 0 {
		return decimal.Zero, fmt.Errorf("get amount err")
	}
	//log.Printf("%s get amount %v",e.Data,amount)
	return decimal.NewFromBigInt(amount, 0), nil
}

func (e *Event) Valided() bool {
	if e.Address != VTHOContract {
		return false
	}
	if e.Topics[0] != VTHOInit {
		return false
	}
	return true
}

func (e *Transfer) GetSender() string {
	return e.Sender
}

func (e *Transfer) GetRecipient() string {
	return e.Recipient
}

func (e *Transfer) GetAmount() (decimal.Decimal, error) {
	amount := new(big.Int)
	amount.SetString(e.Amount, 0)
	if amount.Sign() < 0 {
		return decimal.Zero, fmt.Errorf("get amount err")
	}
	//i, err := strconv.ParseUint(e.Amount, 16, 64)
	//if err != nil {
	//	return decimal.Zero, err
	//}
	//log.Printf("%s get amount %v",e.Amount,amount)
	return decimal.NewFromBigInt(amount, 0), nil
}

func GetVTHOFeeFromStr(paid string) (decimal.Decimal, error) {
	fee := new(big.Int)
	fee.SetString(paid, 0)
	if fee.Sign() < 0 {
		return decimal.Zero, fmt.Errorf("get amount err")
	}
	return decimal.NewFromBigInt(fee, -18), nil
}
