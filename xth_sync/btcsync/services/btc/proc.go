package btc

import (
	"btcsync/common"
	"btcsync/common/conf"
	"btcsync/common/log"
	"btcsync/models/bo"
	dao "btcsync/models/po/btc"
	"btcsync/services"
	rpc "btcsync/utils/btcrpc"
	"btcsync/utils/dingding"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/onethefour/common/xutils"
)

type Processor struct {
	*rpc.RpcClient
	Usdt   *rpc.RpcClient
	watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		Usdt:      rpc.NewRpcClient(node.Usdt, node.RPCKey, node.RPCSecret),
		watch:     watch,
		conf:      conf.Sync,
	}
}
func (p *Processor) Init() error {

	for _, v := range p.watch.WatchAddrs {
		go p.UpdateAmount(v[0].Address)

	}
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

	for k, txInfo := range task.TxInfos {
		watchAddrs := make(map[string]bool)
		//过滤出监控地址的vin,vout
		p.parseContractTX(txInfo, watchAddrs)

		if p.conf.FullBackup || len(watchAddrs) > 0 {
			dao.InsertTx(txInfo.Tx)
			dao.InsertTxVin(txInfo.Vins)
			dao.InsertTxVout(txInfo.Vouts)
			dao.InsertBtcUsdtTx(txInfo.Contractxs)
		}

		if len(watchAddrs) > 0 {
			p.processPush(task.TxInfos[k].Tx, task.TxInfos[k].Vouts, task.TxInfos[k].Vins, task.TxInfos[k].Contractxs, watchAddrs, task.BestHeight)
		}

	}
	dao.InsertBlock(task.Block)
	return nil
}
func (s *Processor) processPush(blocktx *dao.BlockTx, txvouts []*dao.BlockTxVout, txvins []*dao.BlockTxVin, txusdts []*dao.BtcUsdtTx, tmpWatchList map[string]bool, bestHeight int64) error {
	if len(txvins) == 0 && len(txvouts) == 0 {
		log.Info(tmpWatchList)
		panic("")
	}
	pushBlockTx := &bo.PushUtxoBlockInfo{
		Type:          bo.PushTypeTX,
		CoinName:      s.conf.Name,
		Height:        blocktx.Height,
		Hash:          blocktx.Blockhash,
		Confirmations: bestHeight - blocktx.Height + 1,
		Time:          blocktx.Timestamp.Unix(),
	}

	pushUtxoTx := &bo.PushUtxoTx{
		Txid:     blocktx.Txid,
		Fee:      blocktx.Fee,
		Coinbase: blocktx.Iscoinbase,
	}

	for _, txvout := range txvouts {
		rawAddr, _ := ChangeAddrLtcToBTC(txvout.Address)
		pushUtxoTx.Vout = append(pushUtxoTx.Vout, &bo.PushTxOutput{
			RawAddresse: rawAddr,
			Addresse:    txvout.Address,
			Value:       txvout.Value,
			N:           txvout.VoutN,
		})
	}

	for _, txvin := range txvins {
		pushUtxoTx.Vin = append(pushUtxoTx.Vin, &bo.PushTxInput{
			Txid:     txvin.Txid,
			Vout:     txvin.VoutN,
			Addresse: txvin.Address,
			Value:    txvin.Value,
		})
	}
	for _, usdtTx := range txusdts {
		if s.conf.Name != "btc" {
			continue
		}
		pushUtxoTx.Contract = append(pushUtxoTx.Contract, &bo.PushContractTx{
			Coin:    "usdt",
			From:    usdtTx.Sendingaddress,
			To:      usdtTx.Referenceaddress,
			Amount:  usdtTx.Amount,
			Fee:     blocktx.Fee,
			Valid:   usdtTx.Valid,
			FeeCoin: s.conf.Name,
		})
	}
	pushBlockTx.Txs = append(pushBlockTx.Txs, pushUtxoTx)
	pushdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return err
	}
	log.Info("push", string(pushdata))
	dingding.ErrorDingBot.Notify(pushdata)
	if s.pusher != nil {
		s.pusher.AddPushTask(pushBlockTx.Height, pushUtxoTx.Txid, tmpWatchList, pushdata)
	}
	return nil
}
func (p *Processor) parseContractTX(txs *TxInfo, watchaddrs map[string]bool) (vouts []*dao.BlockTxVout, vins []*dao.BlockTxVin, usdtxs []*dao.BtcUsdtTx, err error) {
	//txj, _ := json.Marshal(txs)
	//log.Info(string(txj))
	for _, vin := range txs.Vins {
		if p.watch.IsWatchAddressExist(vin.Address) {
			vins = append(vins, vin)
			watchaddrs[vin.Address] = true
		}
	}
	for _, vout := range txs.Vouts {
		if p.watch.IsWatchAddressExist(vout.Address) {
			vouts = append(vouts, vout)
			watchaddrs[vout.Address] = true
		}
	}
	for _, usdtx := range txs.Contractxs {
		if p.watch.IsWatchAddressExist(usdtx.Sendingaddress) || p.watch.IsWatchAddressExist(usdtx.Referenceaddress) {
			usdtxs = append(usdtxs, usdtx)
			if p.watch.IsWatchAddressExist(usdtx.Sendingaddress) {
				watchaddrs[usdtx.Sendingaddress] = true
			} else {
				watchaddrs[usdtx.Referenceaddress] = true
			}
		}
	}
	//log.Info(vouts, len(vouts), xutils.String(watchaddrs))
	return
}
func (p *Processor) UpdateAmount(addr string) error {

	return nil
}

