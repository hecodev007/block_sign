package wtc

import (
	"encoding/json"
	"steemsync/utils/rpc"
	"steemsync/utils/rpc/transports/rpcclient"
	"steemsync/utils/rpc/types"

	"fmt"
	"github.com/onethefour/common/xutils"

	"steemsync/common"
	"steemsync/common/conf"
	"steemsync/common/log"
	"steemsync/models/bo"
	dao "steemsync/models/po/yotta"
	"steemsync/services"

	"time"

	"github.com/shopspring/decimal"
)

type Processor struct {
	*rpc.Client
	Watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	t := rpcclient.NewRpcClient(node.Url)
	clt, err := rpc.NewClient(t)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &Processor{
		Client: clt,
		Watch:  watch,
		conf:   conf.Sync,
	}
}
func (p *Processor) Init() error {

	for _, v := range p.Watch.WatchAddrs {
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
	return p.conf.CoinName, dbheight, err
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

	for _, txInfo := range task.TxInfos {

		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)

		if p.conf.FullBackup {
			dao.InsertTx(txInfo)

		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo)
		}
		if len(watchAddrs) > 0 {
			p.processPush([]*dao.BlockTx{txInfo}, watchAddrs, task.BestHeight)
		}

	}
	dao.InsertBlock(task.Block)
	return nil
}
func (s *Processor) processPush(blocktxs []*dao.BlockTx, tmpWatchList map[string]bool, bestHeight int64) error {

	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktxs[0].BlockHeight,
		Hash:          blocktxs[0].BlockHash,
		CoinName:      s.conf.CoinName,
		Token:         "", //blocktx.CoinName
		Confirmations: bestHeight - blocktxs[0].BlockHeight + 1,
		Time:          blocktxs[0].Timestamp.Unix(),
	}
	for _, blocktx := range blocktxs {
		pushBlockTx.Txs = append(pushBlockTx.Txs, &bo.PushAccountTx{
			Name:     blocktx.CoinName,
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
	log.Info(string(pusdata))
	return nil
}
func (p *Processor) parseContractTX(tx *dao.BlockTx) (watchaddrs map[string]bool) {
	//txj, _ := json.Marshal(txs)
	//log.Info(string(txj))
	watchaddrs = make(map[string]bool)
	if tx.ContractAddress != "" && !p.Watch.IsContractExist(tx.ContractAddress) {
		return
	}

	if p.Watch.IsWatchAddressExist(tx.FromAddress) {
		watchaddrs[tx.FromAddress] = true
	}
	if p.Watch.IsWatchAddressExist(tx.ToAddress) {
		watchaddrs[tx.ToAddress] = true
	}
	return
}
func (p *Processor) parseContractTxs(txs []*dao.BlockTx) (watchaddrs map[string]bool) {
	//txj, _ := json.Marshal(txs)
	//log.Info(string(txj))
	watchaddrs = make(map[string]bool)
	for _, tx := range txs {

		if p.Watch.IsWatchAddressExist(tx.FromAddress) {
			watchaddrs[tx.FromAddress] = true
		}
		if p.Watch.IsWatchAddressExist(tx.ToAddress) {
			watchaddrs[tx.ToAddress] = true
		}
	}
	return
}
func (p *Processor) UpdateAmount(addr string) error {

	return nil
}

func (s *Processor) RepushTx(params *bo.RePushRequest) error {
	var (
		err     error
		txinfos []*dao.BlockTx
	)
	log.Infof("补数据传参: %v", xutils.String(params))
	txid := params.Txid
	if txid == "" {
		return fmt.Errorf("txid 不能为空")
	}

	if txinfos, err = s.getBlockTxFromNode(txid); err != nil {
		log.Info("getBlockTxFromNode:", err)
		return fmt.Errorf("%v", err)
	}
	if len(txinfos) == 0 {
		return fmt.Errorf("txid:%v 不存在监控地址", txid)
	}
	bestBlockHeight, err := s.BlockNumber()
	if err != nil {
		return err
	}
	watchaddrs := s.parseContractTxs(txinfos)
	if len(watchaddrs) == 0 {
		return fmt.Errorf("txid %s 不含监控地址或合约", txid)
	}

	return s.processPush(txinfos, watchaddrs, bestBlockHeight)
}

func (s *Processor) BlockNumber() (int64, error) {
	//i++
	//return i, nil
	props, err := s.Client.Database.GetDynamicGlobalProperties()
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return int64(props.LastIrreversibleBlockNum), err
}

func (s *Processor) getBlockTxFromNode(txid string) ([]*dao.BlockTx, error) {
	tx, err := s.Client.Database.GetTransactionInfo(txid)
	if err != nil || tx == nil {
		return nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}
	log.Infof("GetTransactionInfo:%+v\n", *tx)
	block, err := s.Client.Database.GetBlock(uint32(tx.BlockNum))
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}
	bestHeight, err := s.BlockNumber()
	if err != nil {
		return nil, fmt.Errorf("BlockNumber err: %v ", err)
	}
	daoBlock := &dao.BlockInfo{
		Height: int64(block.Number),
		Hash:   block.TransactionMerkleRoot,
		//Previousblockhash: block.ParentHash,
		Timestamp:     *block.Timestamp.Time,
		Transactions:  len(block.Transactions),
		Confirmations: bestHeight - int64(block.Number) + 1,
		Createtime:    time.Now(),
	}
	return s.parseBlockTX(tx, daoBlock)
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
			p.processPush([]*dao.BlockTx{task.TxInfos[k]}, watchAddrs, task.BestHeight)
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
		CoinName:      p.conf.CoinName,
		Confirmations: task.BestHeight - task.Block.Height + 1,
		Time:          task.Block.Timestamp.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	log.Info(string(pushdata))
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}

// 解析交易
func (s *Processor) parseBlockTX(tx *types.TransactionInfo, block *dao.BlockInfo) (blocktxs []*dao.BlockTx, err error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}

	//08:21:03 +0000 UTC Operations:[{Type:transfer_operation Value:{From:gateiodeposit To:steemgoapi Memo: Amount:{Amount:14310 Precision:3 Nai:@@000000021}}}] Extensions:[] Signatures:[1f6c5e39b19e56f759e4a28e05f44d09c6a035cc50548ceb574e9eb2b932047720121adaa760f4b874e352dbdf5777078525605a6e26b33fb515e62472fd5d2c1c] TransactionId:47ebbcc58183804521be6eb320ecfd6be31f2cba BlockNum:65102191 TransactionNum:5}
	//Watch := s.Watch
	//rpc := s.RpcClient
	blocktxs = make([]*dao.BlockTx, 0)
	log.Info("parseBlockTX...")
	for _, op := range tx.Operations {
		operation := op.Value
		//amounts := strings.Split(operation.Amount.Amount, "")
		amount, _ := decimal.NewFromString(operation.Amount.Amount)
		amount = amount.Div(decimal.New(1, 3))
		log.Info("trans info->", operation.From, operation.To)
		log.Info("trans info->", amount, operation.Amount.Amount)
		//eth 转账交易
		if amount.GreaterThan(decimal.Zero) && (s.Watch.IsWatchAddressExist(operation.From) || s.Watch.IsWatchAddressExist(operation.To)) {
			blocktx := dao.BlockTx{
				CoinName:        conf.Cfg.Sync.CoinName,
				Txid:            tx.TransactionId,
				BlockHeight:     block.Height,
				ContractAddress: "",
				FromAddress:     operation.From,
				ToAddress:       operation.To,
				Amount:          amount,
				//Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
				BlockHash: block.Hash,
				Status:    "success",
				Timestamp: block.Timestamp,
			}
			fmt.Println("trans info->", operation.From, operation.To)
			blocktxs = append(blocktxs, &blocktx)
			return
		}
	}

	return
}
