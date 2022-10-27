package nyzo

import (
	"errors"
)

type InfoRet struct {
	RetentionEdge int64 `json:"frozen_edge"`
}
func (rpc *RpcClient) GetBlockCount()(int64,error){
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
	Signature string `json:"signature"`
	Raw       string `json:"raw"`

	ScheduledBlock  int64  `json:"scheduled_block"`
	ValidationError string `json:"validation_error"`
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
	return ret.Signature, ret.Raw, nil
}

type BlockRet struct {
	Hash string `json:"hash"`
	Height int64 `json:"height"`
	Transactions []*Transaction `json:"transactions"`
}
type Transaction struct{
	Type string `json:"type_enum"`
	From string `json:"sender_nyzo_string"`
	To string `json:"receiver_nyzo_string"`
	Memo string `json:"sender_data"`
	Amount int64
	Fee int64
	Id string `json:"id"`
	Signature string `json:"signature"`
}
func (rpc *RpcClient) Block(height int64) (ret *BlockRet,err error){
	params := make(map[string]interface{}, 0)
	params["height"] = height
	ret = new(BlockRet)
	err = rpc.CallNoAuth("block", ret, params)
	if err != nil {
		return nil, err
	}
	return ret,err
}

func (rpc *RpcClient) GetBlockByHeight(height int64) (ret *BlockRet,err error){
	ret,err = rpc.Block(height)

	if err != nil && (err.Error() == "java.lang.NullPointerException\n" || err.Error() == "unknown block"){
		ret = &BlockRet{
			Height: height,
			Hash :err.Error(),
			Transactions: make([]*Transaction,0),
		}
		//log.Infof("java.lang.NullPointerException %v",height)
		return ret,nil
	}
	return ret,err
}
