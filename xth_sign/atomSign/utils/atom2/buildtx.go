package atom2

import (
	"atomSign/common/validator"
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptoAmino "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	MainnetDenom  = "uatom"
	GasAdjustment = 1.0
)

func SignTx(params *validator.SignParams, pri []byte) (rawtx string, err error) {
	//txCfg := NewTxConfig()
	txCfg := simapp.MakeTestEncodingConfig().TxConfig
	txBuilder := txCfg.NewTxBuilder()
	sendMsg, err := MakeMsgSend(params.FromAddr, params.ToAddr, params.Amount)
	if err != nil {
		return "", err
	}
	fee := sdk.Coins{{MainnetDenom, sdk.NewInt(params.Fee)}}
	txBuilder.SetMsgs(sendMsg)
	txBuilder.SetFeeAmount(fee)
	txBuilder.SetGasLimit(params.Gas)
	txBuilder.SetMemo(params.Memo)
	txBuilder.SetTimeoutHeight(0)

	var privKey secp256k1.PrivKey
	privKey.Key = pri[:]
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())
	if params.FromAddr != acc.String() {
		panic("from地址与私钥不匹配" + params.FromAddr + " != " + acc.String())
	}

	signingData := authsigning.SignerData{
		ChainID:       params.ChainID,
		AccountNumber: params.AccountNumber,
		Sequence:      params.Sequence,
	}
	//verification failed; please verify account number (96689) and chain-id (cosmoshub-4): unauthorized
	fmt.Println(params.ChainID, params.AccountNumber, params.Sequence)
	modeHandler := txCfg.SignModeHandler()
	//txCfg.SignModeHandler()
	sig, err := tx.SignWithPrivKey(modeHandler.Modes()[1], signingData, txBuilder, &privKey, txCfg, params.Sequence)
	if err != nil {
		return "", err
	}
	//fmt.Println(modeHandler.Modes()[1], int(modeHandler.Modes()[1]))
	//signBytes, err := modeHandler.GetSignBytes(modeHandler.DefaultMode(), signingData, txBuilder.GetTx())
	//if err != nil {
	//	return "", err
	//}
	//sigData.Signature, err = privKey.Sign(signBytes)
	//if err != nil {
	//	return "", err
	//}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	rawtxbytes, err := txCfg.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	//jsonbytes, _ := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	//fmt.Println(string(jsonbytes))
	//txBuilder.GetTx().GetSignaturesV2()
	return "0x" + hex.EncodeToString(rawtxbytes), nil
}
func MakeMsgSend(fromAddr, toAddr string, amount int64) (*types.MsgSend, error) {

	to, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return nil, err
	}

	msg := types.NewMsgSend(from, to, sdk.Coins{{MainnetDenom, sdk.NewInt(amount)}})

	return msg, nil
}
func NewCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	cryptoAmino.RegisterCrypto(cdc)
	cdc.RegisterConcrete(&testdata.TestMsg{}, "cosmos-sdk/Main", nil)
	return cdc
}

func NewTxConfig() legacytx.StdTxConfig {
	cdc := NewCodec()
	txGen := legacytx.StdTxConfig{Cdc: cdc}
	return txGen
}
