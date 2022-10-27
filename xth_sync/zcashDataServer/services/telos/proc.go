package telos

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"zcashDataServer/common"
	"zcashDataServer/common/log"
	"zcashDataServer/conf"
	"zcashDataServer/models/bo"
	dao "zcashDataServer/models/po/telos"
	"zcashDataServer/services"
	"zcashDataServer/utils/eos"
)

type Processor struct {
	*eos.API
	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {

	return &Processor{
		API:   eos.NewAPI(node.Url, node.RPCKey),
		watch: watch,
		conf:  conf.Sync,
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

func (s *Processor) RepushTx(userid int64, txid string) error {
	var (
		err     error
		blocktx *dao.BlockTX
	)

	log.Infof("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}

	blocktx, err = s.getBlockTxFromDB(txid)
	if err != nil {
		blocktx, err = s.getBlockTxFromNode(txid)
		if err != nil {
			return fmt.Errorf("don't get block tx %v", err)
		}
	}

	bestBlockHeight, err := s.GetBestHeight()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}

	return s.processTX(blocktx, bestBlockHeight, false)
}

func (s *Processor) Info() (string, int64, error) {
	dbheight, err := dao.GetMaxBlockIndex()
	return s.conf.Name, dbheight, err
}

func (s *Processor) Init() error {
	return nil
}

func (s *Processor) Clear() {

}

func (s *Processor) CheckIrreverseBlock(hash string) error {
	cnt, err := dao.GetBlockCountByHash(hash)
	if err != nil {
		return fmt.Errorf("get BlockCount ByHash err: %v", err)
	}

	if cnt > 0 {
		return fmt.Errorf("already have block  hash: %s , count: %d", hash, cnt)
	}
	return nil
}

func (s *Processor) ProcIrreverseTxs(tmps []interface{}, bestHeight int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}

	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		wg.Add(len(tmps))
		for _, tmp := range tmps {
			tx := tmp.(*dao.BlockTX)
			go func(w *sync.WaitGroup) {
				defer w.Done()
				err := s.processTX(tx, bestHeight, true)
				if err == nil {
				}
			}(wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.(*dao.BlockTX)
			err := s.processTX(tx, bestHeight, true)
			if err == nil {
			}
		}
	}

	return nil
}

func (s *Processor) ProcReverseTxs(tmps []interface{}, bestHeight int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}

	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		wg.Add(len(tmps))
		for _, tmp := range tmps {
			tx := tmp.(*dao.BlockTX)
			go func(w *sync.WaitGroup) {
				defer w.Done()
				err := s.processTX(tx, bestHeight, false)
				if err == nil {
				}
			}(wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.(*dao.BlockTX)
			err := s.processTX(tx, bestHeight, false)
			if err == nil {
			}
		}
	}

	return nil
}

func (s *Processor) ProcIrreverseBlock(b interface{}) error {
	block := b.(*dao.BlockInfo)
	if _, err := dao.InsertBlockInfo(block); err != nil {
		return fmt.Errorf("block %d Insert Block err : %v", block.Height, err)
	}
	return nil
}

func (s *Processor) UpdateIrreverseConfirms() {
	//查找所有未确认的区块
	//bs, err := dao.GetUnconfirmBlockInfos(s.conf.Confirmations + 6)
	//if err == nil && bs != nil && len(bs) > 0 {
	//	var ids []int64
	//	//开始同步更新确认数
	//	for _, blk := range bs {
	//		blk.Confirmations++
	//		s.confirmsPush(blk)
	//		ids = append(ids, blk.Id)
	//	}
	//	//批量更新订单确认数。
	//	if err := dao.BatchUpdateConfirmations(ids, 1); err != nil {
	//		log.Errorf("batch update confirmations err: %v", err)
	//	}
	//}
}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	s.confirmsPush(b.(*dao.BlockInfo))
}

func (s *Processor) getBlockTxFromDB(txid string) (*dao.BlockTX, error) {
	return dao.SelecBlockTxByHash(txid)
}

func (s *Processor) getBlockTxFromNode(txid string) (*dao.BlockTX, error) {
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

	blocktx := &dao.BlockTX{
		BlockHeight: int64(block.BlockNum),
		BlockHash:   block.ID.String(),
		Txid:        tx.ID.String(),
		Status:      tx.Receipt.Status,
		CoinName:    s.conf.Name,
		Timestamp:   tx.BlockTime.Time,
		CreateTime:  time.Now(),
	}

	if err := parseActionForBlocktx(blocktx, tx.Traces[0].Action); err != nil {
		return nil, fmt.Errorf("parseActionForBlocktx err: %v ", err)
	}

	if blocktx.FromAddress == "" || blocktx.ToAddress == "" {
		return nil, fmt.Errorf("tx. from : %s , to :%s", blocktx.FromAddress, blocktx.ToAddress)
	}

	return blocktx, nil
}

