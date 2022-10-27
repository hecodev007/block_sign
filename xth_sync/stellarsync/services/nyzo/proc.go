package nyzo

import (
	"stellarsync/common"
	"stellarsync/common/conf"
	"stellarsync/common/log"
	"stellarsync/models/bo"
	dao "stellarsync/models/po/nyzo"
	"stellarsync/services"
	rpc "stellarsync/utils/stellar"
	"encoding/json"
	"fmt"
)

type Processor struct {
	*rpc.RpcClient
	watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		watch:     watch,
		conf:      conf.Sync,
	}
}
func (p *Processor) Init() error {
	return nil
}
func (p *Processor) Clear() {
}
func (p *Processor) SetPusher(push common.Pusher) {
	pusher, ok := push.(*services.PushServer)
	if ok {
		p.pusher = pusher
	}
}
func (p *Processor) RemovePusher() {
	p.pusher = nil
}

func (p *Processor) Info() (string, int64, error) {
	dbheight, err := dao.MaxBlockHeight()
	return p.conf.Name, dbheight, err
}

func (p *Processor) CheckIrreverseBlock(hash string) error {
	if has, err := dao.BlockHashExist(hash); err != nil {
		return fmt.Errorf("get BlockCount ByHash err: %v", err)
	} else if has {
		return fmt.Errorf("already have block  hash: %s , count: %d", hash, 1)
	}
	return nil
}

//以上全世界都一样

////暂没用到 ，查询数据是否已有这个区块
//CheckIrreverseBlock(hash string) error
////处理不可逆区块交易(推送交易，保存到数据库)
//ProcIrreverseTxs(ProcTask) error
////推送不可逆交易确认数（） 待定
//UpdateIrreverseConfirms(ProcTask)
////处理不可逆区块（保存到数据库）
//ProcIrreverseBlock(ProcTask) error
////处理可逆区块交易(是否需要推送)；有需要推送？，error
//ProcReverseTxs(ProcTask) (bool, error)
////推送可逆区块确认数(不需要保存数据库，直接推送数据)
//UpdateReverseConfirms(ProcTask)

//处理不可逆区块交易(推送交易，保存到数据库)
func (p *Processor) ProcIrreverseTask(procTask common.ProcTask) error {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	//tj, _ := json.Marshal(task)
	//log.Info(string(tj))
	//panic("")
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := p.parseWatchAddr(txInfo)

		if p.conf.FullBackup {
			dao.InsertTx(txInfo)
		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo)
		}

		if len(watchAddrs) > 0 {
			p.processPush(task.TxInfos[k], watchAddrs, task.BestHeight)
		}

	}
	dao.InsertBlock(task.Block)
	return nil
}
func (s *Processor) processPush(blocktx *dao.BlockTx, tmpWatchList map[string]bool, bestHeight int64) error {

	pushBlockInfo := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		CoinName:      s.conf.Name,
		Height:        blocktx.BlockHeight,
		Hash:          blocktx.BlockHash,
		Confirmations: bestHeight - blocktx.BlockHeight + 1,
		Time:          blocktx.Timestamp.Unix(),
		Txs:           make([]*bo.PushAccountTx, 0),
	}

	pushAccountTx := &bo.PushAccountTx{
		Name:     s.conf.Name,
		Txid:     blocktx.Txid,
		Fee:      blocktx.Fee,
		From:     blocktx.From,
		To:       blocktx.To,
		Amount:   blocktx.Value,
		Memo:     blocktx.Memo,
		Contract: "",
	}
	pushBlockInfo.Txs = append(pushBlockInfo.Txs, pushAccountTx)
	pushData, _ := json.Marshal(pushBlockInfo)
	//println(string(pushData))
	if s.pusher != nil {
		s.pusher.AddPushTask(blocktx.BlockHeight, blocktx.Txid, tmpWatchList, pushData)
	}
	return nil
}
func (p *Processor) parseWatchAddr(tx *dao.BlockTx) (watchaddrs map[string]bool) {
	watchaddrs = make(map[string]bool)
	if p.watch.IsWatchAddressExist(tx.From) {
		watchaddrs[tx.From] = true
	}
	if p.watch.IsWatchAddressExist(tx.To) {
		watchaddrs[tx.To] = true
	}
	return
}
func (p *Processor) UpdateAmount(addr string) error {
	return nil
}

func (s *Processor) RepushTx(userid int64, txid string,height int64) error {
	var (
		err    error
		txinfos []*dao.BlockTx
	)

	log.Infof("RepushTx user: %d , txid : %s, height:%v", userid, txid,height)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	 txinfos, err = s.getBlockTxInfosFromNode(txid,height);
	 if err != nil {
		return fmt.Errorf("get block falied: %v", err)
	}
	if len(txinfos) == 0 {
		return fmt.Errorf("交易没找到或失败的交易")
	}
	bestBlockHeight, err := s.GetBlockCount()
	if err != nil {
		return err
	}
	var lenwatchaddr int
	for _,txinfo := range txinfos {
		watchaddrs := s.parseWatchAddr(txinfo)
		lenwatchaddr += len(watchaddrs)
		if len(watchaddrs) != 0 {
			s.processPush(txinfo, watchaddrs, bestBlockHeight)
		}
	}
	if lenwatchaddr == 0 {
		return fmt.Errorf("交易没有监控的地址")
	}
	return nil
}

