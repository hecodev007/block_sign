package kava

import (
	"encoding/json"
	"fmt"
	"log"
	"rsksync/common"
	"rsksync/conf"
	"rsksync/models/bo"
	dao "rsksync/models/po/kava"
	"rsksync/services"
	"rsksync/utils/kava"
	"sync"
	"time"
)

type Processor struct {
	client  *kava.HttpClient
	watch   *services.WatchControl
	pusher  *services.PushServer
	mempool map[string]*ProcTask
	conf    conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		client:  kava.NewHttpClient(node.Url),
		watch:   watch,
		mempool: make(map[string]*ProcTask),
		conf:    conf.Sync,
	}
}

func (s *Processor) RepushTxByIsInternal(userId int64, txid string, isInternal bool) error {
	panic("implement me")
}

func (s *Processor) SetPusher(p common.Pusher) {
	pusher, ok := p.(*services.PushServer)
	if ok {
		s.pusher = pusher
	}
}
func (s *Processor) RemovePusher() {
	s.pusher = nil
}
func (s *Processor) RepushTx(userid int64, txid string) error {
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	txInfo, err := s.getBlockTxInfosFromDB(txid)
	if err != nil {
		txInfo, err = s.getBlockTxInfosFromNode(txid)
		if err != nil {
			return err
		}
	}
	bestHeight, err := s.client.GetLastBlockHeight()
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
		return fmt.Errorf("already have block  hash: %s , count: %d", hash, cnt)
	}
	return nil
}
func (s *Processor) ProcIrreverseTxs(tmps []interface{}, bestHeight int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}
	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		for _, tmp := range tmps {
			go func(txinfo *TxInfo, w *sync.WaitGroup) {
				w.Add(1)
				defer w.Done()
				if err := s.processTX(txinfo, bestHeight, true); err == nil {
					if num, err := dao.InsertBlockTX(txinfo.tx); num <= 0 || err != nil {
						log.Printf("block tx insert err: %v", err)
					}
				} else {
					log.Printf("err : %v", err)
				}
			}(tmp.(*TxInfo), wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			txinfo := tmp.(*TxInfo)
			if err := s.processTX(txinfo, bestHeight, true); err == nil {
				if num, err := dao.InsertBlockTX(txinfo.tx); num <= 0 || err != nil {
					log.Printf("block tx insert err: %v", err)
				}
			} else {
				log.Printf("err : %v", err)
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
		for _, tmp := range tmps {
			go func(txinfo *TxInfo, w *sync.WaitGroup) {
				w.Add(1)
				defer w.Done()
				if err := s.processTX(txinfo, bestHeight, false); err == nil {
					//log.Printf("err : %v", err)
				}
			}(tmp.(*TxInfo), wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			txinfo := tmp.(*TxInfo)
			if err := s.processTX(txinfo, bestHeight, false); err == nil {
				//log.Printf("err : %v", err)
			}
		}
	}
	return nil
}

func (s *Processor) ProcIrreverseBlock(b interface{}) error {
	block := b.(*dao.BlockInfo)
	if _, err := dao.InsertBlockInfo(block); err != nil {
		return fmt.Errorf("block %d Insert Block err : %v", block.Height, err)
	}
	return nil
}

func (s *Processor) UpdateIrreverseConfirms() {
	//查找所有未确认的区块
	bs, err := dao.GetUnconfirmBlockInfos(s.conf.Confirmations * 2)
	if err == nil && bs != nil && len(bs) > 0 {
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
	txs, err := dao.SelectBlockTXMsgsByTxid(txid)
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
	tx, err := s.client.GetTransactionByHash(txid, time.Now())
	if err != nil || tx == nil {
		return nil, err
	}
	log.Printf("GetTransactionByHash %v ", tx)
	block, err := s.client.GetBlockByHeight(tx.BlockHeight)
	if err != nil {
		return nil, err
	}
	tx.Timestamp = block.Timestamp
	return parseBlockTX(tx, block.Hash)
}

// 解析交易信息到db
func (s *Processor) processTX(txInfo *TxInfo, bestHeight int64, isStore bool) error {
	if txInfo == nil {
		return fmt.Errorf("tx is null")
	}
	log.Printf("processTX 1: %v, tx num : %d", txInfo, len(txInfo.txmsgs))
	var careMsgs []*dao.BlockTXMsg
	//检测是否为关心的地址
	tmpWatchList := make(map[string]bool)
	for _, txmsg := range txInfo.txmsgs {
		if s.watch.IsWatchAddressExist(txmsg.FromAddress) {
			tmpWatchList[txmsg.FromAddress] = true
		}
		if s.watch.IsWatchAddressExist(txmsg.ToAddress) {
			tmpWatchList[txmsg.ToAddress] = true
		}
		if s.conf.FullBackup {
			if isStore {
				if _, err := dao.InsertBlockTXMsg(txmsg); err != nil {
					log.Printf("insert block txmsg err : %v", err)
				}
			}
			if txmsg.Status == 1 {
				careMsgs = append(careMsgs, txmsg)
			}
		} else {
			if s.watch.IsWatchAddressExist(txmsg.FromAddress) || s.watch.IsWatchAddressExist(txmsg.ToAddress) {
				if isStore {
					if _, err := dao.InsertBlockTXMsg(txmsg); err != nil {
						log.Printf("insert block txdetail err : %v", err)
					}
				}
				if txmsg.Status == 1 {
					careMsgs = append(careMsgs, txmsg)
				}
			}
		}
	}
	if !s.conf.FullBackup {
		if len(tmpWatchList) <= 0 {
			return fmt.Errorf("dont't have care of watch address ")
		}
	}
	if len(tmpWatchList) > 0 {
		if bestHeight > 0 {
			oldTask, ok := s.mempool[txInfo.tx.BlockHash]
			if !ok {
				s.processPush(txInfo.tx, careMsgs, tmpWatchList, bestHeight)
			} else {
				//如果之前有推送，就推确认数
				if oldTask.block.Confirmations < bestHeight-txInfo.tx.BlockHeight+1 {
					oldTask.block.Confirmations = bestHeight - txInfo.tx.BlockHeight + 1
					oldTask.block.Timestamp = txInfo.tx.Timestamp
					s.confirmsPush(oldTask.block)
				}
			}
		}
	}
	log.Printf("processTX 2: %v, tx num : %d", txInfo, len(txInfo.txmsgs))
	return nil
}

func (s *Processor) processPush(blocktx *dao.BlockTX, txmsgs []*dao.BlockTXMsg, tmpWatchList map[string]bool, bestHeight int64) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktx.BlockHeight,
		Hash:          blocktx.BlockHash,
		CoinName:      s.conf.Name,
		Confirmations: bestHeight - blocktx.BlockHeight + 1,
		Time:          blocktx.Timestamp.Unix(),
	}
	for _, txMsg := range txmsgs {
		if txMsg.Type != "bank" {
			continue
		}
		pushtx := bo.PushAccountTx{
			Name:   s.conf.Name,
			Txid:   txMsg.Txid,
			From:   txMsg.FromAddress,
			To:     txMsg.ToAddress,
			Fee:    kava.GetKavaNum(blocktx.Fee).String(),
			Amount: kava.GetKavaNum(txMsg.Amount).String(),
		}
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)
	}
	if len(pushBlockTx.Txs) == 0 {
		return fmt.Errorf("haven't need push tx msg")
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
	if err == nil && s.pusher != nil {
		s.pusher.AddPushUserTask(blockInfo.Height, pushdata)
	}
	return nil
}
