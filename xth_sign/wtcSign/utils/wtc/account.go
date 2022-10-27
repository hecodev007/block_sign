package wtc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"wtcSign/common/validator"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenAccount() (string, string, error) {
	privatekey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	privateBytes := crypto.FromECDSA(privatekey)
	addr := crypto.PubkeyToAddress(privatekey.PublicKey)
	return addr.String(), hex.EncodeToString(privateBytes), nil
}

func SignTx(params *validator.TelosSignParams, pri string) (txid string, rawtx string, err error) {
	privatekey, err := crypto.HexToECDSA(pri)
	if err != nil {
		return
	}
	signer := types.NewEIP155Signer(big.NewInt(15))
	var to common.Address
	var amount *big.Int
	var data []byte
	if params.Token != "" {
		to = common.HexToAddress(params.Token)
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

	tx := types.NewTransaction(params.Nonce, to, amount, params.GasLimit, params.GasPrice.BigInt(), data)
	tx, err = types.SignTx(tx, signer, privatekey)
	if err != nil {
		return "", "", err
	}
	var raw []byte
	buffer := bytes.NewBuffer(raw)
	err = tx.EncodeRLP(buffer)
	return tx.Hash().String(), "0x" + hex.EncodeToString(buffer.Bytes()), err
}
