package alaya
import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)
func (client *RpcClient)SendRawTransaction(rawTx string) error{
	return client.CallNoAuth("platon_sendRawTransaction",nil,rawTx)
}

func (client *RpcClient)PendingNonceAt(addr string) (uint64,error){
	var result hexutil.Uint64
	err:= client.CallNoAuth("platon_getTransactionCount",&result,addr,"pending")
	return uint64(result),err
}

func (client *RpcClient)SuggestGasPrice() (*big.Int, error) {
	var hex hexutil.Big
	err:= client.CallNoAuth("platon_gasPrice",&hex)
	return (*big.Int)(&hex), err
}
func (client *RpcClient) BalanceAt(addr string) (*big.Int, error) {
	var result hexutil.Big
	err := client.CallNoAuth(  "platon_getBalance", &result,addr, "pending")
	return (*big.Int)(&result), err
}