package btc

import (
	"avaxDataServer/common"
	"avaxDataServer/common/log"
	"avaxDataServer/conf"
	"avaxDataServer/models/bo"
	dao "avaxDataServer/models/po/btc"
	"avaxDataServer/services"
	"avaxDataServer/utils/btc"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

type Processor struct {
	*btc.RpcClient
	wg   *sync.WaitGroup
	lock *sync.Mutex

	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {

	return &Processor{
		RpcClient: btc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		watch:     watch,
		wg:        &sync.WaitGroup{},
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
	}
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
	var (
		err     error
		txinfo  *TxInfo
		confirm int64
	)

	log.Infof("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	//if txinfo, err = s.getBlockTxInfosFromDB(txid); err != nil {
	if txinfo, confirm, err = s.getBlockTxInfosFromNode(txid); err != nil {
		return fmt.Errorf("don't get block txinfos %v", err)
	}
	//}

	dbHeight, err := dao.GetMaxBlockIndex()
	if err != nil {
		return fmt.Errorf("GetMaxBlockIndex height: %d , err: %v", dbHeight, err)
	}

	if txinfo.tx.BlockHeight > dbHeight {
		return fmt.Errorf("don't sync reach %d, current %d", txinfo.tx.BlockHeight, dbHeight)
	}

	bestBlockHeight, err := s.GetBlockCount()
	if err != nil {
		return err
	}

	watchaddrs, _, _, err := s.processTX(txinfo)
	if err != nil {
		return err
	}

	if len(watchaddrs) == 0 {
		return fmt.Errorf("txid %s don't have care of address", txid)
	}

	return s.processPush(txinfo.tx, txinfo.vouts, txinfo.vins, watchaddrs, bestBlockHeight, confirm)
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

//处理不可逆交易
func (s *Processor) ProcIrreverseTxs(tmps []interface{}, bestHeight int64, confirms int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}

	//统一处理vout
	s.processTXVouts(tmps, bestHeight)

	var txs []*dao.BlockTX
	var updateVins []*dao.BlockTXVout //批量更新
	var insertVins []*dao.BlockTXVout //批量插入

	if s.conf.EnableGoroutine {
		lock := &sync.Mutex{}
		wg := &sync.WaitGroup{}
		wg.Add(len(tmps))
		for _, tmp := range tmps {
			//txinfo := tmp.(*TxInfo)
			go func(w *sync.WaitGroup, txinfo *TxInfo) {
				defer w.Done()
				if watchaddrs, updates, inserts, err := s.processTX(txinfo); err == nil {

					lock.Lock()
					txs = append(txs, txinfo.tx)
					if len(updates) > 0 {
						updateVins = append(updateVins, updates...)
					}
					if len(inserts) > 0 {
						insertVins = append(insertVins, inserts...)
					}
					lock.Unlock()

					if len(watchaddrs) > 0 {
						s.processPush(txinfo.tx, txinfo.vouts, txinfo.vins, watchaddrs, bestHeight, confirms)
					}
				}
			}(wg, tmp.(*TxInfo))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			txinfo := tmp.(*TxInfo)
			if watchaddrs, updates, inserts, err := s.processTX(txinfo); err == nil {
				txs = append(txs, txinfo.tx)
				if len(updates) > 0 {
					updateVins = append(updateVins, updates...)
				}
				if len(inserts) > 0 {
					insertVins = append(insertVins, inserts...)
				}
				if len(watchaddrs) > 0 {
					s.processPush(txinfo.tx, txinfo.vouts, txinfo.vins, watchaddrs, bestHeight, confirms)
				}
			}
		}
	}

	if err := dao.BatchInsertBlockTXVouts(insertVins); err != nil {
		log.Errorf("Batch Insert Vins err : %v", err)
	}

	if err := dao.BatchUpdateBlockTXVouts(updateVins); err != nil {
		log.Errorf("Batch Update Vins err : %v", err)
	}

	if err := dao.BatchInsertBlockTXs(txs); err != nil {
		log.Errorf("Batch Insert TXs err : %v", err)
	}

	return nil
}

//处理可逆交易
func (s *Processor) ProcReverseTxs(tmps []interface{}, bestHeight int64, confirm int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}

	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		wg.Add(len(tmps))
		for _, tmp := range tmps {
			//txinfo := tmp.(*TxInfo)
			go func(w *sync.WaitGroup, txinfo *TxInfo) {
				defer w.Done()
				if watchaddrs, _, _, err := s.processTX(txinfo); err == nil {
					if len(watchaddrs) > 0 {
						s.processPush(txinfo.tx, txinfo.vouts, txinfo.vins, watchaddrs, bestHeight, confirm)
					}
				}
			}(wg, tmp.(*TxInfo))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			txinfo := tmp.(*TxInfo)
			if watchaddrs, _, _, err := s.processTX(txinfo); err == nil {
				if len(watchaddrs) > 0 {
					s.processPush(txinfo.tx, txinfo.vouts, txinfo.vins, watchaddrs, bestHeight, confirm)
				}
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
			log.Errorf("batch update confirmations err: %v", err)
		}
	}
}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	s.confirmsPush(b.(*dao.BlockInfo))
}

func (s *Processor) processTXVouts(tmp []interface{}, height int64) {
	if len(tmp) == 0 {
		return
	}

	var insertVouts []*dao.BlockTXVout

	for _, txinfotmp := range tmp {
		txinfo := txinfotmp.(*TxInfo)
		for _, vout := range txinfo.vouts {
			if vout.Address == "" {
				continue
			}

			if s.conf.FullBackup {
				insertVouts = append(insertVouts, vout)
			} else {
				if s.watch.IsWatchAddressExist(vout.Address) {
					insertVouts = append(insertVouts, vout)
				}
			}
		}
	}

	if err := dao.BatchInsertBlockTXVouts(insertVouts); err != nil {
		log.Errorf("block %d Batch Insert Vouts err : %v", height, err)
	}

	//睡眠1秒保证数据库一致
	time.Sleep(time.Millisecond * 10)
}

func (s *Processor) processTX(txInfo *TxInfo) (map[string]bool, []*dao.BlockTXVout, []*dao.BlockTXVout, error) {

	if txInfo == nil {
		return nil, nil, nil, fmt.Errorf("tx info don't allow nil")
	}

	var updateVins, insertVins []*dao.BlockTXVout
	tmpWatchList := make(map[string]bool)

	//starttime := time.Now()
	amtout := decimal.Zero
	for _, txvout := range txInfo.vouts {
		amtout = amtout.Add(txvout.Value)

		if s.watch.IsWatchAddressExist(txvout.Address) {
			tmpWatchList[txvout.Address] = true
		}
	}

	amtin := decimal.Zero
	for _, txvin := range txInfo.vins {

		if txvin.Txid == "coinbase" {
			continue
		}

		if vout, err := dao.SelectBlockTXVout(txvin.Txid, txvin.Voutn); err == nil {
			txvin.Id = vout.Id
			txvin.Value = vout.Value
			txvin.Address = vout.Address

			if s.conf.FullBackup {
				updateVins = append(updateVins, txvin)
			} else {
				if s.watch.IsWatchAddressExist(txvin.Address) {
					updateVins = append(updateVins, txvin)
				}
			}
		} else {
			if err != gorm.ErrRecordNotFound {
				log.Errorf("processTX SelectBlockTXVout txid: %s, n: %d, err: %v", txvin.Txid, txvin.Voutn, err)
				continue
			}
			tx, err := s.GetRawTransaction(txvin.Txid)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("processTX : GetRawTransaction %s", txvin.Txid)
			}
			txvin.BlockHash = tx.BlockHash
			txvin.Value = decimal.NewFromFloat(tx.Vout[txvin.Voutn].Value)
			txvin.Timestamp = time.Unix(tx.Time, 0)
			txvin.CreateTime = time.Now()
			address, err := tx.Vout[txvin.Voutn].ScriptPubkey.GetAddress()
			if err == nil {
				txvin.Address = address[0]
			}
			data, _ := json.Marshal(tx.Vout[txvin.Voutn].ScriptPubkey)
			txvin.ScriptPubKey = string(data)

			if s.conf.FullBackup {
				insertVins = append(insertVins, txvin)
			} else {
				if s.watch.IsWatchAddressExist(txvin.Address) {
					insertVins = append(insertVins, txvin)
				}
			}
		}

		amtin = amtin.Add(txvin.Value)
		if s.watch.IsWatchAddressExist(txvin.Address) {
			tmpWatchList[txvin.Address] = true
		}
	}

	txInfo.tx.Fee = amtin.Sub(amtout)
	if txInfo.tx.Fee.IsNegative() && s.conf.Name != "doge" {
		return nil, nil, nil, fmt.Errorf("tx fee don't allow negative,vin:%v, vout:%v", amtin, amtout)
	}

	//log.Infof("processTX : %s ,vins : %d , vouts : %d , used time : %f 's", blocktx.Txid, len(txvins), len(insertVouts), time.Since(starttime).Seconds())
	return tmpWatchList, updateVins, insertVins, nil
}

