package ghost

import (
	"encoding/json"
	"fmt"
	"marsDataServer/common"
	"marsDataServer/common/log"
	"marsDataServer/common/conf"
	"marsDataServer/models/bo"
	dao "marsDataServer/models/po/ghost"
	"marsDataServer/services"
	"marsDataServer/utils/btc"
	"github.com/shopspring/decimal"
	"sync"
)

type Processor struct {
	*btc.RpcClient
	wg   *sync.WaitGroup
	lock *sync.Mutex

	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {

	return &Processor{
		RpcClient: btc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
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


func (s *Processor) Info() (string, int64, error) {
	dbheight, err := dao.MaxBlockHeight()
	return s.conf.Name, dbheight, err
}

func (s *Processor) Init() error {
	return nil
}

func (s *Processor) Clear() {
}

func (s *Processor) CheckIrreverseBlock(hash string) error {
	if has, err := dao.BlockHashExist(hash); err != nil {
		return fmt.Errorf("get BlockCount ByHash err: %v", err)
	} else if has {
		return fmt.Errorf("already have block  hash: %s , count: %d", hash, 1)
	}
	return nil
}

//处理不可逆交易
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
		vouts, _, _ := p.parseContractTX(txInfo, watchAddrs)

		if p.conf.FullBackup {
			dao.InsertTx(txInfo.Tx)
			dao.InsertTxVout(txInfo.Vouts)
		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo.Tx)
			dao.InsertTxVout(vouts)
		}

		if len(watchAddrs) > 0 {
			p.processPush(task.TxInfos[k].Tx, task.TxInfos[k].Vouts, task.TxInfos[k].Vins, watchAddrs, task.BestHeight)
		}
	}
	dao.InsertBlock(task.Block)
	return nil
}
func (s *Processor) processPush(blocktx *dao.BlockTx, txvouts []*dao.BlockTxVout, txvins []*dao.BlockTxVout, tmpWatchList map[string]bool, bestHeight int64) error {
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
		Txid: blocktx.Txid,
		Fee:  blocktx.Fee,
	}

	for _, txvout := range txvouts {
		pushUtxoTx.Vout = append(pushUtxoTx.Vout, &bo.PushTxOutput{
			Addresse: txvout.Address,
			Value:    txvout.Value,
			N:        txvout.VoutN,
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
func (p *Processor) parseContractTX(txs *TxInfo, watchaddrs map[string]bool) (vouts []*dao.BlockTxVout, vins []*dao.BlockTxVout, err error) {
	//txj, _ := json.Marshal(txs)
	//log.Info(string(txj))
	inAmount := decimal.NewFromInt(0)
	outAmount := decimal.NewFromInt(0)
	for _, vin := range txs.Vins {
		tmpAmount,err := decimal.NewFromString(vin.Value)
		if err == nil {
			inAmount = inAmount.Add(tmpAmount)
		}
		if p.watch.IsWatchAddressExist(vin.Address) {
			vins = append(vins, vin)
			watchaddrs[vin.Address] = true
		}

	}
	for _, vout := range txs.Vouts {
		tmpAmount,err := decimal.NewFromString(vout.Value)
		if err == nil {
			outAmount = outAmount.Add(tmpAmount)
		}
		if p.watch.IsWatchAddressExist(vout.Address) {
			vouts = append(vouts, vout)
			watchaddrs[vout.Address] = true
		}
	}
	fee := inAmount.Sub(outAmount)
	if fee.GreaterThan(decimal.NewFromInt(0)){
		txs.Tx.Fee=fee.String()
	}
	return
}
//处理可逆交易
func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	//bj,_ :=json.Marshal(task)
	//fmt.Println(string(bj))
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := make(map[string]bool)
		vouts, vins, _ := p.parseContractTX(txInfo, watchAddrs)

		if len(watchAddrs) > 0 {
			ret = true
			p.processPush(task.TxInfos[k].Tx, vouts, vins, watchAddrs, task.BestHeight)
		}
	}
	return ret, nil
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

	if txinfo,_, err = s.getBlockTxInfosFromNode(txid); err != nil {
		return fmt.Errorf("don't get block txinfos %v", err)
	}

	bestBlockHeight, err := s.GetBlockCount()
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
func (s *Processor) getBlockTxInfosFromNode(txid string) (*TxInfo, int64, error) {

	tx, err := s.GetRawTransaction(txid)
	if err != nil {
		return nil, 0, fmt.Errorf("GetRawTransaction txid: %s , err: %v", txid, err)
	}

	block, err := s.GetBlockByHash(tx.BlockHash)
	if err != nil {
		return nil, 0, fmt.Errorf("GetBlockByHash hash: %s , err: %v", tx.BlockHash, err)
	}

	txinfo, err := parseBlockRawTX(s.RpcClient,&tx, tx.BlockHash, block.Height)
	return txinfo, block.Confirmations, err
}
