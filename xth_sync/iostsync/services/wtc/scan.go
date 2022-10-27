package wtc

import (
	"encoding/json"
	"errors"
	"fmt"
	"iostsync/common"
	"iostsync/common/conf"
	"iostsync/common/log"
	dao "iostsync/models/po/yotta"
	"iostsync/services"
	"iostsync/utils"
	rpc "iostsync/utils/iost"
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
	block, err := s.BlockByNumber(height)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}
	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Number.IntPart(),
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Timestamp:         time.UnixMilli(int64(block.Time.IntPart())),
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1,
			Createtime:        time.Now(),
		},
	}
	//log.Info(task.Block.Height, len(task.TxInfos), xutils.String(task))

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	workpool := utils.NewWorkPool(1) //一次性发太多请求会让节点窒息

	for i, _ := range block.Transactions {
		workpool.Incr()
		go func(tx *rpc.Transaction, task *ProcTask) {
			defer workpool.Dec()
			s.batchParseTx(tx, task)
		}(block.Transactions[i], task)

	}
	workpool.Wait()
	_ = starttime
	//log.Info(xutils.String(task.TxInfos))
	//panic("")
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

	for _, action := range tx.Actions {

		//非转账交易
		if action.Contract != "token.iost" {
			continue
		}
		if action.ActionName != "transfer" {
			log.Info("transfer")
			continue
		}
		var datas []string
		err := json.Unmarshal([]byte(action.Data), &datas)
		if err != nil {
			panic(err.Error() + (action.Data))
		}
		//合约地址,from账户,to账户,额度,memo
		//log.Info(datas[0], datas[1], datas[2], datas[3], datas[4])
		//参数个数必须为5
		if len(datas) != 5 {
			panic("")
		}
		if datas[4] != "" {
			//panic(tx.Hash)
		}
		if !watch.IsContractExist(datas[0]) {
			//log.Info("IsContractExist")
			continue
		}
		//必须是监控的地址
		if !watch.IsWatchAddressExist(datas[1]) && !watch.IsWatchAddressExist(datas[2]) {
			//log.Info(datas[1], datas[2])
			continue
		}
		//延时必须是0
		if tx.Delay.IntPart() != 0 {
			//log.Info(tx.Delay.IntPart())
			continue
		}
		//log.Info(String(tx.TxReceipt))
		if tx.TxReceipt.StatusCode != "SUCCESS" {
			//log.Info("faild" + tx.TxReceipt.StatusCode)
			continue
		}
		amount, err := decimal.NewFromString(datas[3])
		if err != nil {
			panic(err.Error())
		}
		contractinfo, _ := watch.GetContract(datas[0])
		blocktx := dao.BlockTx{
			CoinName:        contractinfo.Name,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: datas[0],
			FromAddress:     datas[1],
			ToAddress:       datas[2],
			Memo:            datas[4],
			Amount:          amount,
			Fee:             decimal.NewFromInt(0),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		//log.Info(String(blocktx))
		blocktxs = append(blocktxs, &blocktx)
	}
	//log.Info(xutils.String(blocktxs))
	return blocktxs, nil
}

// 解析交易
func parseBlockTxInternal(watch *services.WatchControl, rpc *rpc.RpcClient, txresp *rpc.TransactionResponse, blockhash string) (blocktxs []*dao.BlockTx, err error) {
	tx := txresp.Transaction
	if txresp == nil {
		return nil, fmt.Errorf("tx is null")
	}
	blocktxs = make([]*dao.BlockTx, 0)
	//log.Info(xutils.String(txresp))
	for _, action := range tx.Actions {
		//不是监控的合约
		//非转账交易
		if action.ActionName != "transfer" {
			err = errors.New("非转账交易")
			continue
		}
		var datas []string
		err = json.Unmarshal([]byte(action.Data), &datas)
		if err != nil {
			panic(err.Error())
		}
		//参数个数必须为5
		if len(datas) != 5 {
			panic("")
		}
		if !watch.IsContractExist(datas[0]) {
			err = errors.New("不含监控的合约:" + datas[0])
			continue
		}
		//必须是监控的地址
		if !watch.IsWatchAddressExist(datas[1]) && !watch.IsWatchAddressExist(datas[2]) {
			err = errors.New("不含监控的用户地址")
			continue
		}
		//延时必须是0
		if tx.Delay.IntPart() != 0 {
			err = errors.New("交易延时不为0")
			continue
		}
		if tx.TxReceipt.StatusCode != "SUCCESS" {
			err = errors.New("失败的交易")
			continue
		}
		amount, err := decimal.NewFromString(datas[3])
		if err != nil {
			panic(err.Error())
		}
		contractinfo, _ := watch.GetContract(datas[0])

		blocktx := dao.BlockTx{
			CoinName:        contractinfo.Name,
			Txid:            tx.Hash,
			BlockHeight:     txresp.BlockNumber.IntPart(),
			ContractAddress: datas[0],
			FromAddress:     datas[1],
			ToAddress:       datas[2],
			Memo:            datas[4],
			Amount:          amount,
			Fee:             decimal.NewFromInt(0),
			BlockHash:       blockhash,
			Status:          "success",
			Timestamp:       time.Now(),
		}
		blocktxs = append(blocktxs, &blocktx)
	}
	log.Info(err)
	return blocktxs, err
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
