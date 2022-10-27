package atom

import (
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"strings"
)

// TxBuilder implements a transaction context created in SDK modules.
type TxBuilder struct {
	txEncoder          sdk.TxEncoder
	account            Account
	accountNumber      uint64
	sequence           uint64
	gas                uint64
	gasAdjustment      float64
	simulateAndExecute bool
	chainID            string
	memo               string
	fees               sdk.Coins
	gasPrices          sdk.DecCoins
}

// NewTxBuilder returns a new initialized TxBuilder.
func NewTxBuilder(txEncoder sdk.TxEncoder, acc Account, accNumber, seq, gas uint64, gasAdj float64,
	simulateAndExecute bool, chainID, memo string, fees sdk.Coins) *TxBuilder {

	return &TxBuilder{
		txEncoder:          txEncoder,
		account:            acc,
		accountNumber:      accNumber,
		sequence:           seq,
		gas:                gas,
		gasAdjustment:      gasAdj,
		simulateAndExecute: simulateAndExecute,
		chainID:            chainID,
		memo:               memo,
		fees:               fees,
	}
}

// TxEncoder returns the transaction encoder
func (bldr TxBuilder) TxEncoder() sdk.TxEncoder { return bldr.txEncoder }

// AccountNumber returns the account number
func (bldr TxBuilder) AccountNumber() uint64 { return bldr.accountNumber }

// Sequence returns the transaction sequence
func (bldr TxBuilder) Sequence() uint64 { return bldr.sequence }

// Gas returns the gas for the transaction
func (bldr TxBuilder) Gas() uint64 { return bldr.gas }

// GasAdjustment returns the gas adjustment
func (bldr TxBuilder) GasAdjustment() float64 { return bldr.gasAdjustment }

// Keybase returns the keybase
func (bldr TxBuilder) Account() Account { return bldr.account }

// SimulateAndExecute returns the option to simulate and then execute the transaction
// using the gas from the simulation results
func (bldr TxBuilder) SimulateAndExecute() bool { return bldr.simulateAndExecute }

// ChainID returns the chain id
func (bldr TxBuilder) ChainID() string { return bldr.chainID }

// Memo returns the memo message
func (bldr TxBuilder) Memo() string { return bldr.memo }

// Fees returns the fees for the transaction
func (bldr TxBuilder) Fees() sdk.Coins { return bldr.fees }

// GasPrices returns the gas prices set for the transaction, if any.
func (bldr TxBuilder) GasPrices() sdk.DecCoins { return bldr.gasPrices }

// WithTxEncoder returns a copy of the context with an updated codec.
func (bldr TxBuilder) WithTxEncoder(txEncoder sdk.TxEncoder) TxBuilder {
	bldr.txEncoder = txEncoder
	return bldr
}

// WithChainID returns a copy of the context with an updated chainID.
func (bldr TxBuilder) WithChainID(chainID string) TxBuilder {
	bldr.chainID = chainID
	return bldr
}

// WithGas returns a copy of the context with an updated gas.
func (bldr TxBuilder) WithGas(gas uint64) TxBuilder {
	bldr.gas = gas
	return bldr
}

// WithFees returns a copy of the context with an updated fee.
func (bldr TxBuilder) WithFees(fees string) TxBuilder {
	parsedFees, err := sdk.ParseCoins(fees)
	if err != nil {
		panic(err)
	}

	bldr.fees = parsedFees
	return bldr
}

// WithGasPrices returns a copy of the context with updated gas prices.
func (bldr TxBuilder) WithGasPrices(gasPrices string) TxBuilder {
	parsedGasPrices, err := sdk.ParseDecCoins(gasPrices)
	if err != nil {
		panic(err)
	}

	bldr.gasPrices = parsedGasPrices
	return bldr
}

// WithKeybase returns a copy of the context with updated keybase.
func (bldr TxBuilder) WithAccount(acc Account) TxBuilder {
	bldr.account = acc
	return bldr
}

// WithSequence returns a copy of the context with an updated sequence number.
func (bldr TxBuilder) WithSequence(sequence uint64) TxBuilder {
	bldr.sequence = sequence
	return bldr
}

// WithMemo returns a copy of the context with an updated memo.
func (bldr TxBuilder) WithMemo(memo string) TxBuilder {
	bldr.memo = strings.TrimSpace(memo)
	return bldr
}

