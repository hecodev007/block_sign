package crust

import (
	"fmt"

	"github.com/yanyushr/go-substrate-rpc-client/v3/signature"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"

	"crustsign/common/validator"
)

// func BuildTx(params *validator.TelosSignParams, privateKey string) (rawtx string, err error) {
// 	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

// 	originTx := tx.NewSubstrateTransaction(params.FromAddress, params.Nonce)
// 	ed, err := expand.NewMetadataExpand(params.Meta)
// 	if err != nil {
// 		return "", fmt.Errorf("get metadata expand error: %v", err)
// 	}
// 	call, err := ed.BalanceTransferCall(params.ToAddress, uint64(params.Amount.IntPart()))
// 	if err != nil {
// 		return "", fmt.Errorf("get Balances.transfer call error: %v", err)
// 	}

// 	originTx.SetGenesisHashAndBlockHash(params.GenesisHash, params.BlockHash).
// 		SetSpecVersionAndCallId(params.SpecVersion, params.TransactionVersion).
// 		SetCall(call)
// 	//获取私钥

// 	var (
// 		sig    string
// 		errSig error
// 	)
// 	sig, errSig = originTx.SignTransaction(privateKey, crypto.Sr25519Type)
// 	if errSig != nil {
// 		return "", fmt.Errorf("sign error,Err=[%v]", errSig)
// 	}

// 	log.Info("********:", originTx)

// 	return sig, nil
// }

func BuildTx(params *validator.TelosSignParams, privateKey string) (*types.Extrinsic, error) {
	pubKey := GetPublicFromAddr(params.FromAddress, CRustPrefix)

	from := signature.KeyringPair{
		URI:       privateKey,
		Address:   params.FromAddress,
		PublicKey: pubKey,
	}

	toKey := GetPublicFromAddr(params.ToAddress, CRustPrefix)
	to := types.NewAccountID(toKey)

	// newAmount := params.Amount.Mul(decimal.NewFromInt(1000000000000))

	// ss := types.NewUCompactFromUInt(newAmount.BigInt().Uint64())

	c, err := types.NewCall(params.Meta, "Balances.transfer", to, types.NewUCompactFromUInt(uint64(params.Amount.IntPart())))
	if err != nil {
		return nil, err
	}
	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, _ := types.NewHashFromHexString(params.GenesisHash)
	blockHash, _ := types.NewHashFromHexString(params.BlockHash)
	o := types.SignatureOptions{
		BlockHash:          blockHash,
		Era:                types.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(params.Nonce),
		SpecVersion:        types.U32(params.SpecVersion),
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: types.U32(params.TransactionVersion),
	}

	// ee := getEra(params.BlockNumber, params.BlockNumber-5)
	// if ee != nil {
	// 	o.Era = *ee
	// }

	fmt.Println("***********:", o)

	err = ext.Sign(from, o)
	if err != nil {
		return nil, err
	}

	return &ext, nil
}

func getEra(blockNumber uint64, eraPeriod uint64) *types.ExtrinsicEra {
	if blockNumber == 0 || eraPeriod == 0 {
		return nil
	}
	phase := blockNumber % eraPeriod
	index := uint64(6)
	trailingZero := index - 1

	var encoded uint64
	if trailingZero > 1 {
		encoded = trailingZero
	} else {
		encoded = 1
	}

	if trailingZero < 15 {
		encoded = trailingZero
	} else {
		encoded = 15
	}
	encoded += phase / 1 << 4
	first := byte(encoded >> 8)
	second := byte(encoded & 0xff)
	era := new(types.ExtrinsicEra)
	era.IsMortalEra = true
	era.AsMortalEra.First = first
	era.AsMortalEra.Second = second
	return era
}
