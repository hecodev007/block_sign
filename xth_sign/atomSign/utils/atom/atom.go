package atom

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/std"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

const (
	MainnetDenom  = "uatom"
	GasAdjustment = 1.0
)

var atomCdc = makeCodec()

type Account struct {
	Address    sdk.AccAddress
	PublicKey  crypto.PubKey
	PrivateKey secp256k1.PrivKey
}

// custom tx codec
func makeCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()

	bm.RegisterLegacyAminoCodec(cdc)
	std.RegisterLegacyAminoCodec(cdc)

	return cdc
}

// GenerateAccount generates a random Account
func GenerateAccount() (act Account) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())

	// Build the account
	act.Address = acc
	act.PublicKey = pubKey
	act.PrivateKey = privKey
	return
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

func MakeAccount(sk []byte) (*Account, error) {
	if len(sk) != 32 {
		return nil, fmt.Errorf("make account sk len isn't 32 , %d", len(sk))
	}

	var privKey secp256k1.PrivKey
	copy(privKey[:], sk[:32])
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())

	// Build the account
	act := &Account{}
	act.Address = acc
	act.PublicKey = pubKey
	act.PrivateKey = privKey
	return act, nil
}

func MakeTxBuilder(acc Account, accNumber, seq, gas uint64, fee int64, chainID, memo string) (*TxBuilder, error) {
	//txEnc := auth.DefaultTxEncoder(atomCdc)
	//gasPrices := sdk.DecCoins{
	//sdk.NewDecCoinFromDec(MainnetDenom,
	//sdk.NewDecWithPrec(GasPrices, sdk.Precision))}
	return NewTxBuilder(
		auth.DefaultTxEncoder(atomCdc),
		acc,
		accNumber,
		seq,
		gas,
		GasAdjustment,
		false,
		chainID,
		memo,
		sdk.Coins{{MainnetDenom, sdk.NewInt(fee)}}), nil
}

func MakeSignTx(txBldr *TxBuilder, msgs []sdk.Msg) ([]byte, error) { // build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(msgs)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func DecoderTx(txBytes []byte) (sdk.Tx, error) {
	var tx = auth.StdTx{}

	if len(txBytes) == 0 {
		return nil, fmt.Errorf("txBytes are empty")
	}

	// StdTx.Msg is an interface. The concrete types
	// are registered by MakeTxCodec
	err := atomCdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	if err != nil {
		return nil, fmt.Errorf("error decoding transaction %v", err)
	}

	return tx, nil
}