func (s *Processor) processPush(blocktx *dao.BlockTX, txvouts []*dao.BlockTXVout, txvins []*dao.BlockTXVout, tmpWatchList map[string]bool, bestHeight int64, confirms int64) error {
	pushBlockTx := &bo.PushUtxoBlockInfo{
		Type:          bo.PushTypeTX,
		CoinName:      s.conf.Name,
		Height:        blocktx.BlockHeight,
		Hash:          blocktx.BlockHash,
		Confirmations: confirms,
		Time:          blocktx.Timestamp.Unix(),
	}

	pushUtxoTx := bo.PushUtxoTx{
		Txid: blocktx.Txid,
		Fee:  blocktx.Fee.String(),
	}

	for _, txvout := range txvouts {
		pushUtxoTx.Vout = append(pushUtxoTx.Vout, bo.PushTxOutput{
			Addresse: txvout.Address,
			Value:    txvout.Value.String(),
			N:        txvout.Voutn,
		})
	}

	for _, txvin := range txvins {
		pushUtxoTx.Vin = append(pushUtxoTx.Vin, bo.PushTxInput{
			Txid:     txvin.Txid,
			Vout:     txvin.Voutn,
			Addresse: txvin.Address,
			Value:    txvin.Value.String(),
		})
	}

	pushBlockTx.Txs = append(pushBlockTx.Txs, pushUtxoTx)
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return err
	}

	if s.pusher != nil {
		s.pusher.AddPushTask(pushBlockTx.Height, pushUtxoTx.Txid, tmpWatchList, pusdata)
	}
	return nil
}

