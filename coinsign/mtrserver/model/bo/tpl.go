package bo

import (
	"errors"
	"github.com/group-coldwallet/mtrserver/model"
	"github.com/group-coldwallet/mtrserver/pkg/mtrutil"
	"github.com/shopspring/decimal"
	"strings"
)

//mtr 签名模板
type TxTpl struct {
	mtrutil.MtrTpl
}

type MtrParams struct {
	model.MchInfoReq
	FromAddr      string `json:"fromAddr"`
	ToAddr        string `json:"toAddr"` //支持多个输出，但是目前业务不需要，因此限制一个
	ToAmountInt64 string `json:"toAmountInt64"`
	FeeAddr       string `json:"feeAddr,omitempty"`
	Token         int64  `json:"token"` //目前限制0和1。0=MTR 1= MTRG
	BlockRef      uint32 `json:"blockRef"`
	//21000，类似gaslimit
	//Gas uint64 `json:"gas"`
	//gasprice
	//GasPriceCoef uint8 `json:"gasPriceCoef"`
}

func (m *MtrParams) Check() error {
	if strings.ToLower(m.CoinName) != "mtr" {
		return errors.New("error coinName:" + strings.ToLower(m.CoinName))
	}
	if strings.ToLower(m.FromAddr) == "" {
		return errors.New("error FromAddr")
	}
	if strings.ToLower(m.ToAddr) == "" {
		return errors.New("error ToAddr")
	}
	if m.Token != 0 && m.Token != 1 {
		return errors.New("error token")
	}
	am, err := decimal.NewFromString(m.ToAmountInt64)
	if err != nil {
		return err
	}
	if am.LessThanOrEqual(decimal.Zero) {
		return errors.New("error amount")
	}
	return nil
}
