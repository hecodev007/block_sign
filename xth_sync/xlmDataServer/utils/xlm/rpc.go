package xlm

import (
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
)

type RpcClient struct {
	*horizonclient.Client
}

func NewRpcClient(url, s, k string) *RpcClient {
	rpc := &RpcClient{
		Client: &horizonclient.Client{
			HorizonURL: url,
			HTTP:       http.DefaultClient,

		},
	}
	if url == ""  || url == "https://horizon.stellar.org/"{
		rpc.Client =horizonclient.DefaultPublicNetClient
	}
	return rpc
}
func (rpc *RpcClient) GetBlockCount() (int64, error) {
	resp, err := http.Get(rpc.Client.HorizonURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return gjson.Get(string(body), "history_latest_ledger").Int(), nil
}
func (rpc *RpcClient) GetBlockByHeight(height int64) (ledger hProtocol.Ledger, err error) {
	return rpc.LedgerDetail(uint32(height))
}
func (rpc *RpcClient) GetBlockTransactions(height int64) (txpages *hProtocol.TransactionsPage, err error) {
	num := uint(200)
	ret ,err := rpc.Client.Transactions(horizonclient.TransactionRequest{ForLedger: uint(height), Limit: num, IncludeFailed: false})
	if err != nil {
		return
	}
	//log.Info(xutils.String(horizonclient.TransactionRequest{ForLedger: uint(height), Limit: num, IncludeFailed: false}))
	//log.Info(ret.Links.Next.Href)
	tmpret := ret
next:
	//log.Info(tmpret.Links.Self.Href)
	//log.Info(tmpret.Links.Next.Href)
	if len(tmpret.Embedded.Records) == 200 {
		link := strings.Split(tmpret.Links.Next.Href,"cursor=")
		if len(link) !=2 {
			return
		}
		//log.Info(link[1])
		cursors := strings.Split(link[1],"&limit")
		if len(cursors) !=2 {
			return
		}
		//log.Info(cursors[0])
		//log.Info(horizonclient.TransactionRequest{ ForLedger: uint(height),Cursor:cursors[0],Limit: num, IncludeFailed: false}.BuildURL())
		//ForLedger: uint(height),
		tmpret ,err = rpc.Client.Transactions(horizonclient.TransactionRequest{ ForLedger: uint(height),Cursor:cursors[0],Limit: num, IncludeFailed: false})
		//log.Info("rpc.Client.Transactions end")
		//log.Info(tmpret.Links.Next)
		if err != nil {
			return
		}
		ret.Embedded.Records = append(ret.Embedded.Records,tmpret.Embedded.Records...)
		goto next
	}
	return &ret,err
}
func (rpc *RpcClient) GetRawTransaction(txid string) (hProtocol.Transaction, error) {
	return rpc.Client.TransactionDetail(txid)
}

func (rpc *RpcClient) GetOpsByHeight(height int64) {
	rpc.Client.Payments(horizonclient.OperationRequest{ForLedger: uint(height), Limit: 200})
}
func (rpc *RpcClient) UpdateBlocks(start,end int64,addresses []string)(map[int64]bool,error){
	ret := make(map[int64]bool)

	for _,addr := range addresses{
		txs,err := rpc.Transactions(horizonclient.TransactionRequest{ForAccount: strings.ToUpper(addr),Order: "desc",Limit: 200})
		if err != nil {
			return nil, err
		}
		for _,v := range txs.Embedded.Records{
			ret[int64(v.Ledger)] = true
		}
	}
	return ret,nil
}