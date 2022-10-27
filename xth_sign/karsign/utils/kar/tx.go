package kar

import (
	"github.com/yanyushr/go-substrate-rpc-client/v3/signature"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
)

func BuildTx(params *Txparam, privateKey string) (*types.Extrinsic, error) {
	pubKey := GetPublicFromAddr(params.FromAddress, KARPrefix)

	from := signature.KeyringPair{
		URI:       privateKey,
		Address:   params.FromAddress,
		PublicKey: pubKey,
	}

	toKey := GetPublicFromAddr(params.ToAddress, KARPrefix)
	to := types.MultiAddress{
		IsID: true,
		AsID: types.NewAccountID(toKey),
	}

	c, err := types.NewCall(params.Meta, "Balances.transfer_keep_alive", to, types.NewUCompactFromUInt(uint64(params.Amount.IntPart())))
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

	return &ext, nil
}
