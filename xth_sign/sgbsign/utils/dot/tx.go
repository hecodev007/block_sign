package dot

import (
	"github.com/JFJun/bifrost-go/expand"
	"github.com/JFJun/bifrost-go/tx"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"

	"sgbsign/common/validator"
	"fmt"
)

func BuildTx(params *validator.TelosSignParams,privateKey string) (rawtx string ,err error){
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	originTx := tx.NewSubstrateTransaction(params.FromAddress, params.Nonce)
	ed, err := expand.NewMetadataExpand(params.Meta)
	if err != nil {
		return "", fmt.Errorf("get metadata expand error: %v", err)
	}
	call, err := ed.BalanceTransferKeepAliveCall(params.ToAddress, uint64(params.Amount.IntPart()))
	if err != nil {
		return "", fmt.Errorf("get Balances.transfer call error: %v", err)
	}

	originTx.SetGenesisHashAndBlockHash(params.GenesisHash,params.GenesisHash).
		SetSpecVersionAndCallId(params.SpecVersion, params.TransactionVersion).
		//SetEra(params.BlockNumber,0).
		SetCall(call)
	//获取私钥

	var (
		sig    string
		errSig error
	)
	sig, errSig = originTx.SignTransaction(privateKey, crypto.Sr25519Type)
	if errSig != nil {
		return "", fmt.Errorf("sign error,Err=[%v]", errSig)
	}

	return sig, nil
}