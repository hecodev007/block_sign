package vet

import (
	"encoding/json"
	"fmt"
	"log"
	"rsksync/common"
	"rsksync/conf"
	"rsksync/models/bo"
	dao "rsksync/models/po/vet"
	"rsksync/services"
	"rsksync/utils/vet"
	"sync"
	"time"
)

type Processor struct {
	*vet.VetHttpClient

	watch  *services.WatchControl
	pusher *services.PushServer

	mempool map[string]*VetProcTask
	conf    conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {

	return &Processor{
		VetHttpClient: vet.NewVetHttpClient(node.Url),
		watch:         watch,

		mempool: make(map[string]*VetProcTask),
		conf:    conf.Sync,
	}
}

func (s *Processor) RepushTxByIsInternal(userId int64, txid string, isInternal bool) error {
	panic("implement me")
}

func (s *Processor) SetPusher(p common.Pusher) {
	if pusher, ok := p.(*services.PushServer); ok {
		s.pusher = pusher
	}
}

func (s *Processor) RemovePusher() {
	s.pusher = nil
}

func (s *Processor) RepushTx(userid int64, txid string) error {
	var (
		txInfo *TxInfo
		err    error
	)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	if txInfo, err = s.getBlockTxInfosFromDB(txid); err != nil {
		if txInfo, err = s.getBlockTxInfosFromNode(txid); err != nil {
			return err
		}
	}

	bestHeight, err := s.GetBestHeight()
	if err != nil {
		return err
	}

	return s.processTX(txInfo, bestHeight, false)
}

func (s *Processor) Info() (string, int64, error) {
	dbheight, err := dao.GetMaxBlockIndex()
	return s.conf.Name, dbheight, err
}

func (s *Processor) Init() error {
	return nil
}

func (s *Processor) Clear() {

}

func (s *Processor) CheckIrreverseBlock(hash string) error {
	cnt, err := dao.GetBlockCountByHash(hash)
	if err != nil {
		return fmt.Errorf("get BlockCount ByHash err: %v", err)
	}

	if cnt > 0 {
		return fmt.Errorf("already have block hash: %s , count: %d", hash, cnt)
	}
	return nil
}

func (s *Processor) ProcIrreverseTxs(tmps []interface{}, bestHeight int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}

	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		wg.Add(len(tmps))
		for _, tmp := range tmps {
			tx := tmp.(*TxInfo)
			//go s.batchProcessTX(tx, task.block.Timestamp, task.bestHeight, wg)
			go func(w *sync.WaitGroup) {
				defer w.Done()
				if err := s.processTX(tx, bestHeight, true); err == nil {
					if num, err := dao.InsertBlockTX(tx.tx); num <= 0 || err != nil {
						log.Printf("block tx insert err: %v", err)
					}
				}
			}(wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.(*TxInfo)
			if err := s.processTX(tx, bestHeight, true); err == nil {
				if num, err := dao.InsertBlockTX(tx.tx); num <= 0 || err != nil {
					log.Printf("block tx insert err: %v", err)
				}
			}
		}
	}

	return nil
}

func (s *Processor) ProcReverseTxs(tmps []interface{}, bestHeight int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}

	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		wg.Add(len(tmps))
		for _, tmp := range tmps {
			tx := tmp.(*TxInfo)
			go func(w *sync.WaitGroup) {
				defer w.Done()
				if err := s.processTX(tx, bestHeight, false); err == nil {
				}
			}(wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.(*TxInfo)
			if err := s.processTX(tx, bestHeight, false); err == nil {
			}
		}
	}

	return nil
}

func (s *Processor) ProcIrreverseBlock(b interface{}) error {
	block := b.(*dao.BlockInfo)
	if num, err := dao.InsertBlockInfo(block); num <= 0 || err != nil {
		return fmt.Errorf("block %d Insert Block err : %v", block.Height, err)
	}
	return nil
}

func (s *Processor) UpdateIrreverseConfirms() {
	//查找所有未确认的区块
	if bs, err := dao.GetUnconfirmBlockInfos(s.conf.Confirmations + 6); err == nil && bs != nil && len(bs) > 0 {
		var ids []int64
		//开始同步更新确认数
		for _, blk := range bs {
			blk.Confirmations++
			s.confirmsPush(blk)
			ids = append(ids, blk.Id)
		}
		//批量更新订单确认数。
		if err := dao.BatchUpdateConfirmations(ids, 1); err != nil {
			log.Printf("batch update confirmations err: %v", err)
		}
	}
}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	s.confirmsPush(b.(*dao.BlockInfo))
}

func (s *Processor) getBlockTxInfosFromDB(txid string) (*TxInfo, error) {
	blocktx, err := dao.SelecBlockTxByHash(txid)
	if err != nil {
		return nil, err
	}

	txs, err := dao.SelecBlockTxDetailByHash(txid)
	if err != nil {
		return nil, err
	}

	return &TxInfo{
		blocktx,
		txs,
	}, nil
}

