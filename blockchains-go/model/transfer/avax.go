package transfer

import "encoding/json"

type AvaxOrderRequest struct {
	OrderRequestHead
	Amount       int64                   `json:"amount,omitempty"`
	Fee          int64                   `json:"fee,omitempty"`
	OrderAddress []*AvaxOrderAddrRequest `json:"order_address,omitempty"`
}

type AvaxOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type AvaxGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//热钱包结构

//{
//"coin_name": "avax",
//"order_no": "12306",
//"mch_name": "goapi",
//"fromAddr": "x-avax1enmnreekmjhrfk88znkfsut6z5gu92fhde3ypg",
//"toAddr": "x-avax13hmeyeczh8stl80cx58r3ejk90wecm42erpcnn",
//"amount": 10000,
//"fee": 400000,
//"utxos": [
//"111111111NPFRLJ5udj3uwynofYfhqRUsg32CqhaTmXyjLiTLvjCUUqiVvKssQNF8nTWnnGSu7cgFFtJHvrkzLpB8ieryXgVCjbNSYczpPPfAXgfptEucL9ffckzKL6hnhoPANftAppQJb1X8zw6w86uhRAeoZqXjHVv7XrjCVTTLHNkevXCmARTd3yYCXP5d239ienbN4Cm4XaXoRddvdJk5yH1zwP5qvUkx6qEAurvkLXPEbfb8mHXA2fzkTUWWw9mYTKM5eJ6nFXVyFDMFRafvzJHWrQNSWqaoEYm6EZsAFmMcqST9LEeJyih3p4vvJnvQW58JWaGuMG2h4ARgebetEj4tECdQpGZF7CZjU7iW5J5RopjMzP8y8uGxihNzXMsuPahqD8uPhUKKeqNrLiBp227mpkDWLTBPFKf8TrnNhj3SLjKcWPfW3Q9JYtMi8wB5jn78F24m1ZLvZEUGxZQon26Equi5hC2v8aaQvCr28rJ3FHXuWbeHznyUhtXML76Ctk"
//]
//}
type AvaxTxTpl struct {
	CoinName   string   `json:"coinName"`
	OrderNo    string   `json:"orderNo"`
	MchName    string   `json:"mchName"`
	FromAddr   string   `json:"fromAddr"`
	ToAddr     string   `json:"toAddr"`
	ChangeAddr string   `json:"changeAddr"`
	Amount     int64    `json:"amount"`
	Fee        int64    `json:"fee"`
	Utxos      []string `json:"utxos"`
}

//交易接口，返回orderId
type AvaxRespTranfer struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
	Txid    string `json:"txid"`
}

func DecodRespTranfer(data []byte) *AvaxRespTranfer {
	if len(data) != 0 {
		result := new(AvaxRespTranfer)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
