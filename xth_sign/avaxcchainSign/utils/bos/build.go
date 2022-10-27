package bos

import (
	"bytes"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strings"
	"avaxcchainSign/common/validator"
	"errors"
)

func SignTx(params *validator.SignParams, pri string) (txid string, rawtx string, err error) {
	privatekey, err := crypto.HexToECDSA(pri)
	if err != nil {
		return
	}
	signer := types.NewEIP155Signer(big.NewInt(66))
	var to common.Address
	var amount *big.Int
	var data []byte
	if params.ContractAddress != "" {
		to = common.HexToAddress(params.ContractAddress)
		amount = big.NewInt(0)
		datastr := "a9059cbb000000000000000000000000" + strings.TrimPrefix(params.ToAddress, "0x")
		valueByte := params.Value.BigInt().Bytes()
		valuehex := hex.EncodeToString(valueByte)
		valueparam := "0000000000000000000000000000000000000000000000000000000000000000"
		valueparam = valueparam[0:64-len(valuehex)] + valuehex
		datastr += valueparam
		if len(datastr) != 136 {
			return "", "", errors.New("合约转账data生成错误")
		}
		data, _ = hex.DecodeString(datastr)
	} else {
		to = common.HexToAddress(params.ToAddress)
		amount = params.Value.BigInt()
	}

	tx := types.NewTransaction(params.Nonce, to, amount, params.Gaslimit, params.Gasprice.BigInt(), data)
	tx, err = types.SignTx(tx, signer, privatekey)
	if err != nil {
		return "", "", err
	}
	var raw []byte
	buffer := bytes.NewBuffer(raw)
	err = tx.EncodeRLP(buffer)
	return tx.Hash().String(), "0x" + hex.EncodeToString(buffer.Bytes()), err
}
