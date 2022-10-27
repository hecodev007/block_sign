package iota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v2"
	"github.com/shopspring/decimal"
	"iotasync/common"
	"iotasync/common/conf"
	"iotasync/common/log"
	"iotasync/models/bo"
	dao "iotasync/models/po/iota"
	"iotasync/services"
	"time"
)

type Processor struct {
	//*rpc.RpcClient
	nodeAPI *iotago.NodeHTTPAPIClient
	watch   *services.WatchControl
	pusher  *services.PushServer
	conf    conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		nodeAPI: iotago.NewNodeHTTPAPIClient(node.Url),
		watch:   watch,
		conf:    conf.Sync,
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
		vouts, vins, _ := p.parseContractTX(txInfo, watchAddrs)

		if p.conf.FullBackup {
			dao.InsertTx(txInfo.Tx)
			dao.InsertTxVin(txInfo.Vins)
			dao.InsertTxVout(txInfo.Vouts)
		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo.Tx)
			dao.InsertTxVin(vins)
			dao.InsertTxVout(vouts)
		}

		if len(watchAddrs) > 0 {
			p.processPush(task.TxInfos[k].Tx, task.TxInfos[k].Vouts, task.TxInfos[k].Vins, watchAddrs, task.BestHeight)
		}

	}
	dao.InsertBlock(task.Block)
	return nil
}
func (s *Processor) processPush(blocktx *dao.BlockTx, txvouts []*dao.BlockTxVout, txvins []*dao.BlockTxVin, tmpWatchList map[string]bool, bestHeight int64) error {
	if len(txvins) == 0 && len(txvouts) == 0 {
		log.Error(tmpWatchList)
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
		value, _ := decimal.NewFromString(txvout.Value)
		pushUtxoTx.Vout = append(pushUtxoTx.Vout, &bo.PushTxOutput{
			Addresse: txvout.Address,
			Value:    value.Div(decimal.New(1, 6)).String(),
			//N:        txvout.VoutN,
		})
	}

	for _, txvin := range txvins {
		value, _ := decimal.NewFromString(txvin.Value)
		pushUtxoTx.Vin = append(pushUtxoTx.Vin, &bo.PushTxInput{
			Txid: txvin.Txid,
			//Vout:     txvin.VoutN,
			Addresse: txvin.Address,
			Value:    value.Div(decimal.New(1, 6)).String(),
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
func (p *Processor) parseContractTX(txs *TxInfo, watchaddrs map[string]bool) (vouts []*dao.BlockTxVout, vins []*dao.BlockTxVin, err error) {
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
	//log.Info(vouts, len(vouts))
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

	log.Info("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	if txinfo, err = s.getBlockTxInfosFromNode(txid); err != nil {
		return fmt.Errorf("don't get block txinfos %v", err)
	}

	bestBlockHeight, err := s.getBlockCount()
	if err != nil {
		return err
	}
	watchaddrs := make(map[string]bool)
	_, _, err = s.parseContractTX(txinfo, watchaddrs)
	if err != nil {
		return err
	}

	if len(watchaddrs) == 0 {
		return fmt.Errorf("txid %s don't have care of address", txid)
	}

	return s.processPush(txinfo.Tx, txinfo.Vouts, txinfo.Vins, watchaddrs, bestBlockHeight)
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
	//nodeAPI := iotago.NewNodeHTTPAPIClient("https://api.lb-0.h.chrysalis-devnet.iota.cafe:443")
	//"7deef0a8a2868aced7e3ff7722cef2295f1120038a82f7bf46408f60da8e804c"
	messageId, err := iotago.MessageIDFromHexString(txid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	//ms, err := nodeAPI.MessageJSONByMessageID(context.Background(), messageId)
	//fmt.Printf("%+v\n", ms.Payload)
	//conflicting
	//included
	resp, err := s.nodeAPI.MessageMetadataByMessageID(context.Background(), messageId)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if *resp.LedgerInclusionState != "included" {
		return nil, errors.New("conflicting transaction")
	}

	message, err := s.nodeAPI.MessageByMessageID(context.Background(), messageId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	//fmt.Printf("%+v\n", message.Payload)

	tx, ok := message.Payload.(*iotago.Transaction)
	if !ok {
		return nil, errors.New("not transaction")
	}
	ess := tx.Essence.(*iotago.TransactionEssence)
	stone, err := s.nodeAPI.MessageMetadataByMessageID(context.Background(), messageId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	fmt.Printf("%d\n", *stone.ReferencedByMilestoneIndex)
	block, err := s.nodeAPI.MilestoneByIndex(context.Background(), *stone.ReferencedByMilestoneIndex)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return s.parseBlockRawTX(ess.Outputs, ess.Inputs, txid, block.MessageID, int64(*stone.ReferencedByMilestoneIndex), block.Time)
}

func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := make(map[string]bool)
		_, _, _ = p.parseContractTX(txInfo, watchAddrs)

		if len(watchAddrs) > 0 {
			ret = true
			p.processPush(task.TxInfos[k].Tx, task.TxInfos[k].Vouts, task.TxInfos[k].Vins, watchAddrs, task.BestHeight)
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
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}

func (s *Processor) getBlockCount() (int64, error) {

	info, err := s.nodeAPI.Info(context.Background())
	return int64(info.ConfirmedMilestoneIndex), err
}

//解析交易
func (s *Processor) parseBlockRawTX(Outputs, Inputs serializer.Serializables, txId, blockhash string, height, timestap int64) (txInfo *TxInfo, err error) {

	blockTx := &dao.BlockTx{
		Txid:      txId,
		Blockhash: blockhash,
		Height:    height,
		//Vincount:   len(ConsumeOutPuts),
		//Voutcount:  len(CreatedOutputs),
		Createtime: time.Now(),
		Timestamp:  time.Unix(timestap, 0),
	}
	blockTx.Vincount = len(Inputs)
	blockTx.Voutcount = len(Outputs)
	txInfo = &TxInfo{Tx: blockTx}
	for _, o := range Outputs {
		output, ok := o.(iotago.Output)
		if ok {
			tar, err := output.Target()
			if err != nil {
				log.Error(err)
				continue
			}
			addr := fmt.Sprintf("%v", tar)
			edAddr, err := iotago.ParseEd25519AddressFromHexString(addr)
			value, err := output.Deposit()
			if err != nil {
				log.Error(err)
				continue
			}
			txOut := &dao.BlockTxVout{
				Height:     height,
				Address:    edAddr.Bech32(iotago.PrefixMainnet),
				Value:      fmt.Sprintf("%v", value),
				Txid:       txId,
				MessageId:  txId,
				Createtime: time.Now(),
			}
			txInfo.Vouts = append(txInfo.Vouts, txOut)
		}
	}
	//
	for _, o := range Inputs {
		h, ok := o.(*iotago.UTXOInput)
		if ok {
			//fmt.Println(h.TransactionOutputIndex)
			//fmt.Println(h.ID().ToHex())
			utxoInput, err := iotago.OutputIDHex(h.ID().ToHex()).AsUTXOInput()
			if err != nil {
				log.Error(err)
				continue
			}
			outputRes, err := s.nodeAPI.OutputByID(context.Background(), utxoInput.ID())
			output, err := outputRes.Output()
			tar, err := output.Target()
			if err != nil {
				log.Error(err)
				continue
			}
			addr := fmt.Sprintf("%v", tar)
			edAddr, err := iotago.ParseEd25519AddressFromHexString(addr)

			value, err := output.Deposit()
			if err != nil {
				log.Error(err)
				continue
			}
			txVin := &dao.BlockTxVin{
				Height:      height,
				Address:     edAddr.Bech32(iotago.PrefixMainnet),
				Value:       fmt.Sprintf("%v", value),
				Txid:        outputRes.TransactionID,
				OutputIndex: outputRes.OutputIndex,
				LedgerIndex: outputRes.LedgerIndex,
				MessageId:   outputRes.MessageID,
				Spent:       outputRes.Spent,
				Createtime:  time.Now(),
			}
			txInfo.Vins = append(txInfo.Vins, txVin)
		}
	}

	return txInfo, nil
}
