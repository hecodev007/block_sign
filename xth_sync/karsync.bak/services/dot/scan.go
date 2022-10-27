package dot

import (
	"encoding/json"
	"karsync/common"
	"karsync/common/conf"
	"karsync/common/log"
	dao "karsync/models/po/dot"
	"karsync/services"
	rpc "karsync/utils/kar"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.Client
	lock  *sync.Mutex
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	client, err := rpc.NewClient(node.Url)
	if err != nil {
		panic(err)
	}

	return &Scanner{
		Client: client,
		lock:   &sync.Mutex{},
		conf:   conf.Sync,
		watch:  watch,
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
	if err != nil {
		task := &ProcTask{
			BestHeight: bestHeight,
			Block: &dao.BlockInfo{
				Height: height,
				Hash:   "",
				Time:   time.Now(),
			},
		}
		log.Info(err.Error())
		return task, nil
		//return task, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height,
			Hash:              block.BlockHash,
			Previousblockhash: block.ParentHash,
			Transactions:      len(block.Extrinsics),
			Confirmations:     bestHeight - block.Height + 1,
			Time:              time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	for _, tx := range block.Extrinsics {
		if tx.Status != "success" {
			continue
		}
		tmpDaoTx := new(dao.BlockTx)
		tmpDaoTx.Height = block.Height
		tmpDaoTx.Hash = block.BlockHash
		tmpDaoTx.Txid = tx.Txid
		tmpDaoTx.Fromaccount = tx.FromAddress
		tmpDaoTx.Toaccount = tx.ToAddress
		fee, err := decimal.NewFromString(tx.Fee)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		tmpDaoTx.SysFee = fee.Shift(-12).String()
		amount, err := decimal.NewFromString(tx.Amount)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		tmpDaoTx.Amount = amount.Shift(-12).String()
		task.TxInfos = append(task.TxInfos, tmpDaoTx)
	}
	_ = starttime
	//log.Infof("scanBlock %d ,used time : %f
	return task, nil
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
