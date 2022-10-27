package neo

import (
	"github.com/shopspring/decimal"
	"strconv"
	"time"
	"xrpDataServer/common"
	"xrpDataServer/common/conf"
	"xrpDataServer/common/log"

	"encoding/json"
	"fmt"
	"xrpDataServer/models/bo"
	dao "xrpDataServer/models/po/neo"
	"xrpDataServer/services"
	rpc "xrpDataServer/utils/neo"
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
	watchAddrs := make(map[string]bool)
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		pushtx := p.parseContractTX(txInfo, watchAddrs)
		if p.conf.FullBackup {
			dao.InsertContractTx(txInfo.Contractxs)
			dao.InsertTx(txInfo.Tx)
		} else {
			dao.InsertContractTx(pushtx)
			if len(watchAddrs) > 0 {
				dao.InsertTx(txInfo.Tx)
			}
		}
		if len(watchAddrs) > 0 {
			p.processPush(task, k, pushtx, watchAddrs)
		}
	}
	dao.InsertBlock(task.Block)
	return nil
}
func (p *Processor) processPush(task *ProcTask, index int, txs []*dao.ContractTx, watchlist map[string]bool) error {
	if len(txs) == 0 {
		return nil
	}
	coinName, _, _ := p.watch.GetContractNameAndDecimal(txs[0].Contract)
	if coinName == "" {
		coinName = p.conf.Name
	}
	pushBlockTx := &bo.PushUtxoBlockInfo{
		Type:          bo.PushTypeTX,
		CoinName:      coinName,
		Height:        task.GetHeight(),
		Hash:          task.GetBlockHash(),
		Confirmations: task.BestHeight - task.GetHeight(),
		Time:          task.Block.Time.Unix(),
	}
	pushUtxoTx := &bo.PushUtxoTx{
		Txid: task.TxInfos[index].Tx.Txid,
	}
	pushBlockTx.Txs = append(pushBlockTx.Txs, pushUtxoTx)
	for _, tx := range txs {

		pushContractTx := &bo.PushContractTx{
			Contract: tx.Contract,
			From:     tx.From,
			To:       tx.To,
			Amount:   strconv.FormatInt(tx.Value, 10),
		}
		pushBlockTx.Txs[0].Contract = append(pushBlockTx.Txs[0].Contract, pushContractTx)
	}

	pusdata, _ := json.Marshal(&pushBlockTx)

	if p.pusher != nil {
		p.pusher.AddPushTask(pushBlockTx.Height, task.TxInfos[index].Tx.Txid, watchlist, pusdata)
	}
	return nil
}
func (p *Processor) parseContractTX(txs *TxInfo, watchaddrs map[string]bool) (ret []*dao.ContractTx) {
	for _, ctx := range txs.Contractxs {
		if !p.watch.IsContractExist(ctx.Contract) {
			continue
		}
		if coinName, dec, err := p.watch.GetContractNameAndDecimal(ctx.Contract); err == nil {
			ctx.Coinname = coinName
			amount := decimal.NewFromInt(ctx.Value)
			ctx.Vdecimal = amount.Shift(0 - int32(dec)).String()
		}

		if p.watch.IsWatchAddressExist(ctx.From) {
			ret = append(ret, ctx)
			watchaddrs[ctx.From] = true
		}
		if p.watch.IsWatchAddressExist(ctx.To) {
			ret = append(ret, ctx)
			watchaddrs[ctx.To] = true
		}
	}
	return
}
func (p *Processor) RepushTx(userId int64, txid string) error {
	rawTx, err := p.RpcClient.GetRawTransaction(txid)
	if err != nil {
		return err
	}
	if rawTx.BlockHash == "" {
		return fmt.Errorf("tx not found")
	}
	blockHash := rawTx.BlockHash
	block, err := p.RpcClient.GetBlockByHash(blockHash)
	if err != nil {
		return err
	}
	txinfo, err := p.parseBlockRawTX(&rawTx, blockHash, block.Height)
	if err != nil {
		return err
	}
	watchAddrs := make(map[string]bool)

	pushtx := p.parseContractTX(txinfo, watchAddrs)
	if len(pushtx) == 0 {
		return fmt.Errorf("have no watched address")
	}

	procTack := &ProcTask{
		Irreversible: rawTx.Confirmations > p.conf.Confirmations,
		BestHeight:   block.Height + rawTx.Confirmations,
		Block:        &dao.BlockInfo{Height: block.Height, Hash: blockHash, Time: time.Unix(block.Time, 0)},
		TxInfos:      []*TxInfo{txinfo},
	}
	p.processPush(procTack, 0, pushtx, watchAddrs)
	return nil
}
func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	watchAddrs := make(map[string]bool)
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		pushtx := p.parseContractTX(txInfo, watchAddrs)

		if len(watchAddrs) > 0 {
			p.processPush(task, k, pushtx, watchAddrs)
		}
	}
	return len(watchAddrs) > 0, nil
}

