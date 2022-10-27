package nyzo

import (
	"errors"
	"time"
)

type InfoRet struct {
	RetentionEdge int64 `json:"frozen_edge"`
}

func (rpc *RpcClient) GetBlockCount() (int64, error) {
	return rpc.Info()
}
func (rpc *RpcClient) Info() (int64, error) {
	ret := new(InfoRet)
	err := rpc.CallNoAuth("info", ret, nil)
	if err != nil {
		return 0, err
	}
	if ret.RetentionEdge == 0 {
		return 0, errors.New("rpc请求:info,返回错误")
	}
	return ret.RetentionEdge, nil
}

type SendTransactionRet struct {
	Signature      string `json:"signature"`
	Raw            string `json:"raw"`
	ScheduledBlock int64  `json:"scheduled_block"`

	ValidationWarning string `json:"validation_warning"`
	ValidationError   string `json:"validation_error"`
}

func (rpc *RpcClient) SendTransaction(from string, to string, amount uint64, memo string, pri string, broadcast bool) (txhash string, rawtx string, err error) {
	params := make(map[string]interface{}, 0)
	params["sender_nyzo_string"] = from
	params["receiver_nyzo_string"] = to
	params["sender_data"] = memo
	params["private_nyzo_string"] = pri
	params["amount"] = amount
	params["broadcast"] = broadcast

	ret := new(SendTransactionRet)
	err = rpc.CallNoAuth("rawtransaction", ret, params)
	if err != nil {
		return "", "", err
	}
	if ret.ValidationError != "" {
		return "", "", errors.New(ret.ValidationError)
	}
	st := time.Now()
	if ret.ValidationWarning != "" { //可能会失败的情况
		for { //等待交易确认区块
			if time.Since(st) > 45*time.Second {
				return ret.Signature, ret.Raw, nil
			}
			height, _ := rpc.GetBlockCount()
			if ret.ScheduledBlock <= height {
				break
			}
			time.Sleep(time.Second)
		}
		block, err := rpc.Block(ret.ScheduledBlock)
		if err != nil {
			return ret.Signature, ret.Raw, nil
		}
		for _, tx := range block.Transactions { //区块是否含有此交易
			if tx.Signature == ret.Signature {
				return ret.Signature, ret.Raw, nil
			}
		}
		return ret.Signature, ret.Raw, errors.New(ret.ValidationWarning)

	}
	return ret.Signature, ret.Raw, nil
}

type BalanceRet struct {
	Balance    uint64 `json:"balance"`
	ListLength int    `json:"list_length"`
}

func (rpc *RpcClient) GetBalance(addr string) (uint64, error) {
	params := make(map[string]interface{}, 0)
	params["nyzo_string"] = addr
	ret := new(BalanceRet)
	err := rpc.CallNoAuth("balance", ret, params)
	return ret.Balance, err

}

type BlockRet struct {
	Hash         string         `json:"hash"`
	Height       int64          `json:"height"`
	Transactions []*Transaction `json:"transactions"`
}
type Transaction struct {
	Type      string `json:"type_enum"`
	From      string `json:"sender_nyzo_string"`
	To        string `json:"receiver_nyzo_string"`
	Memo      string `json:"sender_data"`
	Amount    int64
	Fee       int64
	Id        string `json:"id"`
	Signature string `json:"signature"`
}

func (rpc *RpcClient) Block(height int64) (ret *BlockRet, err error) {
	params := make(map[string]interface{}, 0)
	params["height"] = height
	ret = new(BlockRet)
	err = rpc.CallNoAuth("block", ret, params)
	if err != nil {
		return nil, err
	}
	return ret, err
}
