package cfx

import (
	"cfxDataServer/common"
	"cfxDataServer/common/conf"
	"cfxDataServer/common/log"
	"cfxDataServer/models/bo"
	dao "cfxDataServer/models/po/cfx"
	"cfxDataServer/services"
	rpc "cfxDataServer/utils/mob"
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

	for k, txInfo := range task.TxInfos {
		watchAddrs := p.parseContractTX(txInfo)
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

	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		CoinName:      s.conf.Name,
		Height:        blocktx.BlockHeight,
		Hash:          blocktx.BlockHash,
		Confirmations: bestHeight - blocktx.BlockHeight + 1,
		Time:          blocktx.Timestamp.Unix(),
	}

	pushUtxoTx := &bo.PushAccountTx{
		Name:     blocktx.CoinName,
		Txid:     blocktx.Txid,
		Fee:      blocktx.Fee,
		From:     blocktx.FromAddress,
		To:       blocktx.ToAddress,
		Amount:   blocktx.Amount,
		Memo:     "",
		Contract: blocktx.ContractAddress,
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
func (p *Processor) parseContractTX(tx *TxInfo) (watchaddrs map[string]bool) {
	watchaddrs = make(map[string]bool)
	//txj, _ := json.Marshal(tx)
	//log.Info(string(txj))
	if p.watch.IsWatchAddressExist(tx.FromAddress) {
		watchaddrs[tx.FromAddress] = true
	}
	if p.watch.IsWatchAddressExist(tx.ToAddress) {
		watchaddrs[tx.ToAddress] = true
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
		txinfos []*TxInfo
	)

	log.Infof("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	if txinfos, err = s.getBlockTxInfosFromNode(txid); err != nil {
		log.Info(err.Error())
		return fmt.Errorf("%v", err)
	}
	if txinfos == nil {
		return fmt.Errorf("txid:%v 不符合过滤条件", txid)
	}
	//log.Info("getBlockTxInfosFromNode")
	bestBlockHeight, err := s.GetBlockCount()
	if err != nil {
		log.Info(err.Error())
		return err
	}
	str ,_ := json.Marshal(txinfos)
	log.Info(string(str))
	for _,txinfo := range txinfos {
		watchaddrs := s.parseContractTX(txinfo)
		if err != nil {
			log.Info(err.Error())
			return err
		}

		if len(watchaddrs) == 0 {
			log.Info("don't have care of ad")
			return fmt.Errorf("txid %s don't have care of address", txid)
		}

		if err = s.processPush(txinfo, watchaddrs, bestBlockHeight);err != nil {
			return err
		}

	}
	return nil
}

func (s *Processor) getBlockTxInfosFromNode(txid string) (txinfos []*TxInfo, err error) {
	//够log.Info("getBlockTxInfosFromNode start",txid)
	blockheight, err := s.RpcClient.GetBlockIndexByTxid(txid)
	if err != nil {
		return nil, fmt.Errorf("GetRawTransaction txid: %s , err: %v", txid, err)
	}
	if blockheight == 0 {
		return nil,fmt.Errorf("没获取到交易内容")
	}
	for _,monitorid := range monitorids{
		block,err := s.RpcClient.ProcessBlock(monitorid,blockheight)
		if err != nil {
			return nil,err
		}
		if len(block.Txs) == 0 {
			continue
		}
		txinfo,err :=parseBlockRawTX(block.Txs,fmt.Sprintf("%v",blockheight),blockheight)
		if txinfo != nil {
			txinfos = append(txinfos,txinfo...)
		}
	}

	return txinfos,nil
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
//func (s *Processor) parseBlockRawTX(tx *rpc.Transaction,rec blockhash string) (*TxInfo, error) {
//	return parseBlockRawTX(s.RpcClient, tx, blockhash)
//}
