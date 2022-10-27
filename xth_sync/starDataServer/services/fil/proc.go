package fil

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"starDataServer/common"
	"starDataServer/common/conf"
	"starDataServer/common/log"
	"starDataServer/models/bo"
	dao "starDataServer/models/po/fil"
	"starDataServer/services"
	rpc "starDataServer/utils/fil"
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

	for _, txInfo := range task.TxInfos {

		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)
		if p.conf.FullBackup {
			dao.InsertTx(txInfo)
		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo)
		}

		if len(watchAddrs) > 0 {
			p.processPush(txInfo, task.GetHeight(), task.BestHeight, task.BlockChain.Cids[0]["/"], watchAddrs)
		}
	}
	if _, err := dao.InsertBlockChain(task.BlockChain); err != nil {
		panic(err.Error())
	}
	return nil
}
func (p *Processor) processPush(tx *dao.BlockTx, height, bestheight int64, blockHash string, watchlist map[string]bool) error {
	if tx == nil {
		return nil
	}

	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		CoinName:      p.conf.Name,
		Height:        height,
		Hash:          blockHash,
		Confirmations: bestheight - height + 1,
		Time:          time.Now().Unix(),
	}
	fee, _ := decimal.NewFromString(tx.Fee)
	amount, _ := decimal.NewFromString(tx.Amount)
	pushContractTx := &bo.PushAccountTx{
		Contract: "",
		Txid:     tx.Txid,
		From:     tx.FromAddress,
		To:       tx.ToAddress,
		Fee:      fee.Shift(-18).String(),
		Amount:   amount.Shift(-18).String(),
	}
	pushBlockTx.Txs = append(pushBlockTx.Txs, pushContractTx)
	pusdata, _ := json.Marshal(&pushBlockTx)

	if p.pusher != nil {
		p.pusher.AddPushTask(pushBlockTx.Height, tx.Txid, watchlist, pusdata)
	}
	return nil
}
func (p *Processor) parseContractTX(tx *dao.BlockTx) (watchAddrs map[string]bool) {
	watchAddrs = make(map[string]bool)
	if p.watch.IsWatchAddressExist(tx.FromAddress) {
		watchAddrs[tx.FromAddress] = true
		if p.watch.IsWatchAddressExist(tx.ToAddress) {
			watchAddrs[tx.ToAddress] = true
		}
	}

	if p.watch.IsWatchAddressExist(tx.ToAddress) {
		watchAddrs[tx.ToAddress] = true
	}
	return watchAddrs
}

func (p *Processor) RepushTx(userId int64, txid string, height int64) error {
	nextheight := height + 1
	nextblock, err := dao.GetBlockChain(nextheight)
	if err != nil {
		return errors.New("block not found in database")
	}
	for len(nextblock.Cids) == 0 {
		nextheight++
		nextblock, err = dao.GetBlockChain(nextheight)
		if err != nil {
			return errors.New("block not found in database")
		}
	}

	block, err := dao.GetBlockChain(height)
	if err != nil {
		return errors.New("block not found in database")
	}
	bestHeight, err := dao.MaxBlockHeight()
	if err != nil {
		return err
	}
	log.Info(nextheight, nextblock.Cids[0]["/"], txid)
	messages, err := p.RpcClient.GetParentMessages(nextblock.Cids[0]["/"])
	//str, _ := json.Marshal(messages)
	//log.Info(string(str))
	if err != nil {
		return err
	}
	receipts, err := p.RpcClient.GetParentReceipts(nextblock.Cids[0]["/"])
	if err != nil {
		return err
	}
	index := -1
	for k, _ := range messages {
		//		log.Info(messages[k].Cid["/"])
		if messages[k].Cid["/"] == txid {
			index = k
			break
		}
	}
	if index < 0 {
		return errors.New("tx not found")
	}
	if receipts[index].ExitCode != 0 {
		return errors.New("error receipt.ExitCode")
	}
	if messages[index].Message.Method != 0 {
		return errors.New("tx.method != transfer")
	}
	baseFee, _ := decimal.NewFromString(block.Parentbasefee)
	messages[index].Message.Cid = messages[index].Cid["/"]
	messages[index].Message.Fee = baseFee.Add(messages[index].Message.GasPremium).Mul(decimal.NewFromInt(messages[index].Message.GasLimit))
	//log.Info(baseFee.String(), messages[index].Message.GasPremium.String(), messages[index].Message.GasFeeCap.String(), messages[index].Message.Fee.String())
	blockTx, err := p.parseBlockRawTX("", messages[index].Message, block.Cids[0]["/"], height)
	if err != nil {
		return err
	}
	watchlist := make(map[string]bool)
	if p.watch.IsWatchAddressExist(blockTx.FromAddress) {
		watchlist[blockTx.FromAddress] = true
	}
	if p.watch.IsWatchAddressExist(blockTx.ToAddress) {
		watchlist[blockTx.ToAddress] = true
	}

	if len(watchlist) == 0 {
		return errors.New("交易不含监控地址")
	}

	return p.processPush(blockTx, height, bestHeight, block.Cids[0]["/"], watchlist)
}
func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	for _, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)
		if len(watchAddrs) > 0 {
			ret = true
			p.processPush(txInfo, task.GetHeight(), task.BestHeight, task.BlockChain.Cids[0]["/"], watchAddrs)
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
		Height:        task.BlockChain.Height,
		Hash:          task.BlockChain.Cids[0]["/"],
		CoinName:      p.conf.Name,
		Confirmations: task.BestHeight - task.BlockChain.Height + 1,
		Time:          task.BlockChain.Timestamp.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.BlockChain.Height, pushdata)
	}

}

//解析交易
func (s *Processor) parseBlockRawTX(coinName string, tx *rpc.Transaction, blockhash string, height int64) (*dao.BlockTx, error) {
	if tx == nil {
		return nil, nil
	}
	blocktx := &dao.BlockTx{
		Txid:        tx.Cid,
		BlockHeight: height,
		BlockHash:   blockhash,
		Version:     tx.Version,
		FromAddress: tx.From,
		ToAddress:   tx.To,
		Amount:      tx.Value.String(),
		Decimalmnt:  tx.Value.Shift(-18).String(),
		Fee:         tx.Fee.String(),
		Status:      "success",
		Gasfeecap:   tx.GasFeeCap.IntPart(),
		Gaslimit:    tx.GasLimit,
		Gaspremium:  tx.GasPremium.IntPart(),
		Nonce:       tx.Nonce,
		Method:      tx.Method,
		Timestamp:   time.Now(),
		Createtime:  time.Now(),
	}
	return blocktx, nil
}
