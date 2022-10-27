package crust

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/JFJun/substrate-go/ss58"
	scale "github.com/itering/scale.go"
	"github.com/itering/scale.go/types"
)

func HexToTransaction(meta *types.MetadataStruct, rawTx string) (tx *Transaction, err error) {
	defer func() {
		if info := recover(); info != nil {
			tx = new(Transaction)
			err = fmt.Errorf("%v", info)
			//log.Info("panic", info)
		}
	}()
	if rawTx[0:2] == "0x" || rawTx[0:2] == "0X" {
		rawTx = rawTx[2:]
	}
	byteTx, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}
	e := scale.ExtrinsicDecoder{}
	option := types.ScaleDecoderOption{Metadata: meta}
	e.Init(types.ScaleBytes{Data: byteTx}, &option)
	e.Process()
	tx = new(Transaction)
	tx.Txid = "0x" + e.ExtrinsicHash
	for _, v := range e.Params {
		if v.Name == "dest" && v.Type == "Address" {
			tx.To, err = StringToAddress(v.Value.(string))
			if err != nil {
				return tx, err
			}
		}
		if v.Name == "value" && v.Type == "Compact<Balance>" {
			tx.Value, _ = v.Value.(decimal.Decimal)
			tx.Value = tx.Value.Shift(-12)
			//tx.Value = val
		}
	}
	value, ok := e.ScaleDecoder.Value.(map[string]interface{})
	if !ok {
		panic("")
	}
	account_id, ok := value["account_id"]
	if ok {
		tx.From, _ = StringToAddress(account_id.(string))
	}
	call_module_function, ok := value["call_module_function"]
	if ok {
		tx.Function = call_module_function.(string)
	}
	if tx.Function == "transfer" {
		tx.Fee = e.Tip
		tx.Fee = tx.Fee.Shift(-12)
	}
	return tx, err
}
func StringToAddress(pub string) (string, error) {
	bytePub, err := hex.DecodeString(pub)
	if err != nil {
		return "", err
	}
	return PubKeyToAddress(bytePub, []byte{42})
}
func PubKeyToAddress(pubKey, prefix []byte) (string, error) {
	return ss58.Encode(pubKey, prefix)
}
func HexToInt(s string) int64 {
	if len(s) > 2 && (s[0:2] == "0x" || s[0:2] == "0X") {
		s = s[2:]
	}
	i, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		panic(err.Error())
	}
	return i
}
