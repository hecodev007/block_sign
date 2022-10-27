package dag

import (
	"cfxSign/common/log"
	"cfxSign/common/validator"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

func GenAccount() (address string, private string, err error) {
	return GenAccount2()
}
func GenAccount1() (address string, private string, err error) {
	ac := sdk.NewAccountManager("keystore",cfxaddress.NetowrkTypeMainnetID)
	addr, err := ac.Create("")
	if err != nil {
		return "", "", err
	}
	private, err = ac.Export(addr, "")
	return addr.String(), private, err
}
func GenAccount2() (address string, private string, err error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	if len(privateKeyECDSA.D.Bytes()) != 32 {
		return GenAccount2()
	}
	addr := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	baseaddr,err :=cfxaddress.NewFromBytes(addr[:],cfxaddress.NetowrkTypeMainnetID)
	if err != nil {
		return "","",err
	}
	return baseaddr.String(), hex.EncodeToString(crypto.FromECDSA(privateKeyECDSA)), nil
}

func BuildTx(params *validator.TelosSignParams) (tx *types.UnsignedTransaction, err error) {
	tx = &types.UnsignedTransaction{}

	if fromAddress,err := cfxaddress.NewFromBase32(params.FromAddress);err != nil {
		return nil,err
	} else {
		tx.From = &fromAddress

	}
	toAddress,err := cfxaddress.NewFromBase32(params.ToAddress)
	if err != nil{
		return nil, err
	}

	tx.To = &toAddress

	//平台币转账
	if params.Token == ""{
		tx.Value = types.NewBigIntByRaw(params.Value.BigInt())
	} else {//合约转账
		if tokenaddress,err :=cfxaddress.NewFromBase32(params.Token);err != nil {
			return nil,err
		} else {
			tx.To = &tokenaddress
		}
		erctoken,err := abi.JSON(strings.NewReader(TokenABI))
		if err != nil {
			return nil ,err
		}
		data,err :=erctoken.Pack("transfer",toAddress.MustGetCommonAddress(),params.Value.BigInt())
		if err != nil {
			return nil ,err
		}
		tx.Data=data
		log.Info(hex.EncodeToString(tx.Data))
	}
	tx.Gas = types.NewBigIntByRaw(params.GasLimit.BigInt())
	tx.GasPrice = types.NewBigIntByRaw(params.GasPrice.BigInt())
	tx.Nonce = types.NewBigIntByRaw(params.Nonce.BigInt())
	tx.ChainID = types.NewUint(params.ChainID)
	tx.EpochHeight = types.NewUint64(params.EpochHeight)
	return tx, nil
}
func SignTx2(tx *types.UnsignedTransaction, private string) (txhash string, rawTx string, err error) {
	acc := sdk.NewAccountManager("dagkeydir",cfxaddress.NetowrkTypeMainnetID)
	addr, err := acc.ImportKey(private, "123456")
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	fmt.Println("ImportKey",addr.String())
	v, r, s, err := acc.Sign(*tx, "123456")
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	signedTx := &types.SignedTransaction{
		UnsignedTransaction: *tx,
		V:                   v,
		S:                   s,
		R:                   r,
	}
	rawtx, err := signedTx.Encode()
	if err != nil {
		return "", "", err
	}
	hash, err := tx.Hash()
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(hash), hex.EncodeToString(rawtx), nil
}
func SignTx(tx *types.UnsignedTransaction, private string) (txhash string, rawTx []byte, err error) {
	if utils.Has0xPrefix(private) {
		private = private[2:]
	}
	privateKey, err := crypto.HexToECDSA(private)
	if err != nil {
		return "", nil, err
	}
	hash, err := tx.Hash()
	if err != nil {
		return "", nil, err
	}
	sig, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return "", nil, err
	}
	//signedTx := &types.SignedTransaction{
	//	UnsignedTransaction: *tx,
	//	V:                   sig[64],
	//	S:                   sig[32:64],
	//	R:                   sig[0:32],
	//}

	//rawtx, err := signedTx.Encode()
	//if err != nil {
	//	return "", nil, err
	//}
	rawtx, err := tx.EncodeWithSignature(sig[64], sig[0:32], sig[32:64])
	if err != nil {
		return "", nil, err
	}
	return hex.EncodeToString(hash), rawtx, nil
}