func (s *Processor) RepushTx(userid int64, txid string) error {
	var (
		err    error
		txinfo *TxInfo
	)

	log.Infof("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	if txinfo, err = s.getBlockTxInfosFromNode(txid); err != nil {
		return fmt.Errorf("don't get block txinfos %v", err)
	}
	if txinfo == nil {
		return fmt.Errorf("txid:%v 不符合过滤条件", txid)
	}
	bestBlockHeight, err := s.BtcGetBlockCount()
	if err != nil {
		return err
	}
	log.Info(xutils.String(txinfo))
	watchaddrs := make(map[string]bool)
	s.parseContractTX(txinfo, watchaddrs)

	if len(watchaddrs) == 0 {
		return fmt.Errorf("txid %s don't have care of address", txid)
	}

	return s.processPush(txinfo.Tx, txinfo.Vouts, txinfo.Vins, txinfo.Contractxs, watchaddrs, bestBlockHeight.Result)
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
func (s *Processor) getBlockTxInfosFromNode(txid string) (*TxInfo, error) {

	tx, err := s.BtcGetrawtransaction(txid)
	if err != nil {
		return nil, fmt.Errorf("GetRawTransaction txid: %s , err: %v", txid, err)
	}
	log.Info(xutils.String(tx))
	block, err := s.BtcGetblock1(tx.Blockhash)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByHash hash: %s , err: %v", tx.Blockhash, err)
	}
	if block.Result.Height <= 674386 {
		return nil, errors.New("低于674386高度的数据暂时不允许补推")
	}
	return s.parseBlockRawTX(tx, tx.Blockhash, block.Result.Height)
}

func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := make(map[string]bool)
		p.parseContractTX(txInfo, watchAddrs)

		if len(watchAddrs) > 0 {
			ret = true
			p.processPush(task.TxInfos[k].Tx, task.TxInfos[k].Vouts, task.TxInfos[k].Vins, txInfo.Contractxs, watchAddrs, task.BestHeight)
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
		Type:          bo.PushTypeConfir,
		Height:        task.Block.Height,
		Hash:          task.Block.Hash,
		CoinName:      p.conf.Name,
		Confirmations: task.BestHeight - task.Block.Height + 1,
		Time:          task.Block.Timestamp.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	log.Info("push confirm", string(pushdata))
	dingding.ErrorDingBot.Notify(pushdata)
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}

//解析交易
func (s *Processor) parseBlockRawTX(tx *rpc.BtcTxInfo, blockhash string, height int64) (*TxInfo, error) {
	return parseBlockRawTX(s.watch, s.Usdt, s.RpcClient, tx, blockhash, height)
}
