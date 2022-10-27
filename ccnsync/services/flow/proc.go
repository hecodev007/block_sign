package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"solsync/common"
	"solsync/common/conf"
	"solsync/common/log"
	"solsync/models/bo"
	dao "solsync/models/po/yotta"
	"solsync/services"
	rpc "solsync/utils/wtc"
	"strings"
)

type Processor struct {
	client *client.Client
	watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	flowClient, err := client.New(node.Url, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return &Processor{
		client: flowClient,
		watch:  watch,
		conf:   conf.Sync,
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
	//log.Infof(string(tj))

	for _, txInfo := range task.TxInfos {

		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)

		if p.conf.FullBackup {
			dao.InsertTx(txInfo)

		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo)
		}
		if len(watchAddrs) > 0 {
			p.processPush(task.BestHeight, watchAddrs, []*dao.BlockTx{txInfo})
		}

	}
	dao.InsertBlock(task.Block)
	return nil
}
func (s *Processor) processPush(bestHeight int64, tmpWatchList map[string]bool, blocktxs []*dao.BlockTx) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktxs[0].BlockHeight,
		Hash:          blocktxs[0].BlockHash,
		CoinName:      s.conf.Name,
		Token:         "", //blocktx.CoinName
		Confirmations: bestHeight - blocktxs[0].BlockHeight + 1,
		Time:          blocktxs[0].Timestamp.Unix(),
	}
	for _, blocktx := range blocktxs {
		pushBlockTx.Txs = append(pushBlockTx.Txs, &bo.PushAccountTx{
			Name: blocktx.CoinName,
			//Txid:     "0x" + blocktx.Txid,
			Txid:     blocktx.Txid,
			From:     blocktx.FromAddress,
			To:       blocktx.ToAddress,
			Contract: blocktx.ContractAddress,
			Fee:      blocktx.Fee.String(),
			Amount:   blocktx.Amount.String(),
			Memo:     blocktx.Memo,
		})
	}
	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return fmt.Errorf("marshal err %v", err)
	}

	if s.pusher != nil {
		s.pusher.AddPushTask(pushBlockTx.Height, blocktxs[0].Txid, tmpWatchList, pusdata)
	}
	log.Infof(string(pusdata))
	return nil
}
func (p *Processor) parseContractTX(tx *dao.BlockTx) (watchaddrs map[string]bool) {
	//txj, _ := json.Marshal(txs)
	//log.Infof(string(txj))
	watchaddrs = make(map[string]bool)
	if tx.ContractAddress != "" && !p.watch.IsContractExist(tx.ContractAddress) {
		return
	}

	if p.watch.IsWatchAddressExist(tx.FromAddress) {
		watchaddrs[tx.FromAddress] = true
	}
	if p.watch.IsWatchAddressExist(tx.ToAddress) {
		watchaddrs[tx.ToAddress] = true
	}
	return
}
func (p *Processor) parseContractTxs(txs []*dao.BlockTx) (watchaddrs map[string]bool) {
	//txj, _ := json.Marshal(txs)
	//log.Infof(string(txj))
	watchaddrs = make(map[string]bool)
	for _, tx := range txs {

		if p.watch.IsWatchAddressExist(tx.FromAddress) {
			watchaddrs[tx.FromAddress] = true
		}
		if p.watch.IsWatchAddressExist(tx.ToAddress) {
			watchaddrs[tx.ToAddress] = true
		}
	}
	return
}
func (p *Processor) UpdateAmount(addr string) error {

	return nil
}

func (s *Processor) GetBestBlockHeight() (int64, error) {
	status, err := s.client.GetLatestBlockHeader(context.Background(), true)
	if err != nil {
		log.Infof("%+v", err.Error())
		return 0, err
	}
	return int64(status.Height), err
}

func (s *Processor) RepushTxWithHeight(userId int64, txid string, height int64) error {
	log.Infof("RepushTxWithHeight user: %d , txid : %s \n", userId, txid)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	bestBlockHeight, err := s.GetBestBlockHeight()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}
	blocktxs, watchs, err := s.getBlockTxFromNode(txid, height)
	if err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}
	return s.processPush(bestBlockHeight, watchs, blocktxs)
}

