package luna

import (
	"encoding/hex"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/shopspring/decimal"
	"github.com/terra-money/core/app"
	"github.com/terra-money/terra.go/msg"
	"github.com/terra-money/terra.go/tx"
	"lunasign/common/validator"
)

func SignTx(params *validator.SignParams, pri string) (rawtx string, err error) {
	//uusd,uluna 两个主链币签名
	if len(params.Data.Token) < 12{
		return SignMainCoinTx(params, pri)
	} else {
		return SignTokenTx(params, pri)
	}
}
func SignTokenTx(params *validator.SignParams, pri string) (rawtx string, err error) {
	txBuilder := tx.NewTxBuilder(app.MakeEncodingConfig().TxConfig)
	//_ = txBuilder
	fromaddr,err := sdk.AccAddressFromBech32(params.Data.FromAddr)
	if err != nil {
		return "", err
	}
	toaddr,err := sdk.AccAddressFromBech32(params.Data.ToAddr)
	if err != nil {
		return "", err
	}
	contract,err :=sdk.AccAddressFromBech32(params.Data.Token)
	if err != nil {
		return "", err
	}
	transfer := new(TokenTransfer)
	transfer.Transfer.Recipient = toaddr.String()
	transfer.Transfer.Amount = decimal.NewFromInt(params.Data.Amount)
	execMsg,_ := json.Marshal(transfer)
	err = txBuilder.SetMsgs(
		msg.NewMsgExecuteContract(
			fromaddr,
			contract,
			execMsg,
			nil,
		),
	)
	if err != nil {
		return "", err
	}
	txBuilder.SetMemo(params.Data.Memo)
	txBuilder.SetFeeAmount(msg.Coins{
		msg.Coin{
			Denom:"uluna",
			Amount: sdk.NewInt(params.Data.Fee),
		},
	})
	txBuilder.SetGasLimit(params.Data.Gas)
	//pribytes := hex.DecodeString(params.Data)
	pribytes,err := hex.DecodeString(pri)
	if err != nil {
		return "", err
	}
	privKey,err := PrivKeyGen(pribytes)
	if err != nil {
		return "", err
	}
	//log.Info(pri)
	//log.Info(sdk.AccAddress(privKey.PubKey().Address()).String())
	//return "test",nil

	err = txBuilder.Sign(tx.SignModeLegacyAminoJSON, tx.SignerData{
		ChainID:       params.Data.ChainID,
		AccountNumber: params.Data.AccountNumber,
		Sequence:      params.Data.Sequence,
	}, privKey, true)
	if err != nil {
		return "", err
	}

	txbytes,err := txBuilder.GetTxBytes()
	if err != nil {
		return "", err
	}
	return "0x"+hex.EncodeToString(txbytes), err
}

func SignMainCoinTx(params *validator.SignParams, pri string) (rawtx string, err error) {
	txBuilder := tx.NewTxBuilder(app.MakeEncodingConfig().TxConfig)
	//_ = txBuilder
	fromaddr,err := sdk.AccAddressFromBech32(params.Data.FromAddr)
	if err != nil {
		return "", err
	}
	toaddr,err := sdk.AccAddressFromBech32(params.Data.ToAddr)
	if err != nil {
		return "", err
	}

	err = txBuilder.SetMsgs(
		msg.NewMsgSend(
			fromaddr,
			toaddr,
			msg.Coins{msg.Coin{
				Denom:params.Data.Token,
				Amount: sdk.NewInt(params.Data.Amount),
			}},
		),
	)
	if err != nil {
		return "", err
	}
	txBuilder.SetMemo(params.Data.Memo)
	txBuilder.SetFeeAmount(msg.Coins{
		msg.Coin{
			Denom:"uluna",
			Amount: sdk.NewInt(params.Data.Fee),
		},
	})
	txBuilder.SetGasLimit(params.Data.Gas)
	//pribytes := hex.DecodeString(params.Data)
	pribytes,err := hex.DecodeString(pri)
	if err != nil {
		return "", err
	}
	privKey,err := PrivKeyGen(pribytes)
	if err != nil {
		return "", err
	}
	//log.Info(pri)
	//log.Info(sdk.AccAddress(privKey.PubKey().Address()).String())
	//return "test",nil

	err = txBuilder.Sign(tx.SignModeLegacyAminoJSON, tx.SignerData{
		ChainID:       params.Data.ChainID,
		AccountNumber: params.Data.AccountNumber,
		Sequence:      params.Data.Sequence,
	}, privKey, true)
	if err != nil {
		return "", err
	}

	txbytes,err := txBuilder.GetTxBytes()
	if err != nil {
		return "", err
	}
	return "0x"+hex.EncodeToString(txbytes), err
}


type TokenTransfer struct {
	Transfer struct{
		Recipient string `json:"recipient"`
		Amount decimal.Decimal `json:"amount"`
	} `json:"transfer"`
}