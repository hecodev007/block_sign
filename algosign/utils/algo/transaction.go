package algo

import (
	"algoSign/common/validator"

	"github.com/algorand/go-algorand-sdk/future"

	"github.com/algorand/go-algorand-sdk/crypto"

	"encoding/hex"

	"crypto/ed25519"

	"github.com/algorand/go-algorand-sdk/types"
)

func BuildTx(params *validator.TelosSignParams) (*types.Transaction, error) {
	sparam := types.SuggestedParams{
		Fee:             types.MicroAlgos(params.TelosSignParams_Data.Fee.IntPart()),
		FirstRoundValid: types.Round(params.TransactionParams.LastRound),
		LastRoundValid:  types.Round(params.TransactionParams.LastRound + 1000),
		GenesisID:       params.TransactionParams.GenesisID,
		GenesisHash:     params.TransactionParams.GenesisHash,
		FlatFee:         true,
	}
	//log.Info(params.TelosSignParams_Data.Fee.IntPart())
	if params.Assert.IsZero() {
		tx, err := future.MakePaymentTxn(params.FromAddress, params.ToAddress, uint64(params.Value.IntPart()), nil, "", sparam)
		return &tx, err
	} else {
		tx, err := future.MakeAssetTransferTxn(params.FromAddress, params.ToAddress, uint64(params.Value.IntPart()), nil, sparam, "", params.Assert.BigInt().Uint64())
		return &tx, err
	}
}
func SignTx(tx *types.Transaction, pri string) (txid string, rawTx []byte, err error) {
	priBytes, err := hex.DecodeString(pri)
	if err != nil {
		return "", nil, err
	}
	key := ed25519.PrivateKey(priBytes)
	txid, rawTxBytes, err := crypto.SignTransaction(key, *tx)
	if err != nil {
		return "", nil, err
	}

	return txid, rawTxBytes, nil
}
