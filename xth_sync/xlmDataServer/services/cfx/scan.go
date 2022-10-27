package cfx

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
	"xlmDataServer/common"
	"xlmDataServer/common/conf"
	"xlmDataServer/common/log"
	dao "xlmDataServer/models/po/cfx"
	"xlmDataServer/services"
	"xlmDataServer/utils"
	rpc "xlmDataServer/utils/xlm"

	"github.com/shopspring/decimal"

	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/protocols/horizon/operations"

	"github.com/stellar/go/clients/horizonclient"
)

type Scanner struct {
	*rpc.RpcClient
	conf  conf.SyncConfig
	watch *services.WatchControl
	Blocks map[int64]bool
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
		watch:     watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
}
//看下哪些高度有账户的交易
func (s *Scanner)UpdateAcountTxs(start,end int64) error {
	addrs := make([]string,0)
	for k,_ := range s.watch.WatchAddrs{
		addrs = append(addrs,k)
	}
	blocks,err := s.UpdateBlocks(start,end,addrs)
	if err != nil {
		return err
	}
	s.Blocks = blocks
	//s.Client.AccountDetail()
	return nil
}
func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(781001)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBlockCount() //获取到的是区块个数
	return count - 12, err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	//if height != 2071702{
	//	time.Sleep(time.Second*10000)
	//	return nil, errors.New("123")
	//}
	//starttime := time.Now()
	//log.Info("scanBlock", height, bestHeight)
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Info("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 2)
		goto retryGetBlockByHeight
	}

	//log.Info(String(block))
	if height != int64(block.Sequence) {
		panic("")
	}
	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            int64(block.Sequence),
			Hash:              block.Hash,
			Previousblockhash: block.PrevHash,
			Nextblockhash:     "",
			Transactions:      int(block.SuccessfulTransactionCount),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}
	if _,ok := s.Blocks[height];!ok{
		return task,nil
	}
	//log.Info(height)
	//并发处理区块内的交易
GetBlockTransactions:
	Transactions, err := s.GetBlockTransactions(height)
	if err != nil {
		log.Info(height, err.Error())
		time.Sleep(time.Second * 3)
		goto GetBlockTransactions
	}
	lock := new(sync.Mutex)
	workpool := utils.NewWorkPool(2)
	for _, tx := range Transactions.Embedded.Records {
		if tx.OperationCount == 0 {
			continue
		}
		workpool.Incr()
		go func(tx *horizon.Transaction) {
			defer workpool.Dec()
		getPayments:
			payments, err := s.Client.Payments(horizonclient.OperationRequest{ForTransaction: tx.Hash})
			if err != nil {
				if height%20 == 0 {
					log.Info(height, err.Error())
				}
				time.Sleep(time.Second * 2)
				goto getPayments
			}

			if txInfo, err := parseBlockRawTX(tx, payments.Embedded.Records, block.Hash); err != nil {
				log.Info(tx.Hash, err.Error())
			} else if txInfo != nil {

				lock.Lock() //append并发不安全
				task.TxInfos = append(task.TxInfos, txInfo)
				lock.Unlock()
			}
		}(tx)
	}
	workpool.Wait()
	//ts, _ := json.Marshal(task)
	//log.Info("task", height, string(ts), "task")
	return task, nil
}

//解析交易
func parseBlockRawTX(tx *horizon.Transaction, ops []operations.Operation, blockhash string) (blocktx *dao.BlockTx, err error) {
	if !tx.Successful {
		return nil, errors.New("tx.status != success")
	}

	for _, op := range ops {
		if op.GetType() != "payment" && op.GetType() != "create_account" {
			continue
		}

		req, _ := decimal.NewFromString(tx.AccountSequence)
		blocktx = &dao.BlockTx{
			CoinName:        conf.Cfg.Sync.Name,
			Txid:            tx.Hash,
			ContractAddress: "",
			FromAddress:     tx.Account,
			//ToAddress:"",
			//Amount: "0",
			BlockHeight: int64(tx.Ledger),
			BlockHash:   blockhash,
			Fee:         decimal.NewFromInt(tx.MaxFee).Shift(-7).String(),
			Nonce:       req.IntPart(),
			Memo:        tx.Memo,
			Timestamp:   time.Now(),
		}

		if op.GetType() == "payment" {
			v, ok := op.(operations.Payment)
			if !ok {
				panic("payment 类型转换错误,需要处理")
			}
			token := ""
			if v.Asset.Type != "native" {
				token = v.Asset.Code+"-"+v.Asset.Issuer
			}

			if token == "-"{
				continue
			}

			blocktx.ToAddress = v.To
			blocktx.Amount = v.Amount
			blocktx.ContractAddress = token
		} else {
			v, ok := op.(operations.CreateAccount)
			if !ok {
				panic("CreateAccount 类型转换错误,需要处理")
			}
			//str, _ := json.Marshal(v)
			//log.Info(string(str))
			//panic("")
			blocktx.ToAddress = v.Account
			blocktx.Amount = v.StartingBalance
		}
		return
	}

	return blocktx, nil

}

func String(d interface{})string{
	str,_ := json.Marshal(d)
	return string(str)
}