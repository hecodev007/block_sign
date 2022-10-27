package avax

import (
	"avaxDataServer/common"
	"avaxDataServer/common/log"
	"avaxDataServer/conf"
	"avaxDataServer/models/bo"
	dao "avaxDataServer/models/po/avax"
	"avaxDataServer/services"
	"avaxDataServer/utils/avax"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
)

type Processor struct {
	*avax.RpcClient
	wg   *sync.WaitGroup
	lock *sync.Mutex

	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {

	return &Processor{
		RpcClient: avax.NewRpcClient(node.Url, node.Node, node.RPCKey, node.RPCSecret),
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

	log.Infof("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	//
	chaintx, err := s.getBlockTxInfosFromDB(txid)
	if err != nil {
		chaintx,err = s.RpcClient.GetRawTransactionFromScan(txid)
		if err != nil {
			return errors.New("交易没找到")
		}
	}
	log.Info(String(chaintx))
	dbHeight, err := dao.GetMaxBlockIndex()
	if err != nil {
		return fmt.Errorf("GetMaxBlockIndex height: %d , err: %v", dbHeight, err)
	}
	bestBlockHeight, err := s.GetBlockCount()
	if err != nil {
		return err
	}

	watchaddrs, _, _, err := s.processTX(&chaintx)
	if err != nil {
		return err
	}

	if len(watchaddrs) == 0 {
		return fmt.Errorf("txid %s don't have care of address", txid)
	}

	task := &AvaxProcTask{
		Irreversible: true,
		Height:       bestBlockHeight - 10,
		BestHeight:   dbHeight,
		Confirms:     10,
		TxInfos:      []*avax.Transaction{&chaintx},
	}
	for k, _ := range task.TxInfos[0].Inputs {
		td, err := decimal.NewFromString(task.TxInfos[0].Inputs[k].Output.Amount)
		if err != nil {
			task.TxInfos[0].Inputs[k].Output.Amount = "0"
			continue
		}
		task.TxInfos[0].Inputs[k].Output.Amount = td.String()
	}
	for k, _ := range task.TxInfos[0].Outputs {
		td, err := decimal.NewFromString(task.TxInfos[0].Outputs[k].Amount)
		if err != nil {
			task.TxInfos[0].Outputs[k].Amount = "0"
			continue
		}
		task.TxInfos[0].Outputs[k].Amount = td.String()
	}
	return s.processPush(task, 0, watchaddrs)
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
func (s *Processor) ProcIrreverseTxs(task common.ProcTask) error {
	if len(task.GetTxs()) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}
	avaxTask, ok := task.(*AvaxProcTask)
	if !ok {
		log.Error("task.(*AvaxProcTask)")
		return errors.New("error tast type")
	}
	if err := s.CheckIrreverseBlock(avaxTask.GetBlockHash()); err != nil {
		//log.Info(avaxTask.GetBlockHash(),avaxTask.Height)
		log.Info(err.Error())
		if err := dao.BatchInsertBlockTXs(avaxTask.TxInfos, avaxTask.Height); err != nil {
			log.Errorf("Batch Insert TXs err : %v", err)
		}
		return nil
	}
	//统一处理vout
	s.processTXVouts(avaxTask.TxInfos, avaxTask.Height)

	var txs []*avax.Transaction
	var updateVins []*dao.BlockTxVout //批量更新
	var insertVins []*avax.Input      //批量插入

	for k, txinfo := range avaxTask.TxInfos {
		//txinfo := tmp.(*avax.Transaction)
		if watchaddrs, updates, inserts, err := s.processTX(txinfo); err == nil {
			txs = append(txs, txinfo)
			if len(updates) > 0 {
				updateVins = append(updateVins, updates...)
			}
			if len(inserts) > 0 {
				insertVins = append(insertVins, inserts...)
			}
			if len(watchaddrs) > 0 {
				//bj,_ := json.Marshal(avaxTask)
				//log.Info(string(bj))
				s.processPush(avaxTask, k, watchaddrs)
			}
		}
	}

	if err := dao.BatchInsertBlockTXins(insertVins, avaxTask.Height); err != nil {
		log.Errorf("Batch Insert Vins err : %v", err)
	}

	if err := dao.BatchUpdateBlockTXVouts(updateVins); err != nil {
		log.Errorf("Batch Update Vins err : %v", err)
	}

	if err := dao.BatchInsertBlockTXs(avaxTask.TxInfos, avaxTask.Height); err != nil {
		log.Errorf("Batch Insert TXs err : %v", err)
	}

	return nil
}

//处理可逆交易
func (s *Processor) ProcReverseTxs(tmps []interface{}, bestHeight int64, confirm int64) error {

	return nil
}

func (s *Processor) ProcIrreverseBlock(b interface{}) error {
	return nil
}

func (s *Processor) UpdateIrreverseConfirms() {

}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	//s.confirmsPush(b.(*dao.BlockInfo))
}

func (s *Processor) processTXVouts(tmp []*avax.Transaction, height int64) {
	if len(tmp) == 0 {
		return
	}

	var insertVouts []*avax.Output

	for _, txinfo := range tmp {

		for _, vout := range txinfo.Outputs {
			if len(vout.Addresses) == 1 && string(vout.Addresses[0]) == "" {
				continue
			}

			if s.conf.FullBackup {
				insertVouts = append(insertVouts, vout)
			} else if len(vout.Addresses) == 1 && s.watch.IsWatchAddressExist(string(vout.Addresses[0])) {
				insertVouts = append(insertVouts, vout)
			}
		}
	}

	if err := dao.BatchInsertBlockTXVouts(insertVouts, height); err != nil {
		log.Errorf("block %d Batch Insert Vouts err : %v", height, err)
	}
	//睡眠1秒保证数据库一致
	//time.Sleep(time.Millisecond * 10)
}

func (s *Processor) processTX(txInfo *avax.Transaction) (map[string]bool, []*dao.BlockTxVout, []*avax.Input, error) {

	if txInfo == nil {
		return nil, nil, nil, fmt.Errorf("tx info don't allow nil")
	}

	var updateVins []*dao.BlockTxVout
	var insertVins []*avax.Input
	tmpWatchList := make(map[string]bool)
	//log.Infof("%+v",txInfo)
	//starttime := time.Now()
	amtout := decimal.Zero
	for _, txvout := range txInfo.Outputs {
		if txvout.Locktime != 0 {
			log.Info("locktime")
			return nil, nil, nil, nil
		}
		value, err := decimal.NewFromString(txvout.Amount)
		if err != nil {
			log.Error(err.Error())
			return nil, nil, nil, err
		}
		amtout = amtout.Add(value)
		//log.Info(txInfo.ID, string(txvout.Addresses[0]))
		if len(txvout.Addresses) == 1 && s.watch.IsWatchAddressExist(string(txvout.Addresses[0])) && txvout.AssetID == "FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z" {
			log.Info(txInfo.ID, string(txvout.Addresses[0]))
			tmpWatchList[string(txvout.Addresses[0])] = true
		}
	}

	for _, txvin := range txInfo.Inputs {
		if vout, err := dao.SelectBlockTXVout(txvin.Output.TransactionID, txvin.Output.OutputIndex); err == nil {
			vout.SpendTxid = txInfo.ID

			if s.conf.FullBackup {
				updateVins = append(updateVins, vout)
			} else {
				//log.Info(txInfo.ID,string(txvin.Output.Addresses[0]))
				if len(txvin.Output.Addresses) == 1 && s.watch.IsWatchAddressExist(string(txvin.Output.Addresses[0])) && txvin.Output.AssetID == "FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z" {
					log.Info(txInfo.ID, string(txvin.Output.Addresses[0]))
					updateVins = append(updateVins, vout)
				}
			}
		} else {

			if s.conf.FullBackup {
				insertVins = append(insertVins, txvin)
			} else {
				if len(txvin.Output.Addresses) == 1 && s.watch.IsWatchAddressExist(string(txvin.Output.Addresses[0])) && txvin.Output.AssetID == "FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z" {
					insertVins = append(insertVins, txvin)
				}
			}
		}
		//log.Info(txInfo.ID,string(txvin.Output.Addresses[0]))

		if len(txvin.Output.Addresses) == 1 && s.watch.IsWatchAddressExist(string(txvin.Output.Addresses[0])) && txvin.Output.AssetID == "FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z" {
			tmpWatchList[string(txvin.Output.Addresses[0])] = true
		}
	}

	//log.Infof("processTX : %s ,vins : %d , vouts : %d , used time : %f 's", blocktx.Txid, len(txvins), len(insertVouts), time.Since(starttime).Seconds())
	return tmpWatchList, updateVins, insertVins, nil
}

func (s *Processor) processPush(task *AvaxProcTask, index int, tmpWatchList map[string]bool) error {
	pushBlockTx := &bo.PushUtxoBlockInfo{
		Type:          bo.PushTypeTX,
		CoinName:      s.conf.Name,
		Height:        task.Height,
		Hash:          task.TxInfos[index].ID,
		Confirmations: task.Confirms,
		Time:          task.TxInfos[index].Timestamp.Unix(),
	}

	pushUtxoTx := bo.PushUtxoTx{
		Txid: task.TxInfos[index].ID,
		Fee:  "0.001",
		Vout: make([]bo.PushTxOutput, 0, 0),
		Vin:  make([]bo.PushTxInput, 0, 0),
	}

	for _, txvout := range task.TxInfos[index].Outputs {

		pushUtxoTx.Vout = append(pushUtxoTx.Vout, bo.PushTxOutput{
			AssetID:  txvout.AssetID,
			Addresse: string(txvout.Addresses[0]),
			Value:    txvout.Amount,
			N:        int(txvout.OutputIndex),
		})
	}

	for _, txvin := range task.TxInfos[index].Inputs {
		pushUtxoTx.Vin = append(pushUtxoTx.Vin, bo.PushTxInput{
			AssetID:  txvin.Output.AssetID,
			Txid:     txvin.Output.TransactionID,
			Vout:     int(txvin.Output.OutputIndex),
			Addresse: string(txvin.Output.Addresses[0]),
			Value:    txvin.Output.Amount,
		})
	}

	pushBlockTx.Txs = append(pushBlockTx.Txs, pushUtxoTx)
	pusdata, err := json.Marshal(&pushBlockTx)
	log.Info(string(pusdata))
	if err != nil {
		return err
	}

	if s.pusher != nil && (len(pushUtxoTx.Vout) > 0 || len(pushUtxoTx.Vin) > 0) {
		s.pusher.AddPushTask(pushBlockTx.Height, pushUtxoTx.Txid, tmpWatchList, pusdata)
	}
	return nil
}

func (s *Processor) confirmsPush(blockInfo interface{}) error {

	return nil
}

func (s *Processor) getBlockTxInfosFromNode(txid string) (avax.Transaction, error) {
	return s.RpcClient.GetTransactionByHash(txid)
}

func (s *Processor) getBlockTxInfosFromDB(txid string) (tx avax.Transaction, err error) {
	blocktx,err := dao.GetBlockTxByTxid(txid)
	if err != nil {
		return tx,err
	}
	if blocktx == nil {
		return tx,errors.New("交易数据库没找到,找管理员,手动添加此交易")
	}
	err = json.Unmarshal([]byte(blocktx.Rawtx),&tx)
	return

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
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
