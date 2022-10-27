package dot

import (
	"dotsync/common"
	"dotsync/common/conf"
	"dotsync/common/log"
	dao "dotsync/models/po/dot"
	"dotsync/services"
	rpc "dotsync/utils/dot"
	"encoding/json"
	"sync"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	lock  *sync.Mutex
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	rpcClient,err :=rpc.NewRpcClient(node.Node, node.ScanApi, node.ScanKey)
	if err != nil {
		panic(err.Error())
	}
	return &Scanner{
		RpcClient: rpcClient,
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
		watch:     watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
}

func (s *Scanner) Init() error {
	if conf.Cfg.Sync.EnableRollback {
		s.Rollback(conf.Cfg.Sync.RollHeight)
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
	count, err := s.GetBestHeight() //获取到的是区块个数
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
	block, err := s.GetBlockByNum(height)
	if err != nil || block == nil {
		task := &ProcTask{
			BestHeight: bestHeight,
			Block: &dao.BlockInfo{
				Height:           height,
				Hash:              "",
				Time:              time.Now(),
			},
		}
		if err != nil {
			log.Info(err.Error())
		} else {
			log.Info(height,"区块没找到")
		}
		return task,nil
		//return task, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	//log.Info(xutils.String(block))
	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.BlockNum,
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Transactions:      len(block.Extrinsics),
			Confirmations:     bestHeight - block.BlockNum + 1,
			Time:              time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	for i,extrinsic := range block.Extrinsics{
		if !extrinsic.Success {
			continue
		}
		tx,err := block.ToTransaction(i)
		if err != nil{
			log.Info(i,String(block))
			log.Error(err.Error())
			panic("")
			continue
		}
		if tx == nil {
			continue
		}


		tmpDaoTx :=  &dao.BlockTx{
			Height: tx.BlockHeight,
			Hash:tx.BlockHash,
			Txid:tx.Txid,
			Fromaccount: tx.From,
			Toaccount: tx.To,
			Amount:tx.Value,
			SysFee :tx.Fee,
		}
		task.TxInfos = append(task.TxInfos,tmpDaoTx)
	}
	_ = starttime
	//log.Infof("scanBlock %d ,used time : %f
	return task, nil
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
