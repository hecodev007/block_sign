package stellar

import (
	"github.com/shopspring/decimal"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/operations"
	"log"
)

//包装的RPC-HTTP 客户端
type RpcClient struct {

}

// New create new rpc RpcClient with given url
func NewRpcClient(options ...string) *RpcClient {
	return new(RpcClient)
}

func (rpc *RpcClient) GetBlockCount()(int64,error) {
	client :=horizonclient.DefaultPublicNetClient
	lds,err := client.Ledgers(horizonclient.LedgerRequest{Order: "desc",Limit:1})
	if err != nil {
		return 0,err
	}
	return int64(lds.Embedded.Records[0].Sequence),nil
}

type BlockRet struct {
	Hash string `json:"hash"`
	Height int64 `json:"height"`
	Transactions []*Transaction `json:"transactions"`
}
type Transaction struct{
	Txid string
	From string
	To string
	Memo string
	Amount string
	Fee string
}
func (rpc *RpcClient) GetBlockByHeight(height int64) (ret *BlockRet,err error){
	client :=horizonclient.DefaultPublicNetClient
	lg,err := client.LedgerDetail(uint32(height))
	if err != nil {
		return nil, err
	}
	pms,err := client.Payments(horizonclient.OperationRequest{ForLedger: uint(height)})
	if err != nil {
		return nil, err
	}
	ret = new(BlockRet)
	ret.Hash = lg.Hash
	ret.Height = height
	for _,v := range pms.Embedded.Records{
		if !v.IsTransactionSuccessful(){
			continue
		}
		payment := v.(operations.Payment)
		tx := new(Transaction)
		tx.Txid = payment.TransactionHash
		tx.From = payment.From
		tx.To = payment.To
		tx.Amount = payment.Amount
		txdedail,err := client.TransactionDetail(tx.Txid)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		tx.Fee = decimal.NewFromInt(txdedail.FeeCharged).Shift(-7).String()
		tx.Memo = txdedail.Memo
		ret.Transactions = append(ret.Transactions,tx)
	}
	return ret,nil
}



