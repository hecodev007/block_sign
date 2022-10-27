package crust

import (
	"crustDataServer/common"
	"crustDataServer/common/conf"
	"crustDataServer/common/log"
	"crustDataServer/models/bo"
	dao "crustDataServer/models/po/crust"
	"crustDataServer/services"
	rpc "crustDataServer/utils/crust"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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
	//for _, v := range p.watch.WatchAddrs {
	//	go p.UpdateAmount(v[0].Address)
	//}
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
	if blocktx == nil {
		return nil
	}
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeTX,
		CoinName:      s.conf.Name,
		Height:        blocktx.Height,
		Hash:          blocktx.Hash,
		Confirmations: bestHeight - blocktx.Height + 1,
		Time:          time.Now().Unix(),
	}

	pushUtxoTx := &bo.PushAccountTx{
		Name:   s.conf.Name,
		Txid:   blocktx.Txid,
		Fee:    blocktx.Fee,
		From:   blocktx.Fromaccount,
		To:     blocktx.Toaccount,
		Amount: blocktx.Amount,
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
func (p *Processor) parseContractTX(tx *dao.BlockTx, watchaddrs map[string]bool) (err error) {
	//txj, _ := json.Marshal(txs)
	//log.Info(string(txj))
	if p.watch.IsWatchAddressExist(tx.Toaccount) {
		watchaddrs[tx.Toaccount] = true
	}
	if p.watch.IsWatchAddressExist(tx.Fromaccount) {
		watchaddrs[tx.Fromaccount] = true
	}
	return
}
func (p *Processor) UpdateAmount(addr string) error {

	return nil
}

func (s *Processor) RepushTx(userid int64, txid string, height int64) error {
	var (
		err    error
		txinfo *dao.BlockTx
	)

	log.Info("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	if txinfo, err = s.getBlockTxInfosFromNode(txid, height); err != nil {
		return fmt.Errorf("don't get block txinfos %v", err)
	}

	bestBlockHeight, err := s.BlockHeight()
	if err != nil {
		return err
	}
	watchaddrs := make(map[string]bool)
	s.parseContractTX(txinfo, watchaddrs)
	if err != nil {
		return err
	}

	if len(watchaddrs) == 0 {
		return fmt.Errorf("txid %s don't have care of address", txid)
	}

	return s.processPush(txinfo, watchaddrs, bestBlockHeight)
}

func (s *Processor) getBlockTxInfosFromNode(txid string, blockheight int64) (*dao.BlockTx, error) {

	block, err := s.GetBlockByHeight(blockheight)
	if err != nil {
		return nil, fmt.Errorf("GetRawTransaction txid: %s , err: %v", txid, err)
	}
	ret, err := s.RpcClient.GetMetadata()
	if err != nil {
		return nil, err
	}
	for _, rawtx := range block.Block.Extrinsics {
		tx, err := rpc.HexToTransaction(&ret.Metadata, rawtx)
		if err != nil {
			return nil, err
		}
		if tx.Txid == tx.Txid {
			return s.parseBlockRawTX(tx, block.Hash, blockheight)
		}
	}
	return nil, errors.New("txid not found")
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
		Type:          bo.PushTypeConfir,
		Height:        task.Block.Height,
		Hash:          task.Block.Hash,
		CoinName:      p.conf.Name,
		Confirmations: task.BestHeight - task.Block.Height + 1,
		Time:          time.Now().Unix(),
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
