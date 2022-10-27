package nas

import (
	"fmt"
	"github.com/shopspring/decimal"
	"rsksync/common"
	"rsksync/common/log"
	"rsksync/conf"
	dao "rsksync/models/po/nas"
	"rsksync/utils"
	"rsksync/utils/nas"
	"sync"
	"time"
)

type Scanner struct {
	*nas.NasHttpClient
	lock *sync.Mutex
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	c, err := nas.NewNasHttpClient(node.Url)
	if err != nil {
		log.Infof("NewNasClient err: %v", err)
		return nil
	}
	return &Scanner{
		NasHttpClient: c,
		lock:          &sync.Mutex{},
		conf:          conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.DeleteBlockInfo(height)

	dao.DeleteBlockTX(height)
}

//爬数据
func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {

}

func (s *Scanner) GetBestBlockHeight() (int64, error) {
	res, err := s.GetNebState()
	if err != nil {
		return 0, err
	}
	return int64(res.Height), nil
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.GetMaxBlockIndex()
}

//批量扫描多个区块
func (s *Scanner) BatchScanIrreverseBlocks(startHeight, endHeight, bestHeight int64) *sync.Map {
	starttime := time.Now()
	count := endHeight - startHeight
	taskmap := &sync.Map{}
	wg := &sync.WaitGroup{}

	wg.Add(int(count))
	for i := int64(0); i < count; i++ {
		height := startHeight + i
		go func(w *sync.WaitGroup) {
			if task, err := s.ScanIrreverseBlock(height, bestHeight); err == nil {
				taskmap.Store(height, task)
			}
			w.Done()
		}(wg)
	}
	wg.Wait()
	log.Debugf("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
	return taskmap
}

func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个区块
func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()

	block, err := s.GetBlockByHeight(uint64(height), true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := dao.GetBlockCountByHash(block.Hash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &NasProcTask{
		bestHeight: bestHeight,
		block: &dao.BlockInfo{
			Height:         int64(block.Height),
			Hash:           block.Hash,
			FrontBlockHash: block.ParentHash,
			Timestamp:      time.Unix(block.Timestamp, 0),
			Transactions:   len(block.Transactions),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}

	//处理区块内的交易
	if len(block.Transactions) > 0 {
		if s.conf.EnableGoroutine {
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Transactions))
			for _, tmp := range block.Transactions {
				tx := tmp
				go s.batchParseTx(tx, bestHeight, block.Timestamp, task, wg)
			}
			wg.Wait()
		} else {
			for _, tx := range block.Transactions {
				if blockTx, err := s.parseBlockTX(tx, bestHeight, block.Timestamp); err == nil {
					task.txInfos = append(task.txInfos, blockTx)
				}
			}
		}
	}

	log.Infof("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *nas.Transaction, bestHeight, blockTimestamp int64, task *NasProcTask, w *sync.WaitGroup) {
	defer w.Done()

	if blockTx, err := s.parseBlockTX(tx, bestHeight, blockTimestamp); err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *nas.Transaction, bestHeight, blockTimestamp int64) (*dao.BlockTX, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}

	//虽然失败了，但是保存一下区块高度,不然重启会造成二次重扫

	if tx.ChainId != 1 {
		return nil, fmt.Errorf("tx:%s,非主链交易", tx.Hash)
	}
	if tx.Status != 1 {
		return nil, fmt.Errorf("tx:%s,失败交易", tx.Hash)
	}

	blocktx := &dao.BlockTX{
		BlockHeight:   int64(tx.BlockHeight),
		BlockHash:     tx.BlockHash,
		Txid:          tx.Hash,
		FromAddress:   tx.From,
		Nonce:         tx.Nonce,
		GasUsed:       tx.GasUsed,
		GasPrice:      tx.GasPrice.Int64(),
		Type:          tx.Type,
		Data:          string(tx.Data),
		CoinName:      s.conf.Name,
		Decimal:       nas.WEI,
		Timestamp:     time.Unix(blockTimestamp, 0),
		ExecuteResult: tx.ExecuteResult,
		ExecuteError:  tx.ExecuteError,
		Status:        tx.Status,
	}

	switch tx.Type {
	case nas.TxCall:
		//log.Infof("parse : %s ,txdata : %s ",tx.Hash,tx.Data)
		calldata, err := nas.ParseCallData([]byte(tx.Data))
		if err != nil {
			return nil, fmt.Errorf("ParseCallData input : %s, err: %v", blocktx.Data, err)
		}

		if calldata.Function != "transfer" {
			return nil, fmt.Errorf("ParseCallData input : %s, err: %v", blocktx.Data, fmt.Errorf("don't know transfer fuction %s", calldata.Function))
		}

		addr, amt, err := nas.ParseTransferData([]byte(calldata.Args))
		if err != nil {
			return nil, fmt.Errorf("ParseTransferData Args : %s, err: %v", calldata.Args, err)
		}

		blocktx.Amount = amt
		blocktx.ToAddress = addr
		blocktx.ContractAddress = tx.To
		break
	case nas.TxDeploy:
		break
	case nas.TxDip:
		log.Infof("tx type :%s", tx.Type)
		break
	case nas.TxProtocol:
		log.Infof("tx type :%s", tx.Type)
		break
	case nas.TxNormal:
		blocktx.Amount = decimal.NewFromBigInt(tx.Value, 0)
		blocktx.ToAddress = tx.To
		blocktx.ContractAddress = ""
		break
	default:
		return nil, fmt.Errorf("ParseTransferData input : %s, err: %v", blocktx.Data, fmt.Errorf("don't know type %s", tx.Type))
	}

	data, err := utils.Base64Decode([]byte(blocktx.Data))
	if err != nil {
		return nil, err
	}

	blocktx.Data = string(data)

	return blocktx, nil
}