// WithAccountNumber returns a copy of the context with an account number.
func (bldr TxBuilder) WithAccountNumber(accnum uint64) TxBuilder {
	bldr.accountNumber = accnum
	return bldr
}

// BuildSignMsg builds a single message to be signed from a TxBuilder given a
// set of messages. It returns an error if a fee is supplied but cannot be
// parsed.
func (bldr TxBuilder) BuildSignMsg(msgs []sdk.Msg) (StdSignMsg, error) {
	if bldr.chainID == "" {
		return StdSignMsg{}, fmt.Errorf("chain ID required but not specified")
	}

	fees := bldr.fees
	if !bldr.gasPrices.IsZero() {
		if !fees.IsZero() {
			return StdSignMsg{}, errors.New("cannot provide both fees and gas prices")
		}

		glDec := sdk.NewDec(int64(bldr.gas))

		// Derive the fees based on the provided gas prices, where
		// fee = ceil(gasPrice * gasLimit).
		fees = make(sdk.Coins, len(bldr.gasPrices))
		for i, gp := range bldr.gasPrices {
			fee := gp.Amount.Mul(glDec)
			fees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	return StdSignMsg{
		ChainID:       bldr.chainID,
		AccountNumber: bldr.accountNumber,
		Sequence:      bldr.sequence,
		Memo:          bldr.memo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(bldr.gas, fees),
	}, nil
}

// Sign signs a transaction given a name, passphrase, and a single message to
// signed. An error is returned if signing fails.
func (bldr TxBuilder) Sign(msg StdSignMsg) ([]byte, error) {
	sig, err := MakeSignature(bldr.account, msg)
	if err != nil {
		return nil, err
	}

	return bldr.txEncoder(auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo))
}

// BuildAndSign builds a single message to be signed, and signs a transaction
// with the built message given a name, passphrase, and a set of messages.
func (bldr TxBuilder) BuildAndSign(msgs []sdk.Msg) ([]byte, error) {
	msg, err := bldr.BuildSignMsg(msgs)
	if err != nil {
		return nil, err
	}

	return bldr.Sign(msg)
}

// BuildTxForSim creates a StdSignMsg and encodes a transaction with the
// StdSignMsg with a single empty StdSignature for tx simulation.
func (bldr TxBuilder) BuildTxForSim(msgs []sdk.Msg) ([]byte, error) {
	signMsg, err := bldr.BuildSignMsg(msgs)
	if err != nil {
		return nil, err
	}

	// the ante handler will populate with a sentinel pubkey
	sigs := []auth.StdSignature{{}}
	return bldr.txEncoder(auth.NewStdTx(signMsg.Msgs, signMsg.Fee, sigs, signMsg.Memo))
}

// SignStdTx appends a signature to a StdTx and returns a copy of it. If append
// is false, it replaces the signatures already attached with the new signature.
func (bldr TxBuilder) SignStdTx(stdTx auth.StdTx, appendSig bool) (signedStdTx auth.StdTx, err error) {
	//if bldr.chainID == "" {
	//	return auth.StdTx{}, fmt.Errorf("chain ID required but not specified")
	//}
	//
	//stdSignature, err := MakeSignature(bldr.account, StdSignMsg{
	//	ChainID:       bldr.chainID,
	//	AccountNumber: bldr.accountNumber,
	//	Sequence:      bldr.sequence,
	//	Fee:           stdTx.Fee,
	//	Msgs:          stdTx.GetMsgs(),
	//	Memo:          stdTx.GetMemo(),
	//})
	//if err != nil {
	//	return
	//}
	//
	//sigs := stdTx.GetSignatures()
	//if len(sigs) == 0 || !appendSig {
	//	sigs = []auth.StdSignature{stdSignature}
	//} else {
	//	sigs = append(sigs, stdSignature)
	//}
	//signedStdTx = auth.NewStdTx(stdTx.GetMsgs(), stdTx.Fee, sigs, stdTx.GetMemo())
	return
}

// MakeSignature builds a StdSignature given keybase, key name, passphrase, and a StdSignMsg.
func MakeSignature(acc Account, msg StdSignMsg) (sig auth.StdSignature, err error) {

	sigBytes, err := acc.PrivateKey.Sign(msg.Bytes())
	if err != nil {
		return
	}
	return auth.StdSignature{
		PubKey:    acc.PublicKey,
		Signature: sigBytes,
	}, nil
}
