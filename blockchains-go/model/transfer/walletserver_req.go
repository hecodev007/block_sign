package transfer

import "github.com/shopspring/decimal"

//waletserver请求参数
type DirType int

const (
	//0:from, 1: to
	DirTypeFrom   DirType = 0 //from地址
	DirTypeTo     DirType = 1 //to地址
	DirTypeChange DirType = 2 //找零地址
)

type OrderRequestHeadV1 struct {
	MchId    string `json:"mch_no" `
	MchName  string `json:"mch_name" binding:"required"`
	OrderId  string `json:"order_no" `
	CoinName string `json:"coin_name" binding:"required"`
}

//type SignParams struct {
//	SignHeader
//	SignParams_Data
//}
//type SignParams_Data struct {
//	FromAddress string `json:"from_address" binding:"required"`
//	ToAddress   string `json:"to_address" binding:"required"`
//	Token       string `json:"token"`                       //telos主币是：“eosio.token”
//	Quantity    string `json:"quantity" binding:"required"` //“1.001 TLOS”
//	Memo        string `json:"memo,omitempty"`
//	SignPubKey  string `json:"sign_pubkey" binding:"required"`
//	BlockID     string `json:"block_id" ` //最新10w个高度内的一个block ID,like:“0637f2d29169db2dfd3dfee61982edee74fa193bb8648b6419ed2749b08ed7d6”(所属高度104329938)
//}

type OrderRequestHead struct {
	ApplyId        int64  `json:"apply_id,omitempty"`
	ApplyCoinId    int64  `json:"apply_coin_id,omitempty"`
	OuterOrderNo   string `json:"outer_order_no,omitempty"`
	OrderNo        string `json:"order_no,omitempty"`
	MchId          int64  `json:"mch_id,omitempty"`
	MchName        string `json:"mch_name,omitempty"`
	CoinName       string `json:"coin_name,omitempty"`
	Worker         string `json:"worker,omitempty"` //指定机器运行
	RecycleAddress string //零散归集的时候使用，指定from
	// write by flynn  2021-01-20   这两个参数用于热钱包的参数验证
	Sign        string `json:"sign"`
	CurrentTime string `json:"current_time"`
}

type ReqGetBalanceParams struct {
	CoinName                string      `json:"coin_name"`        // 币种主链的名字
	Address                 string      `json:"address"`          //	需要获取余额的地址
	Token                   string      `json:"token"`            // 	token的名字
	ContractAddress         string      `json:"contract_address"` //合约地址
	Params                  interface{} `json:"params"`           //特殊参数（如果有特殊参数，传入到这里面）
	OriginalContractAddress string      `json:"original_contract_address"`
}
type OrderRequestV1 struct {
	OrderRequestHeadV1
	FromAddress string `json:"from_address" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Token       string `json:"token"`                       //telos主币是：“eosio.token”
	Quantity    string `json:"quantity" binding:"required"` //“1.001 TLOS”
	Memo        string `json:"memo,omitempty"`
	SignPubKey  string `json:"sign_pubkey" binding:"required"`
	BlockID     string `json:"block_id" ` //最新10w个高度内的一个block ID,like:“0637f2d29169db2dfd3dfee61982edee74fa193bb8648b6419ed2749b08ed7d6”(所属高度104329938)

}

type OrderRequest struct {
	OrderRequestHead
	Id            int64               `json:"id,omitempty"`
	FromAddress   string              `json:"from_address,omitempty"`
	ToAddress     string              `json:"to_address,omitempty"`
	ChangeAddress string              `json:"change_address,omitempty"`
	Amount        int64               `json:"amount,omitempty"`
	Token         string              `json:"token,omitempty"`
	Quantity      string              `json:"quantity,omitempty"`
	Memo          string              `json:"memo,omitempty"`
	Fee           int64               `json:"fee,omitempty"`
	Feestr        decimal.Decimal     `json:"feestr,omitempty"`
	Decimal       int                 `json:"decimal,omitempty"`
	SignPubKey    string              `json:"sign_pubkey,omitempty"`
	IsRetry       bool                `json:"is_retry,omitempty"`
	CreateData    string              `json:"createData,omitempty"`
	OrderAddress  []*OrderAddrRequest `json:"order_address,omitempty"`
	ExpiryHeight  int64               `json:"expiryHeight,omitempty"` //zec yec过期高度
	IsForce       bool                `json:"is_force,omitempty"`
}

type OrderAddrRequest struct {
	Id           int64   `json:"id,omitempty"`
	OrderId      int64   `json:"order_id,omitempty"`
	Dir          DirType `json:"dir"`
	Address      string  `json:"address"`
	Amount       int64   `json:"amount"`
	TokenAmount  int64   `json:"tokenAmount,omitempty"`
	TxID         string  `json:"txId,omitempty"`
	Vout         int     `json:"vout,omitempty"`
	ScriptPubKey string  `json:"scriptPubKey,omitempty"`
	CreateAt     int64   `json:"create_at,omitempty"`
	UpdateAt     int64   `json:"update_at,omitempty"`
	Quantity     string  `json:"quantity,omitempty"` // 字符串金额
	MuxId        string  `json:"muxId,omitempty"`    //btm使用的MuxId
}

type OrderRequest2 struct {
	OrderRequestHead
	Token        string                 `json:"token,omitempty"`
	Memo         string                 `json:"memo,omitempty"`
	Fee          decimal.Decimal        `json:"fee,omitempty"`
	Decimal      int                    `json:"decimal,omitempty"`
	OrderAddress []*OrderAddrRequest2   `json:"order_address,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

type OrderAddrRequest2 struct {
	Dir         DirType                `json:"dir"`
	Address     string                 `json:"address"`
	Amount      decimal.Decimal        `json:"amount"`
	TokenAmount decimal.Decimal        `json:"token_amount,omitempty"`
	TxID        string                 `json:"txId,omitempty"`
	Vout        int                    `json:"vout,omitempty"`
	InnerExtra  map[string]interface{} `json:"inner_extra,omitempty"`
}

func (r *OrderRequest2) GetToAddresses() []*OrderAddrRequest2 {
	var addrs []*OrderAddrRequest2
	for _, v := range r.OrderAddress {
		if v.Dir == DirTypeTo {
			addrs = append(addrs, v)
		}
	}
	return addrs
}

func (r *OrderRequest2) GetFromAddresses() []*OrderAddrRequest2 {
	var addrs []*OrderAddrRequest2
	for _, v := range r.OrderAddress {
		if v.Dir == DirTypeFrom {
			addrs = append(addrs, v)
		}
	}
	return addrs
}

func (r *OrderRequest2) GetChangeAddresses() []*OrderAddrRequest2 {
	var addrs []*OrderAddrRequest2
	for _, v := range r.OrderAddress {
		if v.Dir == DirTypeChange {
			addrs = append(addrs, v)
		}
	}
	return addrs
}

func (r *OrderRequest2) GetToAmount() decimal.Decimal {
	total := decimal.Zero
	for _, v := range r.OrderAddress {
		if v.Dir == DirTypeTo {
			total = total.Add(v.Amount)
		}
	}
	return total
}

func (r *OrderRequest2) GetExtraByKey(key string) interface{} {
	if v, ok := r.Extra[key]; ok {
		return v
	} else {
		return nil
	}
}

func (r *OrderAddrRequest2) GetInnerExtraByKey(key string) interface{} {
	if v, ok := r.InnerExtra[key]; ok {
		return v
	} else {
		return nil
	}
}
