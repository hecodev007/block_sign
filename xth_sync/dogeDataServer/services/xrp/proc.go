package xrp

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"strconv"
	"dogeDataServer/common"
	"dogeDataServer/common/conf"
	"dogeDataServer/common/log"
	"dogeDataServer/models/bo"
	dao "dogeDataServer/models/po/xrp"
	"dogeDataServer/services"
	rpc "dogeDataServer/utils/xrp"
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
	watchAddrs := make(map[string]bool)
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		pushtx := p.parseContractTX(txInfo, watchAddrs)
		if p.conf.FullBackup {
			dao.InsertTokenTx(txInfo.Contractxs)
			dao.InsertTx(txInfo.Tx)
		} else {
			dao.InsertTokenTx(pushtx)
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
func (p *Processor) processPush(task *ProcTask, index int, txs []*dao.TokenTx, watchlist map[string]bool) error {
	if len(txs) == 0 {
		return nil
	}
	coinName, _, _ := p.watch.GetContractNameAndDecimal(txs[0].Contract)
	if coinName == "" {
		coinName = p.conf.Name
	}
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeTX,
		CoinName:      p.conf.Name,
		Height:        task.GetHeight(),
		Hash:          task.GetBlockHash(),
		Confirmations: task.BestHeight - task.GetHeight() + 1,
		Time:          task.Block.Time.Unix(),
	}
	//pushUtxoTx := &bo.PushUtxoTx{
	//	Txid: task.TxInfos[index].Tx.Txid,
	//}
	//pushBlockTx.Txs = append(pushBlockTx.Txs, pushUtxoTx)
	for _, tx := range txs {
		memo := ""
		if tx.Memo != 0 {
			memo = strconv.FormatInt(tx.Memo, 10)
		}
		pushContractTx := &bo.PushAccountTx{
			Contract: tx.Contract,
			From:     tx.From,
			To:       tx.To,
			Amount:   strconv.FormatInt(tx.Value, 10),
			Memo:     memo,
		}
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushContractTx)
	}

	pusdata, _ := json.Marshal(&pushBlockTx)

	if p.pusher != nil {
		p.pusher.AddPushTask(pushBlockTx.Height, task.TxInfos[index].Tx.Txid, watchlist, pusdata)
	}
	return nil
}
func (p *Processor) parseContractTX(txs *TxInfo, watchaddrs map[string]bool) (ret []*dao.TokenTx) {
	for _, ctx := range txs.Contractxs {

		amount := decimal.NewFromInt(ctx.Value)
		ctx.Vdecimal = amount.Shift(-6).String()

		if p.watch.IsWatchAddressExist(ctx.From) {
			ret = append(ret, ctx)
			watchaddrs[ctx.From] = true
			go p.UpdateAmount(ctx.From)
		}
		if p.watch.IsWatchAddressExist(ctx.To) {
			ret = append(ret, ctx)
			watchaddrs[ctx.To] = true
			go p.UpdateAmount(ctx.To)
		}
	}
	return
}
func (p *Processor) UpdateAmount(addr string) error {
	value, dvalue, blockHeight, err := p.RpcClient.GetBalance(addr)
	if err != nil {
		log.Warn(err.Error())
		return err
	}
	dao.UpdateAmount(addr, value, dvalue, blockHeight)
	if err != nil {
		log.Warn(err.Error())
	}
	return err
}
func (p *Processor) RepushTx(userId int64, txid string) error {
	//https://s1.ripple.com:51234
	//https://s2.ripple.com:51234
	rpcClient := rpc.NewRpcClient("https://s1.ripple.com:51234", "", "")
	rawTx, err := rpcClient.GetTracsaction(txid)
	if err != nil {
		return err
	}
	bestHeight, err := rpcClient.BlockHeight()
	if err != nil {
		return err
	}
	blockHeight := rawTx.LedgerIndex
	block, err := rpcClient.GetBlock(blockHeight)
	if err != nil {
		return err
	}
	txinfo, err := p.parseBlockRawTX(rawTx, block.Hash, block.Height)
	if err != nil {
		return err
	}
	watchAddrs := make(map[string]bool)

	pushtx := p.parseContractTX(txinfo, watchAddrs)
	if len(pushtx) == 0 {
		return fmt.Errorf("have no watched address")
	}

	procTack := &ProcTask{
		Irreversible: bestHeight-rawTx.LedgerIndex >= p.conf.Confirmations,
		BestHeight:   bestHeight,
		Block:        &dao.BlockInfo{Height: block.Height, Hash: block.Hash, Time: block.Time},
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
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeConfir,
		Height:        task.Block.Height,
		Hash:          task.Block.Hash,
		CoinName:      p.conf.Name,
		Confirmations: task.BestHeight - task.Block.Height + 1,
		Time:          task.Block.Time.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}

//解析交易
func (s *Processor) parseBlockRawTX(tx *rpc.Transaction, blockhash string, height int64) (txInfo *TxInfo, err error) {

	if tx == nil || tx.TransactionType != "Payment" || tx.Meta.TransactionResult != "tesSUCCESS" || tx.Flags != 2147483648 {
		return nil, nil
	}
	vinCount := 0
	switch v := tx.Meta.DeliveredAmount.(type) {
	case string:
		vinCount = 1
	case []interface{}:
		vinCount = len(v)
	default:
		return nil, fmt.Errorf("unkonw account type:%v", tx.Hash)
	}
	fee, _ := strconv.ParseInt(tx.Fee, 10, 64)
	blockTx := &dao.BlockTx{
		Txid:     tx.Hash,
		Height:   height,
		Hash:     blockhash,
		Vincount: vinCount,
		Memo:     tx.DestinationTag,
		From:     tx.Account,
		To:       tx.Destination,
		Type:     tx.TransactionType,
		State:    tx.Meta.TransactionResult,
		Fee:      fee,
	}
	txInfo = &TxInfo{Tx: blockTx}
	//获取合约执行状态
	switch v := tx.Meta.DeliveredAmount.(type) {
	case string:
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		tokentx := &dao.TokenTx{
			Contract: "xrp",
			Txid:     tx.Hash,
			Height:   height,
			Hash:     blockhash,
			Vmstate:  tx.Meta.TransactionResult,
			Index:    0,
			From:     tx.Account,
			To:       tx.Destination,
			Value:    value,
			Memo:     tx.DestinationTag,
			Coinname: "xrp",
		}
		txInfo.Contractxs = append(txInfo.Contractxs, tokentx)

	default:
		return nil, nil
	}

	return txInfo, nil
}
