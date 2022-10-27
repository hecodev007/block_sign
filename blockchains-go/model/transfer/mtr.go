package transfer

import "encoding/json"

type MtrOrderRequest struct {
	OrderRequestHead
	FromAddr      string `json:"fromAddr"`
	ToAddr        string `json:"toAddr"` //支持多个输出，但是目前业务不需要，因此限制一个
	ToAmountInt64 string `json:"toAmountInt64"`
	FeeAddr       string `json:"feeAddr,omitempty"`
	Token         int64  `json:"token"` //目前限制0和1。0=MTR 1= MTRG
}

type MtrBalanceResp struct {
	Code int
	Data []MtrBalance
}

type MtrBalance struct {
	CoinName     string `json:"coinName"`
	Decimal      int    `json:"decimal"`
	BalanceFloat string `json:"balanceFloat"`
}

func DecodeMtrBalanceResp(data []byte) *MtrBalanceResp {
	if len(data) != 0 {
		result := new(MtrBalanceResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