func (s *Processor) getBlockTxInfosFromNode(txid string) (*TxInfo, error) {
	if txid == "" {
		return nil, fmt.Errorf("tx is null")
	}

	tx, err := s.GetTransaction(txid)
	if err != nil {
		return nil, err
	}

	txreceipt, err := s.GetTransactionReceipt(txid)
	if err != nil {
		return nil, err
	}

	return createTxDetail(tx, txreceipt)
}

func (s *Processor) batchProcessTX(blocktx *dao.BlockTX, blockTimestamp time.Time, bestHeight int64, wg *sync.WaitGroup) {

	defer wg.Done()
	if err := s.processTX(nil, bestHeight, false); err == nil {
		if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
			log.Printf("block tx insert err: %v", err)
		}
	}
}

// 解析交易信息到db
func (s *Processor) processTX(txInfo *TxInfo, bestHeight int64, isStore bool) error {

	if txInfo == nil {
		return fmt.Errorf("tx is null")
	}

	var txdetals []*dao.BlockTxDetail
	//检测是否为关心的地址
	tmpWatchList := make(map[string]bool)
	for i, txd := range txInfo.details {
		txd.Index = i
		if s.watch.IsWatchAddressExist(txd.FromAddress) {
			tmpWatchList[txd.FromAddress] = true
		}
		if s.watch.IsWatchAddressExist(txd.ToAddress) {
			tmpWatchList[txd.ToAddress] = true
		}

		if s.conf.FullBackup {
			if isStore {
				if _, err := dao.InsertBlockTxDetail(txd); err != nil {
					log.Printf("insert block txdetail err : %v", err)
				}
			}
			txdetals = append(txdetals, txd)
		} else {
			if s.watch.IsWatchAddressExist(txd.FromAddress) || s.watch.IsWatchAddressExist(txd.ToAddress) {
				if isStore {
					if _, err := dao.InsertBlockTxDetail(txd); err != nil {
						log.Printf("insert block txdetail err : %v", err)
					}
				}
				txdetals = append(txdetals, txd)
			}
		}
	}

	if !s.conf.FullBackup {
		if len(tmpWatchList) <= 0 {
			return fmt.Errorf("dont't have care of watch address ")
		}
	}

	if len(tmpWatchList) > 0 {
		if txInfo.tx.Status == 1 {
			if bestHeight > 0 {
				if oldTask, ok := s.mempool[txInfo.tx.BlockHash]; !ok {
					s.processPush(txInfo.tx, txdetals, tmpWatchList, bestHeight)
				} else {
					//如果之前有推送，就推确认数
					if oldTask.block.Confirmations < bestHeight-txInfo.tx.BlockHeight+1 {
						oldTask.block.Confirmations = bestHeight - txInfo.tx.BlockHeight + 1
						oldTask.block.Timestamp = txInfo.tx.Timestamp
						s.confirmsPush(oldTask.block)
					}
				}
			}
		} else {
			log.Printf("block tx status : %d is failed", txInfo.tx.Status)
		}
	}

	return nil
}

func (s *Processor) processPush(blocktx *dao.BlockTX, txdetails []*dao.BlockTxDetail, tmpWatchList map[string]bool, bestHeight int64) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktx.BlockHeight,
		Hash:          blocktx.BlockHash,
		CoinName:      s.conf.Name,
		Confirmations: bestHeight - blocktx.BlockHeight + 1,
		Time:          blocktx.Timestamp.Unix(),
	}

	fee, _ := vet.GetVTHOFeeFromStr(blocktx.PaidVTHO)
	for _, txInfo := range txdetails {
		pushtx := bo.PushAccountTx{
			Name:     txInfo.CoinName,
			Txid:     txInfo.Txid,
			From:     txInfo.FromAddress,
			To:       txInfo.ToAddress,
			Contract: txInfo.ContractAddress,
			Fee:      fee.String(),
			Amount:   txInfo.Amount.Shift(int32(0 - vet.WEI)).String(),
		}
		if pushtx.Contract != "" {
			pushtx.Name = "vtho"
		}
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)
	}

	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return err
	}

	if s.pusher != nil {
		s.pusher.AddPushTask(pushBlockTx.Height, blocktx.Txid, tmpWatchList, pusdata)
	}
	return nil
}

func (s *Processor) confirmsPush(blockInfo *dao.BlockInfo) error {

	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountConfir,
		Height:        blockInfo.Height,
		Hash:          blockInfo.Hash,
		CoinName:      s.conf.Name,
		Confirmations: blockInfo.Confirmations,
		Time:          blockInfo.Timestamp.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return err
	}

	if s.pusher != nil {
		s.pusher.AddPushUserTask(blockInfo.Height, pushdata)
	}

	return nil
}
