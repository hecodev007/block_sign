package cph

import (
	"cphsign/common/validator"
	"cphsign/utils/sgb"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/ed25519"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

var Chainid int64 = 16162 //

func BuildTx(params *validator.TelosSignParams, prihex string) (txhash string, rawtx string, err error) {
	to := ToCommonAddress(params.ToAddress)
	tx := sgb.NewTransaction(uint64(params.Nonce.IntPart()), to, params.Value.BigInt(), uint64(params.GasLimit.IntPart()), params.GasPrice.BigInt(), []byte{}, 290)
	return SignTx(tx, prihex)
}
func SignTx(tx *sgb.Transaction, pri string) (txid string, rawTx string, err error) {
	private, err := StringToPrivateKey(pri)
	if err != nil {
		return "", "", err
	}
	signer := sgb.NewEIP155Signer(nil)
	signedtx, err := sgb.SignTxWithED25519(tx, signer, private, private.Public().(ed25519.PublicKey))
	if err != nil {
		return "", "", err
	}
	rawtx, err := rlp.EncodeToBytes(signedtx)
	if err != nil {
		return "", "", err
	}
	txhash := signedtx.Hash()
	return "0x" + hex.EncodeToString(txhash[:]), "0x" + hex.EncodeToString(rawtx), nil
}

func StringToPrivateKey(privateKeyStr string) (ed25519.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	return privateKey, nil
}

func ToCommonAddress(addr string) (address common.Address) {
	addr = strings.Replace(strings.ToLower(addr), "cph", "0x", 1)
	return common.HexToAddress(addr)
}

func ToCphAddress(addr string) (address string) {
	comaddr := ToCommonAddress(addr)
	return strings.Replace(comaddr.String(), "0x", "CPH", 1)
}
