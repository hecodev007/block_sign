package rose

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature/signers/memory"
	"github.com/oasisprotocol/oasis-core/go/common/encoding/bech32"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
	"strings"
)
func GenAccount()(pri string,addr string,er error){
	seed := make([]byte,32)
	rand.Read(seed)
	rng :=bytes.NewReader(seed)
	fac := memory.NewFactory()
	siger,err := fac.Generate(signature.SignerEntity, rng)

	if err != nil {
		return "", "", err
	}

	pri = hex.EncodeToString(seed)
	addr = staking.NewAddress(siger.Public()).String()
	return
}

func PriToAddr(pri string)(addr string,err error){
	priBytes,err := hex.DecodeString(pri)
	if err != nil{
		return "",err
	}
	rng :=bytes.NewReader(priBytes)
	fac := memory.NewFactory()
	siger,err := fac.Generate(signature.SignerEntity, rng)
	if err != nil {
		return "", err
	}
	addr = staking.NewAddress(siger.Public()).String()
	return
}

func BuildTx(fromPri string,nonce uint64,toAddr string,amount uint64,fee uint64) (rawTx string,err error){
	_,toBytes,err := bech32.Decode(toAddr)

	toAddress := staking.Address{}
	copy(toAddress[:],toBytes)
	transfer := staking.Transfer{To:toAddress }
	transfer.Amount.FromUint64(amount)
	tx := staking.NewTransferTx(nonce, nil, &transfer)

	return toAddress.String(),nil
}
//校验地址
func VarifyAddress(addr string) error{
	if !strings.HasPrefix(addr,"oasis"){
		return errors.New(addr+" to地址前缀错误(oasis)")
	}
	if len(addr) != 46 {
		return errors.New(addr+" to地址长度错误(46)")

	}

	_,toBytes,err := bech32.Decode(addr)
	if err != nil{
		return errors.New(addr+" to地址校验错误")
	}
	address := staking.Address{}
	copy(address[:],toBytes)
	if address.String() != addr{
		return errors.New(addr+" to地址校验错误")
	}
	return nil
}