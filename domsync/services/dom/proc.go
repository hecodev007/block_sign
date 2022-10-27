package dom

import (
	"domsync/common"
	"domsync/common/conf"
	"domsync/models/bo"
	dao "domsync/models/po/dom"
	"domsync/services"
	"domsync/utils/dom"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"time"
)

type Processor struct {
	client *dom.RpcClient
	watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	//如果启动eth，顺便启动定制加载的合约
	//err := InitEthClient(node.Url)
	//if err != nil {
	//	panic(err)
	//}
	return &Processor{
		client: dom.NewRpcClient(node.Url),
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
	//log.Printf(string(tj))

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
	if blocktxs[0].CoinName != s.conf.Name {
		pushBlockTx.Token = blocktxs[0].CoinName
	}
	fee := decimal.NewFromBigInt(big.NewInt(blocktxs[0].GasPrice), int32(0-dom.WEI))
	for _, blocktx := range blocktxs {
		amount := blocktx.Amount.Shift(int32(0 - blocktx.Decimal)).String()
		pushBlockTx.Txs = append(pushBlockTx.Txs, &bo.PushAccountTx{
			Txid:     blocktx.Txid,
			From:     blocktx.FromAddress,
			To:       blocktx.ToAddress,
			Contract: blocktx.ContractAddress,
			Fee:      fee.String(),
			Amount:   amount,
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
	log.Printf(string(pusdata))
	return nil
}
func (p *Processor) parseContractTX(tx *dao.BlockTx) (watchaddrs map[string]bool) {
	//txj, _ := json.Marshal(txs)
	//log.Printf(string(txj))
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
	//log.Printf(string(txj))
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

func (s *Processor) RepushTxWithHeight(userId int64, txid string, height int64) error {
	return nil
}

func (s *Processor) RepushTx(userid int64, txid string) error {
	log.Printf("RepushTx user: %d , txid : %s \n", userid, txid)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	bestBlockHeight, err := s.client.BlockNumber()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}
	blocktxs, watchs, err := s.getBlockTxFromNode(txid)
	if err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}
	//return s.processPush(watchs, bestBlockHeight, blocktxs...)
	return s.processPush(bestBlockHeight, watchs, blocktxs)
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

func (s *Processor) getBlockTxFromNode(txid string) ([]*dao.BlockTx, map[string]bool, error) {
	tx, err := s.client.GetTransactionByHash(txid)
	if err != nil || tx == nil {
		return nil, nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}
	blockHash, err := s.client.GetBlockHashByHeight(tx.Height)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockHashByHeight err: %v ", err)
	}
	block, err := s.client.GetBlockByHash(blockHash, true)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}

	if tx.Receipt.Ty != 2 {
		return nil, nil, fmt.Errorf("this contract tx status err : %d ", tx.Receipt.Ty)
	}

	blocktxs := make([]*dao.BlockTx, 0)
	watchLists := make(map[string]bool)
	blocktx := &dao.BlockTx{
		BlockHeight: block.Items[0].Block.Height,
		BlockHash:   block.Items[0].Block.Txhash,
		Txid:        txid,
		FromAddress: tx.Tx.From,
		Nonce:       tx.Tx.Nonce,
		CoinName:    s.conf.Name,
		Decimal:     dom.WEI,
		Timestamp:   time.Unix(block.Items[0].Block.Blocktime, 0),
		Amount:      decimal.NewFromBigInt(big.NewInt(tx.Tx.Amount), 0),
		ToAddress:   tx.Tx.To,
		GasPrice:    tx.Tx.Fee,
		CreateTime:  time.Now(),
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
	if len(watchLists) == 0 {
		return nil, nil, fmt.Errorf("dont't have care of address")
	}
	return blocktxs, watchLists, nil
}
