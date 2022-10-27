package validator

import (
	"errors"
	"steemsign/utils/rpc/types"
	"time"
)

type ColdSign struct {
	ColdData
	ChainID   string `json:"chain_id" binding:"required"`
	EosCode   string `json:"eos_code" binding:"required"`
	Hash      string `json:"hash" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}
type ColdData struct {
	CoinName       string `json:"coinName" binding:"required"`
	MchID          string `json:"mchId" binding:"required"`
	OrderID        string `json:"orderId" binding:"required"`
	Expiration     Time   `json:"expiration" binding:"required"`
	RefBlockNum    uint32 `json:"ref_block_num" binding:"required"`
	RefBlockPrefix uint32 `json:"ref_block_prefix" binding:"required"`
	Account        string `json:"account" binding:"required"`
	Actor          string `json:"actor" binding:"required"`
	Data           string `json:"data" binding:"required"`
}
type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	formarter := "2006-01-02T15:04:05.000"
	if y := time.Time(t).Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	time.Now().String()
	b := make([]byte, 0, len(formarter)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, formarter)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Time) UnmarshalJSON(data []byte) error {
	formarter := "2006-01-02T15:04:05.000"
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	//var err error
	tmp, err := time.Parse(`"`+formarter+`"`, string(data))
	if err != nil {
		return err
	}
	*t = Time(tmp)
	return err
}

type ColdSignResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ColdData
		Signatures string `json:"signatures"`
	} `json:"data"`
}

type SignHeader struct {
	MchId    string `json:"mch_no" `
	MchName  string `json:"mch_name" binding:"required"`
	OrderId  string `json:"order_no" `
	CoinName string `json:"coin_name" binding:"required"`
}

type SignParams struct {
	SignHeader
	SignParams_Data
}
type SignParams_Data struct {
	FromAddress string `json:"from_address" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Token       string `json:"token"`                       //telos主币是：“eosio.token”
	Quantity    string `json:"quantity" binding:"required"` //“1.001 TLOS”
	Memo        string `json:"memo,omitempty"`
	SignPubKey  string `json:"sign_pubkey" `
	BlockID     string `json:"block_id" ` //最新10w个高度内的一个block ID,like:“0637f2d29169db2dfd3dfee61982edee74fa193bb8648b6419ed2749b08ed7d6”(所属高度104329938)
}

type SignReturns_data struct {
	*types.Transaction
	TxHash interface{} `json:"txid"`
}

type SignReturns struct {
	SignHeader
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    SignReturns_data `json:"data"`
}

type TransferReturns struct {
	//SignHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"txid"` //txid
}

type GetBalanceParams struct {
	CoinName string `json:"coin_name"`
	Address  string `json:"address" binding:"required"`
	Token    string `json:"contract_address" binding:"required"`
	Params   Params `json:"params"`
}
type Params struct {
	Symbol string `json:"symbol"` //币缩写 eg:bos
}
type GetBalanceReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"` //数值
}
type ValidAddressParams struct {
	Address string `json:"address"`
}
type ValidAddressReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    bool   `json:"data"` //数值
}
