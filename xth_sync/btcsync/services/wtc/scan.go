package wtc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"btcsync/common"
	"btcsync/common/conf"
	"btcsync/common/log"
	dao "btcsync/models/po/yotta"
	"btcsync/services"
	"btcsync/utils"
	rpc "btcsync/utils/wtc"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.RpcClient
	lock  *sync.Mutex
	conf  conf.SyncConfig
	Watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
		Watch:     watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	_, err := dao.BlockRollBack(height)
	if err != nil {
		panic(err.Error())
	}
	_, err = dao.TxRollBack(height)
	if err != nil {
		panic(err.Error())
	}
}

func (s *Scanner) Init() error {
	if s.conf.EnableRollback {
		s.Rollback(s.conf.RollHeight)
	}
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(60612620)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.BlockNumber() //获取到的是区块个数
	return count, err
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
	starttime := time.Now()
	//log.Infof("scanBlock %d ", height)
	getblocknumber:
	block, err := s.BlockByNumber(height)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}
	if block.Number.ToInt().Int64() == 0 {
		goto getblocknumber
	}
	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Number.ToInt().Int64(),
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Timestamp:         time.Unix(int64(block.Timestamp), 0),
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1,
			Createtime:        time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息

	for i, _ := range block.Transactions {
		workpool.Incr()
		go func(tx *rpc.Transaction, task *ProcTask) {
			defer workpool.Dec()
			s.batchParseTx(tx, task)
		}(block.Transactions[i], task)

	}
	workpool.Wait()
	_ = starttime
	//log.Infof("scanBlock %d ,used time : %f
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *rpc.Transaction, task *ProcTask) {

	blockTxs, err := parseBlockTX(s.Watch, s.RpcClient, tx, task.Block)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	} else {
		//log.Info(err.Error())
	}
}

// 解析交易
func parseBlockTX(watch *services.WatchControl, rpc *rpc.RpcClient, tx *rpc.Transaction, block *dao.BlockInfo) ([]*dao.BlockTx, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	var blocktxs = make([]*dao.BlockTx, 0)

	if watch.IsContractExist(tx.To) || len(tx.Input) > 75 && watch.IsWatchAddressExist("0x"+tx.Input[34:74]) {
		//if !strings.HasPrefix(tx.Input, "0xa9059cbb") || len(tx.Input)<138{
		//	return nil, errors.New("非合约转账交易")
		//}
		//tmpTo := "0x" + tx.Input[34:74]
		//if !watch.IsWatchAddressExist(tx.From) && !watch.IsWatchAddressExist(tmpTo) {
		//	return nil, errors.New("没有监听的地址")
		//}
	TransactionReceipt:
		receipt, err := rpc.TransactionReceipt(tx.Hash)
		if err != nil {
			log.Error(err.Error())
			time.Sleep(time.Second * 10)
			goto TransactionReceipt
		}

		if receipt.Status != 1 {
			return nil, errors.New("失败的合约转账交易")
		}
		if len(receipt.Logs) != 1 || len(receipt.Logs[0].Topics) <= 0 || receipt.Logs[0].Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			return nil, errors.New("交易reciept分析错误,找管理员处理")
		}
		ContractAddress := receipt.Logs[0].Address
		if !watch.IsContractExist(ContractAddress) {
			return nil, errors.New("不是监控的合约转账")
		}
		contract, _ := watch.GetContract(ContractAddress)
		from := "0x" + receipt.Logs[0].Topics[1][26:66]
		to := "0x" + receipt.Logs[0].Topics[2][26:66]
		if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) {
			return nil, errors.New("没有监听的地址")
		}
		d, err := hex.DecodeString(strings.TrimPrefix(receipt.Logs[0].Data, "0x"))
		if err != nil {
			panic(tx.Hash + err.Error())
		}
		amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(d), 0-int32(contract.Decimal))
		blocktx := dao.BlockTx{
			CoinName:        contract.Name,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: ContractAddress,
			FromAddress:     from,
			ToAddress:       to,
			Amount:          amount,
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice.ToInt(), receipt.GasUsed.ToInt()), -18),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return blocktxs, nil
	}

	//普通转账处理
	if !watch.IsWatchAddressExist(tx.From) && !watch.IsWatchAddressExist(tx.To) {
		return nil, errors.New("没有监听的地址")
	}