//func (s *Processor) processTX(txInfo *TxInfo) (map[string]bool, []*dao.BlockTxVout, []*dao.BlockTxVout, error) {
//
//	if txInfo == nil {
//		return nil, nil, nil, fmt.Errorf("tx info don't allow nil")
//	}
//
//	var updateVins, insertVins []*dao.BlockTxVout
//	tmpWatchList := make(map[string]bool)
//
//	//starttime := time.Now()
//	amtout := decimal.Zero
//	for _, txvout := range txInfo.vouts {
//		amtout = amtout.Add(txvout.Value)
//
//		if s.watch.IsWatchAddressExist(txvout.Address) {
//			tmpWatchList[txvout.Address] = true
//		}
//	}
//
//	amtin := decimal.Zero
//	for _, txvin := range txInfo.vins {
//
//		if txvin.Txid == "coinbase" {
//			continue
//		}
//
//		if vout, err := dao.SelectBlockTXVout(txvin.Txid, txvin.Voutn); err == nil {
//			txvin.Id = vout.Id
//			txvin.Value = vout.Value
//			txvin.Address = vout.Address
//
//			if s.conf.FullBackup {
//				updateVins = append(updateVins, txvin)
//			} else {
//				if s.watch.IsWatchAddressExist(txvin.Address) {
//					updateVins = append(updateVins, txvin)
//				}
//			}
//		} else {
//			if err != gorm.ErrRecordNotFound {
//				log.Printf("processTX SelectBlockTXVout txid: %s, n: %d, err: %v", txvin.Txid, txvin.Voutn, err)
//				continue
//			}
//			tx, err := s.GetRawTransaction(txvin.Txid)
//			if err != nil {
//				log.Println(err.Error())
//				return nil, nil, nil, fmt.Errorf("processTX : GetRawTransaction %s", txvin.Txid)
//			}
//			txvin.BlockHash = tx.BlockHash
//			txvin.Value = decimal.NewFromFloat(tx.Vout[txvin.Voutn].Value)
//			txvin.Timestamp = time.Unix(tx.Time, 0)
//			txvin.CreateTime = time.Now()
//			address, err := tx.Vout[txvin.Voutn].ScriptPubkey.GetAddress()
//			if err == nil {
//				txvin.Address = address[0]
//			}
//			data, _ := json.Marshal(tx.Vout[txvin.Voutn].ScriptPubkey)
//			txvin.ScriptPubKey = string(data)
//
//			if s.conf.FullBackup {
//				insertVins = append(insertVins, txvin)
//			} else {
//				if s.watch.IsWatchAddressExist(txvin.Address) {
//					insertVins = append(insertVins, txvin)
//				}
//			}
//		}
//
//		amtin = amtin.Add(txvin.Value)
//		if s.watch.IsWatchAddressExist(txvin.Address) {
//			tmpWatchList[txvin.Address] = true
//		}
//	}
//
//	txInfo.tx.Fee = amtin.Sub(amtout)
//	if txInfo.tx.Fee.IsNegative() && s.conf.Name != "doge" {
//		return nil, nil, nil, fmt.Errorf("tx fee don't allow negative,vin:%v, vout:%v", amtin, amtout)
//	}
//
//	//log.Printf("processTX : %s ,vins : %d , vouts : %d , used time : %f 's", blocktx.Txid, len(txvins), len(insertVouts), time.Since(starttime).Seconds())
//	return tmpWatchList, updateVins, insertVins, nil
//}
func (s *Processor) getBlockTxInfosFromNode(txid string,height int64) (ret []*dao.BlockTx, err error) {
	block,err := s.GetBlockByHeight(height)
	if err != nil {
		return nil,err
	}
	var txs []*rpc.Transaction
	for k,temptx := range block.Transactions{
		if temptx.Txid == txid {
			txs = append(txs,block.Transactions[k])
		}
	}
	if len(txs) == 0{
		return nil,fmt.Errorf("高度%v上没找到此交易",height)
	}
	for _,tx := range txs{
		blocktx,_ := s.parseBlockRawTX(tx, block.Hash, height)
		if blocktx != nil {
			ret = append(ret,blocktx)
		}
	}
	return
}

func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := p.parseWatchAddr(txInfo)

		if len(watchAddrs) > 0 {
			ret = true
			p.processPush(task.TxInfos[k], watchAddrs, task.BestHeight)
		}
	}
	return ret, nil
}

func (p *Processor) PushReverseConfirms(procTask common.ProcTask) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error ProcTask type")
	}
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountConfir,
		Height:        task.Block.Height,
		Hash:          task.Block.Hash,
		CoinName:      p.conf.Name,
		Confirmations: task.BestHeight - task.Block.Height + 1,
		Time:          task.Block.Timestamp.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}

//解析交易
func (s *Processor) parseBlockRawTX(tx *rpc.Transaction, blockhash string, height int64) (*dao.BlockTx, error) {
	return parseBlockRawTX(s.RpcClient, tx, blockhash, height)
}
