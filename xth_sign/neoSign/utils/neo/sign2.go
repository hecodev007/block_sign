package neo

import (
	"encoding/hex"
	"errors"
	"neoSign/common/validator"
	"neoSign/common/log"
	"strings"

	neotransaction "github.com/x-contract/neo-go-sdk/neotransaction"
	"github.com/x-contract/neo-go-sdk/neoutils"
	//neoutil "github.com/o3labs/neo-utils/neoutils/smartcontract"
)

func GenAccount2() (address, private string, err error) {
	key := neotransaction.GenerateKeyPair()
	addr := key.CreateBasicAddress().Addr
	return addr, key.EncodeWif(), nil
}

func BuildTx2(params *validator.SignParams) (tx *neotransaction.NeoTransaction, err error) {
	tx = neotransaction.CreateContractTransaction()

	if params.Type == "claim" {
		tx = &neotransaction.NeoTransaction{
			Type: neotransaction.ClaimTransaction,
		}
		//tx.Type = neotransaction.ClaimTransaction

	}

	var inMount int64
	var inGasMount int64
	for _, in := range params.TxIns {
		from, err := neotransaction.ParseAddress(in.FromAddr)
		if err != nil {
			return nil, err
		}
		utxo := &neotransaction.UTXO{}
		utxo.TxHash, _ = hex.DecodeString(strings.TrimPrefix(in.FromTxid, "0x"))
		utxo.TxHash = neoutils.Reverse(utxo.TxHash)
		utxo.Index = uint16(in.FromIndex)
		assert := ""
		if in.Assert == "gas" {
			assert = neotransaction.AssetGasID
			inGasMount += in.FromAmountInt64
		} else {
			assert = neotransaction.AssetNeoID
			inMount += in.FromAmountInt64
		}
		utxo.AssetID, _ = hex.DecodeString(assert)
		utxo.AssetID = neoutils.Reverse(utxo.AssetID)
		utxo.Value = in.FromAmountInt64
		utxo.ScriptHash = from.ScripHash
		tx.AppendInput(utxo)
	}
	var outAmount int64
	var outGasAmount int64
	for _, out := range params.TxOuts {
		assert := ""
		if out.Assert == "gas"{
			assert = neotransaction.AssetGasID
			outGasAmount += out.ToAmountInt64
		} else {
			assert = neotransaction.AssetNeoID
			outAmount += out.ToAmountInt64
		}
		tx.AppendOutputByAddrString(out.ToAddr, assert, out.ToAmountInt64)
	}
	if inMount != outAmount {
		return tx, errors.New("neo 输入输出数量不相等")
	}
	if params.Type == "" && inGasMount < outGasAmount{
		return tx, errors.New("neo gas数量输出大于输入")
	}
	if params.Type == "" && inGasMount > outGasAmount+10000000 {
		return tx, errors.New("gas 手续费不能大于0.1")
	}
	return
}

func Sign2(tx *neotransaction.NeoTransaction, privates []string) (rawTx string, txid string, err error) {
	//SigedMap := make(map[string]bool)
	for _, private := range privates {
		//if _, ok := SigedMap[private]; ok {
		//	continue
		//}
		//SigedMap[private] = true

		key, err := neotransaction.DecodeFromWif(private)
		if err != nil {
			return "", "", err
		}
		log.Info(key.CreateBasicAddress().Addr)
		//log.Info(key.PrivateKey)
		tx.AppendBasicSignWitness(key)
	}
	return tx.RawTransactionString(), tx.TXID(), nil
}