TransactionReceipt2:
	receipt, err := rpc.TransactionReceipt(tx.Hash)
	if err != nil {
		log.Error(err.Error())
		time.Sleep(time.Second * 10)
		goto TransactionReceipt2
	}
	blocktx := dao.BlockTx{
		CoinName:    conf.Cfg.Name,
		Txid:        tx.Hash,
		BlockHeight: block.Height,
		BlockHash:   block.Hash,
		FromAddress: tx.From,
		ToAddress:   tx.To,
		Amount:      decimal.NewFromBigInt(tx.Value.ToInt(), -18),
		Fee:         decimal.NewFromBigInt(big.NewInt(0).Mul(receipt.GasUsed.ToInt(), tx.GasPrice.ToInt()), -18),
		Status:      "success",
		Timestamp:   block.Timestamp,
	}
	blocktxs = append(blocktxs, &blocktx)
	return blocktxs, nil
}

// 解析交易
func parseBlockTxInternal(watch *services.WatchControl, rpc *rpc.RpcClient, tx *rpc.Transaction) ([]*dao.BlockTx, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	var blocktxs = make([]*dao.BlockTx, 0)
TransactionReceipt:
	receipt, err := rpc.TransactionReceipt(tx.Hash)
	if err != nil {
		log.Error(err.Error())
		time.Sleep(time.Second * 10)
		goto TransactionReceipt
	}

	if receipt.Status != 1 {
		return nil, errors.New("失败的交易")
	}
	log.Info(String(tx))
	if tx.Value.ToInt().Cmp(big.NewInt(0)) != 0 {
		log.Info(tx.To)
		if watch.IsWatchAddressExist(tx.From) || watch.IsWatchAddressExist(tx.To) {
			blocktx := dao.BlockTx{
				CoinName:    conf.Cfg.Name,
				Txid:        tx.Hash,
				BlockHeight: int64(tx.BlockNumber),
				BlockHash:   tx.BlockHash,
				FromAddress: tx.From,
				ToAddress:   tx.To,
				Amount:      decimal.NewFromBigInt(tx.Value.ToInt(), -18),
				Fee:         decimal.NewFromBigInt(big.NewInt(0).Mul(receipt.GasUsed.ToInt(), tx.GasPrice.ToInt()), -18),
				Status:      "success",
				Timestamp:   time.Now(),
			}
			blocktxs = append(blocktxs, &blocktx)
		}
	}
	log.Info(String(receipt.Logs))
	for _, tlog := range receipt.Logs {
		if !watch.IsContractExist(tlog.Address) {
			log.Info(tlog.Address)
			continue
		}
		if len(tlog.Topics) != 3 || tlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			continue
		}
		from := "0x" + receipt.Logs[0].Topics[1][26:66]
		to := "0x" + receipt.Logs[0].Topics[2][26:66]
		//log.Info(receipt.Logs[0].Topics[1], from, to)
		//log.Info(from, to, tx.To)
		if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) {
			log.Info(from, to)
			continue
		}
		contract, _ := watch.GetContract(tx.To)
		d, err := hex.DecodeString(strings.TrimPrefix(tlog.Data, "0x"))
		if err != nil {
			panic(tx.Hash + err.Error())
		}
		amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(d), 0-int32(contract.Decimal))
		blocktx := dao.BlockTx{
			CoinName:        contract.Name,
			Txid:            tx.Hash,
			BlockHeight:     int64(tx.BlockNumber),
			ContractAddress: tlog.Address,
			FromAddress:     from,
			ToAddress:       to,
			Amount:          amount,
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice.ToInt(), receipt.GasUsed.ToInt()), -18),
			BlockHash:       tx.BlockHash,
			Status:          "success",
			Timestamp:       time.Now(),
		}
		blocktxs = append(blocktxs, &blocktx)
	}
	return blocktxs, nil
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
