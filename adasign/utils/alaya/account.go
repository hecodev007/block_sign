package alaya

import (
	"adasign/common/conf"
	"adasign/common/validator"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"math/big"

	"github.com/adiabat/bech32"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	//"github.com/PlatONnetwork/PlatON-Go/rlp"
	//"github.com/PlatONnetwork/PlatON-Go/common"
	//"github.com/PlatONnetwork/PlatON-Go/crypto"
	//"github.com/PlatONnetwork/PlatON-Go/core/types"
)

var Chainid int64 = 201018 //201030

func GenAccount() (address string, private string, err error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		panic(err.Error())
	}
	private = hex.EncodeToString(crypto.FromECDSA(privateKeyECDSA))
	addr := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	println(addr.String())
	address = bech32.Encode("atp", addr[:])
	return address, private, nil
}
func Bech32ToAddress(addr string) (address common.Address, err error) {
	_, rawaddr, err := bech32.Decode(addr)
	if err != nil {
		return address, err
	}
	address = common.BytesToAddress(rawaddr)
	return address, nil
}
func StringToPrivateKey(privateKeyStr string) (*ecdsa.PrivateKey, error) {
	privateKeyByte, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.ToECDSA(privateKeyByte)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
func BuildTx(params *validator.TelosSignParams) (tx *types.Transaction, err error) {
	to, err := Bech32ToAddress(params.ToAddress)
	if err != nil {
		return nil, err
	}

	tx = types.NewTransaction(uint64(params.Nonce.IntPart()), to, params.Amount.BigInt(), uint64(params.GasLimit.IntPart()), params.GasPrice.BigInt(), nil)
	return tx, nil
}
func SignTx(tx *types.Transaction, pri string) (txid string, rawTx string, err error) {
	private, err := StringToPrivateKey(pri)
	if err != nil {
		return "", "", err
	}
	signer := types.NewEIP155Signer(big.NewInt(Chainid))
	signedtx, err := types.SignTx(tx, signer, private)
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
func SendRawTransaction(rawTx string) (err error) {
	client := NewRpcClient(conf.GetConfig().Node.Url, "", "")

	if !strings.HasPrefix(rawTx, "0x") {
		rawTx = "0x" + rawTx
	}

	err = client.SendRawTransaction(rawTx)
	return err
}
