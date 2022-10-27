package xrp

import (
	"encoding/json"
	"marsDataServer/common/log"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

//获取高度
func (rpc *RpcClient) BlockHeight() (int64, error) {
	rawResp, err := rpc.RawCall("ledger_current", "")
	if err != nil {
		return 0, err
	}
	resp := new(LedgerCurrentResponse)
	if err = json.Unmarshal(rawResp, resp); err != nil {
		return 0, err
	}
	if resp.Status != "success" {
		return 0, resp.XrpError
	}
	return resp.LedgerCurrentIndex, nil
}

//获取交易
func (rpc *RpcClient) GetTracsaction(h string) (*Transaction, error) {
	params := struct {
		Transaction string `json:"transaction"`
		Binary      bool   `json:"binary"`
	}{Transaction: h, Binary: false}
	ret := new(Transaction)
	if err := rpc.CallWithAuth("tx", rpc.Credentials, ret, params); err != nil {
		return nil, err
	}
	if ret.Status != "success" {
		return nil, ret.XrpError
	}
	return ret, nil
}

//获取block
func (rpc *RpcClient) GetBlock(height int64) (*Block, error) {
	params := struct {
		LedgerIndex  int64 `json:"ledger_index"`
		Accounts     bool  `json:"accounts"`
		Full         bool  `json:"full"`
		Transactions bool  `json:"transactions"`
		Expand       bool  `json:"expand"`
		OwnerFunds   bool  `json:"owner_funds"`
	}{LedgerIndex: height, Transactions: true}
	ret := new(GetBlockResult)
	if err := rpc.CallWithAuth("ledger", rpc.Credentials, ret, params); err != nil {
		return nil, err
	}
	if ret.Status != "success" {
		return nil, ret.XrpError
	}
	ret.Block.Height = ret.BlockHeight
	ext, err := time.Parse("2006-Jan-02 15:04:05.000000000 MST", ret.Block.CloseTimeHuman)
	if err == nil {
		ret.Block.Time = ext
	}
	return ret.Block, nil
}

//带tx的block
func (rpc *RpcClient) GetFullBlock(height int64) (*FullBlock, error) {
	block, err := rpc.GetBlock(height)
	if err != nil {
		return nil, err
	}
	ret := new(FullBlock)
	ret.Block = *block
	ret.Transacitons = make([]*Transaction, len(ret.Transacitons))
	group := new(sync.WaitGroup)
	for index, txhash := range block.Transactions {
		group.Add(1)
		go func(index int, txhash string) {
			defer group.Done()
		retry:
			tx, err := rpc.GetTracsaction(txhash)
			if err != nil {
				log.Warn(err.Error(), txhash)
				time.Sleep(time.Second * 3)
				goto retry
			}
			ret.Transacitons = append(ret.Transacitons, tx)
		}(index, txhash)
	}
	group.Wait()

	return ret, nil
}

func (rpc *RpcClient) GetBalance(addr string) (value int64, decimalValue string, blockHeight int64, err error) {
	params := struct {
		Account     string `json:"account"`
		Strict      bool   `json:"strict"`
		LedgerIndex string `json:"ledger_index"`
		Queue       bool   `json:"queue"`
	}{Account: addr, Strict: true, LedgerIndex: "current", Queue: true}

	ret := new(BalanceResponse)
	if err := rpc.CallWithAuth("account_info", rpc.Credentials, ret, params); err != nil {
		return 0, "", 0, err
	}
	//rj, _ := json.Marshal(ret)
	//log.Info(string(rj))
	decimalBalance, err := decimal.NewFromString(ret.AccountData.Balance)
	if err != nil {
		return 0, "", 0, err
	}
	value = decimalBalance.IntPart()
	decimalValue = decimalBalance.Shift(-6).String()
	blockHeight = ret.LedgerCurrentIndex
	return
}
