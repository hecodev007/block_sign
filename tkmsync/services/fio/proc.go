package fio

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	gofio "github.com/fioprotocol/fio-go"
	"rsksync/common"
	"rsksync/conf"
	"rsksync/models/bo"
	dao "rsksync/models/po/fio"
	"rsksync/services"
	"strings"

	"sync"
	"time"
)

type Processor struct {
	*gofio.API
	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func (s *Processor) RepushTx2(userId int64, txid string) error {
	panic("implement me")
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	client, _, _ := gofio.NewConnection(nil, node.Url)
	return &Processor{
		API:   client,
		watch: watch,
		conf:  conf.Sync,
	}
}

func (s *Processor) RepushTxByIsInternal(userId int64, txid string, isInternal bool) error {
	panic("implement me")
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

	//blocktx, err = s.getBlockTxFromDB(txid)
	//if err != nil {
	//	blocktx, err = s.getBlockTxFromNode(txid)
	//	if err != nil {
	//		return fmt.Errorf("don't get block tx %v", err)
	//	}
	//}
	blocktx, err = s.getBlockTxFromNode(txid)
	if err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}
	info, err := s.API.GetInfo()
	if err != nil {
		return fmt.Errorf("get info err : %v", err)
	}
	bestBlockHeight := int64(info.LastIrreversibleBlockNum)
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
	c256, _ := hex.DecodeString(txid)
	tx, err := s.API.GetTransaction(c256)
	if err != nil || tx == nil {
		return nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}

	if len(tx.Traces) <= 0 {
		return nil, fmt.Errorf("GetTransactionByHash tx traces is empty")
	}

	block, err := s.GetBlockByNumOrID(fmt.Sprintf("%d", tx.BlockNum))
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}
	h, err := block.BlockID()
	if err != nil {
		log.Errorf("get block hash error")
	}
	blocktx := &dao.BlockTX{
		BlockHeight: int64(block.BlockNumber()),
		BlockHash:   h.String(),
		Txid:        tx.ID.String(),
		Status:      tx.Receipt.Status.String(),
		CoinName:    s.conf.Name,
		Timestamp:   tx.BlockTime.Time,
		CreateTime:  time.Now(),
	}

	if err := parseActionForBlocktx(s.API, blocktx, tx.Traces[0].Action); err != nil {
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
	if s.watch.IsWatchAddressExist(strings.ToLower(blocktx.FromAddress)) {
		tmpWatchList[blocktx.FromAddress] = true
	}
	if s.watch.IsWatchAddressExist(strings.ToLower(blocktx.ToAddress)) {
		tmpWatchList[blocktx.ToAddress] = true
	}

	if s.conf.FullBackup {
		/*	if blocktx.Status == "delayed" {
				if tx, err := s.GetTransactionFromThird(blocktx.Txid); err == nil && tx != nil {
					blocktx.Status = tx.Receipt.Status
				}
			}
		*/
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
			return fmt.Errorf("dont't have care of watch contract : %s", blocktx.ContractAddress)
		}

		//if blocktx.Status == "delayed" {
		//	if tx, err := s.GetTransactionFromThird(blocktx.Txid); err == nil && tx != nil {
		//		blocktx.Status = tx.Receipt.Status
		//	}
		//}

		if isStore {
			if len(tmpWatchList) > 0 {
				if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
					log.Errorf("block tx insert err: %v", err)
				}
			}
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
		Fee:      blocktx.Fee.String(),
		Amount:   blocktx.Amount.String(),
		Memo:     blocktx.Memo,
	})

	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return fmt.Errorf("marshal err %v", err)
	}

	if s.pusher != nil {
		fmt.Println(111)
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
