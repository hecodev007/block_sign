package transfer

import (
	"encoding/json"
	"fmt"
)

type FlowOrderRequest struct {
	ApplyId        int64  `json:"apply_id,omitempty"`
	ApplyCoinId    int64  `json:"apply_coin_id,omitempty"`
	OuterOrderNo   string `json:"outer_order_no,omitempty"`
	OrderNo        string `json:"order_no,omitempty"`
	MchId          string `json:"mchId,omitempty"`
	MchName        string `json:"mch_name,omitempty"`
	CoinName       string `json:"coin_name,omitempty"`
	Worker         string `json:"worker,omitempty"` //指定机器运行
	RecycleAddress string //零散归集的时候使用，指定from
	Sign           string `json:"sign"`
	CurrentTime    string `json:"current_time"`
	FromAddress    string `json:"from"`
	ToAddress      string `json:"to"` // '接收者地址'
	Amount         int64  `json:"amount"`
	//  token 转账 需要参数
	ContractAddress string `json:"contract_address"`
	Token           string `json:"token"` // 代币的名字，主链转账不传这个值

}

func DecodeFlowTransferResp(data []byte) map[string]interface{} {
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
