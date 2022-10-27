package wit

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"witDataServer/common"
	"witDataServer/common/conf"
	"witDataServer/common/log"
	dao "witDataServer/models/po/wit"
	"witDataServer/services"
	"witDataServer/utils"
	rpc "witDataServer/utils/wit"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.RpcClient
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
	dao.TxVoutRollBack(height)
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(3395436)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBlockCount() //获取到的是区块个数
	return count - 1, err
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
	//starttime := time.Now()
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		os.Exit(0)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}
	//log.Info(Json(block))
	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))
	//if has, err := dao.BlockHashExist(block.Hash); err != nil {
	//	return nil, fmt.Errorf("database err")
	//} else if has {
	//	return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, 1)
	//}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:        block.Height,
			Hash:          block.Hash,
			Transactions:  len(block.TxnsHashes.ValueTransfer),
			Confirmations: bestHeight - height + 1, //block.Confirmations
			Createtime:    time.Now(),
		},
	}

	//并发处理区块内的交易

	lock := new(sync.Mutex)
	workpool := utils.NewWorkPool(4) //一次性发太多请求会让节点窒息
	for index, txid := range block.TxnsHashes.ValueTransfer {
		workpool.Add(1)
		go func(txid string, index int) {
			defer workpool.Dec()
			//过滤非转账交易
			//log.Info("getrawtransaction:", txid, index, len(block.Txs))
			tx := block.Txns.ValueTransferTxns[index]
			//log.Info("getrawtransaction success:", txid, index, len(block.Txs))
			if txInfo, err := parseBlockRawTX(s.RpcClient, txid, tx, block.Hash, height); err != nil {
				log.Info(err.Error())
			} else if txInfo != nil {
				lock.Lock() //append并发不安全
				defer lock.Unlock()
				task.TxInfos = append(task.TxInfos, txInfo)
			}
			//log.Info("TX index:", workpool.Running, index)
		}(txid, index)
	}
	workpool.Wait()
	//if txjson, err := json.Marshal(task); err == nil {
	//	log.Infof("block:%v", string(txjson))
	//} else {
	//	log.Warn(err.Error())
	//}
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, txhash string, tx *rpc.ValueTransferTxn, blockhash string, height int64) (*TxInfo, error) {
	var vouts []*dao.BlockTxVout
	var vins []*dao.BlockTxVin
	if tx == nil {
		return nil, nil
	}

	blocktx := &dao.BlockTx{
		Txid:       txhash,
		Height:     height,
		Blockhash:  blockhash,
		Voutcount:  len(tx.Body.Outputs),
		Vincount:   len(tx.Body.Inputs),
		Createtime: time.Now(),
		Fee:        "0",
	}
	var outAmount, inAmount int64
	for ountn, vout := range tx.Body.Outputs {
		outAmount += vout.Value
		blocktxvout := &dao.BlockTxVout{
			Height:     height,
			Txid:       blocktx.Txid,
			VoutN:      ountn,
			Blockhash:  blocktx.Blockhash,
			Value:      decimal.New(vout.Value, -9).String(),
			Createtime: time.Now(),
			Address:    vout.Address,
		}
		if vout.TimeLock != 0 {
			return nil, errors.New("锁定的交易")
		}
		vouts = append(vouts, blocktxvout)
	}

	for _, vin := range tx.Body.Inputs {
		outpointer := strings.Split(vin.OutputPointer, ":")
		txid := outpointer[0]
		vinN, err := strconv.Atoi(outpointer[1])
		if err != nil {
			panic(txhash + "vin 解析失败")
		}
	GetTransaction:
		vintx, err := RpcClient.GetTransaction(txid)
		if err != nil {
			log.Info(vin.OutputPointer, err.Error())
			time.Sleep(time.Second * 10)
			goto GetTransaction
		}
		//log.Info(vin.OutputPointer)
		var vout *rpc.Output
		if vintx.Transaction.ValueTransferTxn != nil {
			vout = vintx.Transaction.ValueTransferTxn.Body.Outputs[vinN]
		} else if vintx.Transaction.Mint != nil {
			vout = vintx.Transaction.Mint.Outputs[vinN]
		} else {
			panic(vin.OutputPointer + " output 解析失败")
		}
		inAmount += vout.Value

		//获得vin对应的vout
		blocktxvin := &dao.BlockTxVin{
			Blockhash:  vintx.BlockHash,
			Value:      decimal.New(vout.Value, -9).String(),
			Createtime: time.Now(),
			Txid:       txid,
			VoutN:      vinN,
			SpendTxid:  blocktx.Txid,
			Address:    vout.Address,
		}

		vins = append(vins, blocktxvin)
	}
	fee := inAmount - outAmount
	if fee > 0 {
		blocktx.Fee = decimal.New(fee, -9).String()
	}
	return &TxInfo{
		Tx:    blocktx,
		Vouts: vouts,
		Vins:  vins,
	}, nil
}