func (s *Processor) confirmsPush(blockInfo *dao.BlockInfo) error {

	pushBlockTx := &bo.PushUtxoBlockInfo{
		Type:          bo.PushTypeConfir,
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

func (s *Processor) getBlockTxInfosFromNode(txid string) (*TxInfo, int64, error) {

	tx, err := s.GetRawTransaction(txid)
	if err != nil {
		return nil, 0, fmt.Errorf("GetRawTransaction txid: %s , err: %v", txid, err)
	}

	block, err := s.GetBlockByHash(tx.BlockHash)
	if err != nil {
		return nil, 0, fmt.Errorf("GetBlockByHash hash: %s , err: %v", tx.BlockHash, err)
	}
	txinfo, err := parseBlockRawTX(s.conf.Name, &tx, tx.BlockHash, block.Height)
	return txinfo, block.Confirmations, err
}

func (s *Processor) getBlockTxInfosFromDB(txid string) (*TxInfo, error) {
	var (
		err     error
		blocktx *dao.BlockTX
		txvouts []*dao.BlockTXVout
		txvins  []*dao.BlockTXVout
	)

	blocktx, err = dao.SelectBlockTX(txid)
	if err != nil {
		return nil, fmt.Errorf("SelectBlockTX err : %v", err)
	}

	txvouts, err = dao.SelectBlockTXVoutsByTxid(txid)
	if err != nil {
		return nil, fmt.Errorf("SelectBlockTX err : %v", err)
	}

	txvins, err = dao.SelectBlockTXVinsByTxid(txid)
	if err != nil {
		return nil, fmt.Errorf("SelectBlockTX err : %v", err)
	}

	return &TxInfo{
		tx:    blocktx,
		vouts: txvouts,
		vins:  txvins,
	}, nil
}

//func (s *Processor) batchProcessTX(jobs <-chan *TxInfo, results chan<- int, bestHeight int64) {
//	var txs []*po.BlockTX
//
//	var updateVins []*po.BlockTXVout //批量更新
//	var insertVins []*po.BlockTXVout //批量插入
//
//	//统一处理vout
//	s.processTXVouts(jobs)
//
//	count := len(jobs)
//	offset := 0
//	for i := 0; i < count; i++ {
//		select {
//		case txinfo := <-jobs:
//			offset += 1
//			watchaddrs, updates, inserts, err := s.processTX(txinfo)
//
//			if err == nil {
//				txs = append(txs, txinfo.tx)
//
//				if len(updates) > 0 {
//					updateVins = append(updateVins, updates...)
//				}
//				if len(inserts) > 0 {
//					insertVins = append(insertVins, inserts...)
//				}
//				if len(watchaddrs) > 0 {
//					oldTask, ok := s.mempool[txinfo.tx.BlockHash]
//					if !ok {
//						s.processPush(txinfo.tx, txinfo.vouts, txinfo.vins, watchaddrs, bestHeight)
//					} else {
//						//如果之前有推送，就推确认数
//						if oldTask.block.Confirmations < bestHeight-txinfo.tx.BlockHeight {
//							oldTask.block.Confirmations = bestHeight - txinfo.tx.BlockHeight
//							s.confirmsPush(oldTask.block)
//						}
//						//释放内存池区块
//						s.mempool[txinfo.tx.BlockHash] = nil
//					}
//				}
//			}
//
//			if offset >= count {
//				break
//			}
//		default:
//			offset += 1
//			if offset >= count {
//				break
//			}
//		}
//	}
//
//	dao.BatchInsertBlockTXVouts(insertVins)
//	dao.BatchUpdateBlockTXVouts(updateVins)
//	dao.BatchInsertBlockTXs(txs)
//
//	results <- 1
//}
