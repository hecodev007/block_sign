package atom

import (
	"atomSign/common/validator"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"atomSign/common/log"
	"sync"
)

type Signer struct {
	CoinName string
	lock     *sync.Mutex
}

type TxSignParam struct {
	FromAddr      string `json:"from_addr"`
	ToAddr        string `json:"to_addr"`
	Amount        int64  `json:"amount"`
	AccountNumber uint64 `json:"account_number"`
	ChainID       string `json:"chain_id"`
	Sequence      uint64 `json:"sequence"`
	Memo          string `json:"memo"`
	Gas           uint64 `json:"gas"`
	Fee           int64  `json:"fee"`
}

func CreateSigner() *Signer {
	return &Signer{
		CoinName: "atom",
		lock:     &sync.Mutex{},
	}
}


func (s *Signer) SignTx(signParam *validator.SignParams,pri string) (rawtx string,err error) {
	log.Info(pri)
	act,err :=MakeAccount([]byte(pri))
	if err != nil {
		return "",err
	}
	signData, err := s.signPaymentTx(signParam, act)
	if err != nil {
		return "", err
	}
	//return string(Base64Encode(signData)), nil
	return "0x"+hex.EncodeToString(signData),nil
}

func (s *Signer) signPaymentTx(req *validator.SignParams, act *Account) ([]byte, error) {

	msg, err := MakeMsgSend(req.FromAddr, req.ToAddr, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("make msg send err: %v", err)
	}

	builder, err := MakeTxBuilder(*act, req.AccountNumber, req.Sequence, req.Gas, req.Fee, req.ChainID, req.Memo)
	if err != nil {
		return nil, fmt.Errorf("make tx builder err: %v", err)
	}

	stx, err := MakeSignTx(builder, []sdk.Msg{msg})
	if err != nil {
		log.Infof("Failed to sign transaction: %s\n", err)
		return nil, fmt.Errorf("failed to sign transaction: %s", err)
	}

	return stx, nil
}
func Base64Encode(data []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return dst
}