func (s *Processor) RepushTx(userid int64, txid string) error {
	return nil
	//var (
	//	err     error
	//	txinfos []*dao.BlockTx
	//)
	//log.Infof("RepushTx user: %d , txid : %s", userid, txid)
	//
	//if txid == "" {
	//	return fmt.Errorf("don't allow txid is empty")
	//}
	//
	//if txinfos, err = s.getBlockTxFromNode(txid); err != nil {
	//	return fmt.Errorf("%v", err)
	//}
	//if len(txinfos) == 0 {
	//	return fmt.Errorf("txid:%v 不存在监控地址", txid)
	//}
	//bestBlockHeight, err := s.BlockNumber()
	//if err != nil {
	//	return err
	//}
	//watchaddrs := s.parseContractTxs(txinfos)
	//if len(watchaddrs) == 0 {
	//	return fmt.Errorf("txid %s don't have care of address", txid)
	//}
	//
	//return s.processPush(txinfos, watchaddrs, bestBlockHeight)
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
//				log.Infof("processTX SelectBlockTXVout txid: %s, n: %d, err: %v", txvin.Txid, txvin.Voutn, err)
//				continue
//			}
//			tx, err := s.GetRawTransaction(txvin.Txid)
//			if err != nil {
//				log.Infof(err.Error())
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
//	//log.Infof("processTX : %s ,vins : %d , vouts : %d , used time : %f 's", blocktx.Txid, len(txvins), len(insertVouts), time.Since(starttime).Seconds())
//	return tmpWatchList, updateVins, insertVins, nil
//}
func (s *Processor) getBlockTxFromNode(txid string, height int64) ([]*dao.BlockTx, map[string]bool, error) {
	if height == 0 {
		return nil, nil, errors.New("repush flow need height")
	}
	block, err := s.client.GetBlockByHeight(context.Background(), uint64(height))
	if err != nil {
		log.Infof("%+v", err.Error())
		return nil, nil, err
	}
	if strings.HasPrefix(txid, "0x") {
		txid = txid[2:]
	}

	result, err := s.client.GetTransactionResult(context.Background(), flow.HexToID(txid))
	if err != nil {
		log.Infof("%+v", err.Error())
		return nil, nil, err
	}
	if result.Status.String() != "SEALED" {
		return nil, nil, errors.New("Status is not SEALED")
	}
	if result.Error != nil {
		return nil, nil, result.Error
	}
	//
	//tx, err := s.client.GetTransaction(context.Background(), flow.HexToID(txid))
	//if err != nil {
	//	log.Infof("%+v", err.Error())
	//	return nil, nil, err
	//}

	from, to, amount, fee, err := ParseTransaction(result)
	if err != nil {
		return nil, nil, err
	}
	dAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, nil, err
	}
	blocktxs := make([]*dao.BlockTx, 0)
	watchLists := make(map[string]bool)
	blocktx := &dao.BlockTx{
		BlockHeight: height,
		//BlockHash:   "0x" + block.ID.Hex(),
		BlockHash: block.ID.Hex(),
		//Txid:        "0x" + txid,
		Txid:        txid,
		FromAddress: from,
		CoinName:    s.conf.Name,
		Timestamp:   block.Timestamp,
		Amount:      dAmount.Shift(-8),
		ToAddress:   to,
		Fee:         decimal.NewFromInt(fee).Shift(-8),
	}
	blocktxs = append(blocktxs, blocktx)
	if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
		watchLists[blocktx.FromAddress] = true
	}
	if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
		watchLists[blocktx.ToAddress] = true
	}
	if len(blocktxs) == 0 {
		return nil, nil, fmt.Errorf("dont't have care of tx")
	}
	return blocktxs, watchLists, nil
}

func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)
		if len(watchAddrs) > 0 {
			ret = true
			p.processPush(task.BestHeight, watchAddrs, []*dao.BlockTx{task.TxInfos[k]})
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

//解析交易
func (s *Processor) parseBlockRawTX(tx *rpc.Transaction, blockhash string, height int64) ([]*dao.BlockTx, error) {
	return nil, nil
}
