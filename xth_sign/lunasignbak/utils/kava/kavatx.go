package kava

import (
	"encoding/hex"
	"errors"
	"terrasign/common/validator"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/kava-labs/kava/app"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func SignTx(params *validator.SignParams, pri []byte) (rawtx string, err error) {
	to, err := sdk.AccAddressFromBech32(params.Data.ToAddr)
	from, err := sdk.AccAddressFromBech32(params.Data.FromAddr)
	codec := app.MakeCodec()

	var privKey secp256k1.PrivKeySecp256k1
	copy(privKey[:], pri[:])
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())
	if acc.String() != params.Data.FromAddr {
		return "", errors.New("私钥与from地址不匹配")
	}
	stdTx := GenTx(
		[]sdk.Msg{
			bank.NewMsgSend(
				from,
				to,
				sdk.NewCoins(sdk.NewInt64Coin("ukava", params.Data.Amount)),
			),
		},
		sdk.NewCoins(sdk.NewInt64Coin("ukava", params.Data.Fee)), // no fee
		helpers.DefaultGenTxGas,
		params.Data.ChainID,
		[]uint64{params.Data.AccountNumber},
		[]uint64{params.Data.Sequence}, // fixed sequence numbers will cause tests to fail sig verification if the same address is used twice
		params.Data.Memo,
		privKey,
	)
	txBytes, err := auth.DefaultTxEncoder(codec)(stdTx)
	return "0x" + hex.EncodeToString(txBytes), err
}
func GenTx(msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accnums []uint64, seq []uint64, memo string, priv ...crypto.PrivKey) auth.StdTx {
	fee := auth.StdFee{
		Amount: feeAmt,
		Gas:    gas,
	}

	sigs := make([]auth.StdSignature, len(priv))

	for i, p := range priv {
		// use a empty chainID for ease of testing
		sig, err := p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, memo))
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    p.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, memo)
}
