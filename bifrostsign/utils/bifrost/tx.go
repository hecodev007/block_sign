package bifrost

import (
	"github.com/yanyushr/go-substrate-rpc-client/v3/signature"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"

	"bncsign/common/validator"
)

func BuildTx(params *validator.TelosSignParams, privateKey string) (*types.Extrinsic, error) {
	pubKey := GetPublicFromAddr(params.FromAddress, BNCPrefix)

	from := signature.KeyringPair{
		URI:       privateKey,
		Address:   params.FromAddress,
		PublicKey: pubKey,
	}

	toKey := GetPublicFromAddr(params.ToAddress, BNCPrefix)
	to := types.MultiAddress{
		IsID: true,
		AsID: types.NewAccountID(toKey),
	}

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

	err = ext.Sign(from, o)
	if err != nil {
		return nil, err
	}
	//log.Info(ext)

	return &ext, nil
}
