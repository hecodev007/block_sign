package transfer

import (
	"encoding/json"
	"fmt"
)

type StxToAddrAmount struct {
	ToAddr   string `json:"address"`
	ToAmount int64  `json:"amount"`
}

type StxOrderRequest struct {
	ApplyId        int64             `json:"applyid"`
	OuterOrderNo   string            `json:"outerorderno"`
	OrderNo        string            `json:"orderno"`
	MchName        string            `json:"mchname"`
	CoinName       string            `json:"coinname"`
	FromAddress    string            `json:"fromAddress"`
	ChanegeAddress string            `json:"changeAddress"`
	ToAddrs        []StxToAddrAmount `json:"toaddress"`
	TransferFee    int64             `json:"transferfee"`
	Memo           string            `json:"memo"`
}

func DecodeStxTransferResp(data []byte) map[string]interface{} {
	var result map[string]interface{}
	if len(data) != 0 {
		err := json.Unmarshal(data, &result)
		if err == nil {
			return result
		} else {
			fmt.Printf("parse response data error,err=%v", err)
		}
	}
	return nil
}