// 解析交易信息到db
func (s *Processor) processTX(blocktx *dao.BlockTX, bestHeight int64, isStore bool) error {

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
				log.Errorf("block tx insert err: %v", err)
			}
		}
	} else {
		if len(tmpWatchList) == 0 {
			log.Infof("dont't have care of watch address ,from: %s, to: %s", blocktx.FromAddress, blocktx.ToAddress)
			return fmt.Errorf("dont't have care of watch address ,from: %s, to: %s", blocktx.FromAddress, blocktx.ToAddress)
		}

		if _, err := s.watch.GetContract(blocktx.ContractAddress); err != nil {
			fmt.Println("GetContract:GetContract", blocktx.ContractAddress, err.Error())
			return fmt.Errorf("dont't have care of watch contract : %s", blocktx.ContractAddress)
		}

		if isStore {

			if len(tmpWatchList) > 0 {
				if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
					log.Errorf("block tx insert err: %v", err)
				}
			}
		} else {
			fmt.Println("isStore:", isStore)
		}
	}

	if blocktx.Status == "executed" {
		s.processPush(blocktx, tmpWatchList, bestHeight)
	} else {
		log.Infof("block tx %s status : %s is failed", blocktx.Txid, blocktx.Status)
	}

	return nil
}

func (s *Processor) processPush(blocktx *dao.BlockTX, tmpWatchList map[string]bool, bestHeight int64) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktx.BlockHeight,
		Hash:          blocktx.BlockHash,
		CoinName:      s.conf.Name,
		Token:         blocktx.CoinName,
		Confirmations: bestHeight - blocktx.BlockHeight + 1,
		Time:          blocktx.Timestamp.Unix(),
	}

	pushBlockTx.Txs = append(pushBlockTx.Txs, bo.PushAccountTx{
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
	return nil
}

func (s *Processor) confirmsPush(blockInfo *dao.BlockInfo) error {

	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountConfir,
		Height:        blockInfo.Height,
		Hash:          blockInfo.Hash,
		CoinName:      s.conf.Name,
		Confirmations: blockInfo.Confirmations,
		Time:          blockInfo.Timestamp.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && s.pusher != nil {
		s.pusher.AddPushUserTask(blockInfo.Height, pushdata)
	}

	return nil
}

//func (s *Processor) getBlockInfo(txid string) (*po.BlockTX, error) {
//	if txid == "" {
//		return nil, fmt.Errorf("don't allow %s", txid)
//	}
//
//	blocktx, err := dao.SelecBlockTxByHash(txid)
//	if err == nil {
//		return blocktx, fmt.Errorf("SelecBlockTxByHash err: %v ",err)
//	}
//
//	tx, err := s.GetTransactionByHash(txid)
//	if err != nil || tx == nil {
//		return nil, err
//	}
//
//	log.Infof("GetTransactionByHash %v ", tx)
//
//	block, err := s.GetBlockByNumber(tx.BlockNumber, false)
//	if err != nil {
//		return nil, fmt.Errorf("GetBlockByNumber err: %v ",err)
//	}
//
//	blocktx = &po.BlockTX{
//		BlockHeight: tx.BlockNumber,
//		BlockHash:   tx.BlockHash,
//		Txid:        tx.Hash,
//		FromAddress: tx.From,
//		Nonce:       tx.Nonce,
//		GasUsed:     tx.Gas,
//		GasPrice:    tx.GasPrice.Int64(),
//		Input:       tx.Input,
//		CoinName:    s.conf.Name,
//		Decimal:     eth.WEI,
//		Timestamp:   time.Unix(block.Timestamp, 0),
//	}
//
//	if !s.IsContractTx(tx) {
//
//		blocktx.Amount = decimal.NewFromBigInt(tx.Value, 0)
//		blocktx.ToAddress = tx.To
//		blocktx.ContractAddress = ""
//	} else {
//		toAddr, amt, err := eth.ERC20{}.ParseTransferData(tx.Input)
//		if err != nil {
//			return nil, fmt.Errorf("ParseTransferData input : %s, err: %v", tx.Input, err)
//		}
//
//		blocktx.Amount = decimal.NewFromBigInt(amt, 0)
//		blocktx.ToAddress = toAddr
//		blocktx.ContractAddress = tx.To
//	}
//
//	txReceipt, err := s.GetTransactionReceipt(blocktx.Txid)
//	if err != nil {
//		return nil, fmt.Errorf("GetTransactionReceipt err: %v ",err)
//	}
//
//	blocktx.GasUsed = txReceipt.GasUsed
//	blocktx.Status, _ = utils.ParseInt(txReceipt.Status)
//	blocktx.CreateTime = time.Now()
//	blocktx.Decimal = eth.WEI
//	if blocktx.ContractAddress != "" { //如果是代币，检测是否为关心的token
//		contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
//		if err != nil {
//			return nil, fmt.Errorf("ont't have care of watch contract : %s", blocktx.ContractAddress)
//		}
//
//		if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
//			blocktx.Status = 2
//		}
//
//		if txReceipt.Logs != nil {
//			btys, _ := json.Marshal(txReceipt.Logs)
//			blocktx.Logs = string(btys)
//		}
//
//		blocktx.CoinName = contractInfo.Name
//		blocktx.Decimal = contractInfo.Decimal
//	}
//
//	//先写入数据库
//	num, err := dao.InsertBlockTX(blocktx)
//	if num <= 0 || err != nil {
//		return nil, fmt.Errorf("block tx insert err: %v", err)
//	}
//
//	return blocktx, err
//}