func (p *Processor) PushReverseConfirms(procTask common.ProcTask) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error ProcTask type")
	}
	pushBlockTx := &bo.PushUtxoBlockInfo{
		Type:          bo.PushTypeConfir,
		Height:        task.Block.Height,
		Hash:          task.Block.Hash,
		CoinName:      p.conf.Name,
		Confirmations: task.BestHeight - task.Block.Height,
		Time:          task.Block.Time.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}

func (p *Processor) parseBlockRawTX(tx *rpc.Transaction, blockhash string, height int64) (txInfo *TxInfo, err error) {

	if tx == nil || tx.Type != "InvocationTransaction" {
		return nil, fmt.Errorf("txid is null")
	}
	blockTx := &dao.BlockTx{
		Txid:      tx.Txid,
		Height:    height,
		Hash:      blockhash,
		Vincount:  len(tx.Vin),
		Voutcount: len(tx.Vout),
		Type:      tx.Type,
		Vmstate:   "HALT",
	}
	txInfo = &TxInfo{Tx: blockTx}
	//获取合约执行状态
getlog:
	txlog, err := p.RpcClient.GetTransactionLog(tx.Txid)
	if err != nil {
		log.Warn(err.Error() + tx.Txid)
		time.Sleep(time.Second * 3)
		goto getlog
	}
	if len(txlog.Executions) == 0 || txlog.Executions[0].Vmstate != "HALT" {
		return nil, fmt.Errorf("ship tx:%v", txlog.Executions[0].Vmstate)
	}
	for index, nt := range txlog.Executions[0].Notifications {
		if nt.State.Type != "Array" {
			continue
		}
		valueJson, _ := json.Marshal(nt.State.Value)
		values := make([]*rpc.Param, 0)
		if err := json.Unmarshal(valueJson, &values); err != nil {
			log.Warn(err.Error())
			continue
		}

		if len(values) != 4 || values[0].Type != "ByteArray" || values[0].Value != "7472616e73666572" {
			continue
		}
		var amount int64
		if values[3].Type == "Integer" {
			if amount, err = strconv.ParseInt(values[3].Value.(string), 10, 64); err != nil {
				return nil, err
			}
		} else if values[3].Type == "ByteArray" {
			if amount, err = bytesToInt(values[3].Value.(string)); err != nil {
				log.Warn(values[0].Value.(string) + " " + values[3].Value.(string) + "  " + err.Error())
				return nil, err
			}
		} else {
			panic(tx.Txid)
			return nil, fmt.Errorf("Unknow type:%v", values[3].Type)
		}
		contractx := &dao.ContractTx{
			Txid:     tx.Txid,
			Height:   height,
			Hash:     blockhash,
			Contract: nt.Contract,
			Vmstate:  txlog.Executions[0].Vmstate,
			Index:    index,
			From:     BytesToNeoAddr(values[1].Value.(string)),
			To:       BytesToNeoAddr(values[2].Value.(string)),
			Value:    amount,
		}
		txInfo.Contractxs = append(txInfo.Contractxs, contractx)
	}

	return txInfo, nil
}
