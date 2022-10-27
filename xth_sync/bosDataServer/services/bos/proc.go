package telos

import (
	"bosDataServer/common"
	"bosDataServer/common/conf"
	"bosDataServer/common/log"
	"bosDataServer/models/bo"
	dao "bosDataServer/models/po/telos"
	"bosDataServer/services"
	"bosDataServer/utils/eos"
	"encoding/json"
	"fmt"
	"time"
)

type Processor struct {
	*eos.API
	watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		API:   eos.NewAPI(node.Url, node.RPCKey),
		watch: watch,
		conf:  conf.Sync,
	}
}

func (p *Processor) Init() error {
	return nil
}
func (s *Processor) Clear() {
}
func (s *Processor) SetPusher(push common.Pusher) {
	pusher, ok := push.(*services.PushServer)
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
func (s *Processor) CheckIrreverseBlock(hash string) error {
	cnt, err := dao.GetBlockCountByHash(hash)
	if err != nil {
		return fmt.Errorf("get BlockCount ByHash err: %v", err)
	}

	if cnt > 0 {
		return fmt.Errorf("already have Block  hash: %s , count: %d", hash, cnt)
	}
	return nil
}
func (s *Processor) RepushTx(userid int64, txid string) error {
	var (
		err     error
		blocktx *dao.BlockTx
	)

	log.Infof("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	//blocktx, err = s.getBlockTxFromDB(txid)
	//if err != nil {
	blocktx, err = s.getBlockTxFromNode(txid)
	if err != nil {
		return fmt.Errorf("don't get Block tx %v", err)
	}
	//}

	bestBlockHeight, err := s.GetBestHeight()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}

	return s.processTX(blocktx, bestBlockHeight, false)
}

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
			dao.InsertBlockTX(txInfo)
		} else if len(watchAddrs) > 0 {
			dao.InsertBlockTX(txInfo)
		}

		if len(watchAddrs) > 0 {
			p.processPush(task.TxInfos[k], watchAddrs, task.BestHeight)
		}

	}
	dao.InsertBlockInfo(task.Block)
	return nil
}
func (p *Processor) parseContractTX(tx *dao.BlockTx) (watchaddrs map[string]bool) {
	watchaddrs = make(map[string]bool)
	//if !p.watch.IsContractExist(tx.ContractAddress) {
	//	log.Info(tx.ContractAddress)
	//	return
	//}

	if p.watch.IsWatchAddressExist(tx.FromAddress) {
		watchaddrs[tx.FromAddress] = true
	}
	if p.watch.IsWatchAddressExist(tx.ToAddress) {
		watchaddrs[tx.ToAddress] = true
	}
	//log.Info(vouts, len(vouts))
	return
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
func (s *Processor) getBlockTxFromNode(txid string) (*dao.BlockTx, error) {
	tx, err := s.GetTransactionFromThird(txid)
	if err != nil || tx == nil {
		return nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}

	if len(tx.Traces) <= 0 {
		return nil, fmt.Errorf("GetTransactionByHash tx traces is empty")
	}

	block, err := s.GetBlockByNumOrID(tx.BlockNum)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}

	blocktx := &dao.BlockTx{
		BlockHeight: int64(block.BlockNum),
		BlockHash:   block.ID.String(),
		Txid:        tx.ID.String(),
		Status:      tx.Trx.Receipt.Status,
		CoinName:    "",
		Timestamp:   tx.BlockTime.Time,
		Createtime:  time.Now(),
	}

	if err := parseActionForBlocktx(s.watch, blocktx, tx.Traces[0].Action); err != nil {
		return nil, fmt.Errorf("parseActionForBlocktx err: %v ", err)
	}

	if blocktx.FromAddress == "" || blocktx.ToAddress == "" {
		return nil, fmt.Errorf("tx. from : %s , to :%s", blocktx.FromAddress, blocktx.ToAddress)
	}

	return blocktx, nil
}

// 解析交易信息到db
func (s *Processor) processTX(blocktx *dao.BlockTx, bestHeight int64, isStore bool) error {
	if blocktx == nil {
		return fmt.Errorf("tx is null")
	}

	//检测是否为关心的地址
	tmpWatchList := make(map[string]bool)
	if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
		tmpWatchList[blocktx.FromAddress] = true
	}

	if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
		tmpWatchList[blocktx.ToAddress] = true
	}
	if s.conf.FullBackup {
		if isStore {
			if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
				log.Errorf("Block tx insert err: %v", err)
			}
		}
	} else {
		if len(tmpWatchList) == 0 {
			//log.Infof("dont't have care of watch address ,from: %s, to: %s", blocktx.FromAddress, blocktx.ToAddress)
			return fmt.Errorf("dont't have care of watch address ,from: %s, to: %s", blocktx.FromAddress, blocktx.ToAddress)
		}

		if _, err := s.watch.GetContract(blocktx.ContractAddress); err != nil {
			fmt.Println("GetContract:GetContract", blocktx.ContractAddress, err.Error())
			return fmt.Errorf("dont't have care of watch contract : %s", blocktx.ContractAddress)
		}

		if isStore {
			if len(tmpWatchList) > 0 {
				if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
					log.Errorf("Block tx insert err: %v", err)
				}
			}
		} else {
			fmt.Println("isStore:", isStore)
		}
	}
	//	log.Error("blocktx.Status", blocktx.Status, string(a))
	if blocktx.Status == "executed" {
		s.processPush(blocktx, tmpWatchList, bestHeight)
	} else {
		log.Infof("Block tx %s status : %s is failed", blocktx.Txid, blocktx.Status)
	}

	return nil
}
func (s *Processor) processPush(blocktx *dao.BlockTx, tmpWatchList map[string]bool, bestHeight int64) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktx.BlockHeight,
		Hash:          blocktx.BlockHash,
		CoinName:      s.conf.Name,
		Token:         "", //blocktx.CoinName
		Confirmations: bestHeight - blocktx.BlockHeight + 1 + 6,
		Time:          blocktx.Timestamp.Unix(),
	}

	pushBlockTx.Txs = append(pushBlockTx.Txs, &bo.PushAccountTx{
		Txid:     blocktx.Txid,
		From:     blocktx.FromAddress,
		To:       blocktx.ToAddress,
		Contract: blocktx.ContractAddress,
		Fee:      "0",
		Amount:   blocktx.Amount.String(),
		Memo:     blocktx.Memo,
	})

	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return fmt.Errorf("marshal err %v", err)
	}

	if s.pusher != nil {
		s.pusher.AddPushTask(pushBlockTx.Height, blocktx.Txid, tmpWatchList, pusdata)
	}
	//	log.Error(string(pusdata))
	return nil
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
