package nas

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"rsksync/common"
	"rsksync/common/log"
	"rsksync/conf"
	"rsksync/models/bo"
	dao "rsksync/models/po/nas"
	"rsksync/services"
	"rsksync/utils/nas"
	"sync"
	"time"
)

type Processor struct {
	*nas.NasHttpClient
	watch     *services.WatchControl
	pusher    *services.PushServer
	procTasks chan common.ProcTask

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	c, err := nas.NewNasHttpClient(node.Url)
	if err != nil {
		log.Infof("NewNasClient err: %v", err)
		return nil
	}
	return &Processor{
		NasHttpClient: c,
		watch:         watch,
		procTasks:     make(chan common.ProcTask, 10000),
		conf:          conf.Sync,
	}
}

func (s *Processor) RepushTxByIsInternal(userId int64, txid string, isInternal bool) error {
	panic("implement me")
}

func (s *Processor) AddProcTask(t common.ProcTask) {
	s.procTasks <- t
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

	if txid == "" {
		return fmt.Errorf("don't allow %s", txid)
	}

	//if blocktx, err = s.getBlockTxFromDB(txid); err != nil {
	//	if blocktx, err = s.getBlockTxFromNode(txid); err != nil {
	//		return fmt.Errorf("don't get block tx %v", err)
	//	}
	//}
	if blocktx, err = s.getBlockTxFromNode(txid); err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}

	//dbHeight, err := dao.GetMaxBlockIndex()
	//if err != nil {
	//	return fmt.Errorf("GetMaxBlockIndex height: %d , err: %v", dbHeight, err)
	//}
	//
	//if blocktx.BlockHeight > dbHeight {
	//	return fmt.Errorf("don't sync reach %d, current %d", blocktx.BlockHeight, dbHeight)
	//}

	res, err := s.GetNebState()
	if err != nil {
		return err
	}
	bestBlockHeight := res.Height

	return s.processTX(blocktx, int64(bestBlockHeight))
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

//处理不可逆交易
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
				if err := s.processTX(tx, bestHeight); err == nil {
					if num, err := dao.InsertBlockTX(tx); num <= 0 || err != nil {
					}
				}
			}(wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.(*dao.BlockTX)
			if err := s.processTX(tx, bestHeight); err == nil {
				if num, err := dao.InsertBlockTX(tx); num <= 0 || err != nil {
				}
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
				if err := s.processTX(tx, bestHeight); err == nil {
				}
			}(wg)
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.(*dao.BlockTX)
			if err := s.processTX(tx, bestHeight); err == nil {
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
	if bs, err := dao.GetUnconfirmBlockInfos(s.conf.Confirmations + 6); err == nil && bs != nil && len(bs) > 0 {
		var ids []int64
		//开始同步更新确认数
		for _, blk := range bs {
			blk.Confirmations++
			s.confirmsPush(blk)
			ids = append(ids, blk.Id)
		}
		//批量更新订单确认数。
		if err := dao.BatchUpdateConfirmations(ids, 1); err != nil {
			log.Errorf("batch update confirmations err: %v", err)
		}
	}
}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	s.confirmsPush(b.(*dao.BlockInfo))
}

func (s *Processor) getBlockTxFromDB(txid string) (*dao.BlockTX, error) {
	return dao.SelecBlockTxByHash(txid)
}

func (s *Processor) getBlockTxFromNode(txid string) (*dao.BlockTX, error) {

	tx, err := s.GetTransactionReceipt(txid)
	if err != nil || tx == nil {
		return nil, err
	}
	if tx.ChainId != 1 {
		return nil, fmt.Errorf("txid:%s,非主链交易")
	}
	if tx.Status != 1 {
		return nil, fmt.Errorf("txid:%s,失败交易")
	}
	block, err := s.GetBlockByHeight(tx.BlockHeight, false)
	if err != nil {
		return nil, err
	}

	tx.BlockHash = block.Hash

	blocktx := &dao.BlockTX{
		BlockHeight:   int64(tx.BlockHeight),
		BlockHash:     block.Hash,
		Txid:          tx.Hash,
		FromAddress:   tx.From,
		Nonce:         tx.Nonce,
		GasUsed:       tx.GasUsed,
		GasPrice:      tx.GasPrice.Int64(),
		Type:          tx.Type,
		Data:          tx.Data,
		CoinName:      s.conf.Name,
		Decimal:       nas.WEI,
		Timestamp:     time.Unix(block.Timestamp, 0),
		ExecuteResult: tx.ExecuteResult,
		ExecuteError:  tx.ExecuteError,
		Status:        tx.Status,
	}

	switch tx.Type {
	case nas.TxCall:
		log.Infof("parse : %s ,txdata : %s ", tx.Hash, tx.Data)
		calldata, err := nas.ParseCallData([]byte(tx.Data))
		if err != nil {
			return nil, fmt.Errorf("ParseCallData input : %s, err: %v", blocktx.Data, err)
		}

		if calldata.Function != "transfer" {
			return nil, fmt.Errorf("ParseCallData input : %s, err: %v", blocktx.Data, fmt.Errorf("don't know transfer fuction %s", calldata.Function))
		}

		addr, amt, err := nas.ParseTransferData([]byte(calldata.Args))
		if err != nil {
			return nil, fmt.Errorf("ParseTransferData Args : %s, err: %v", calldata.Args, err)
		}

		blocktx.Amount = amt
		blocktx.ToAddress = addr
		blocktx.ContractAddress = tx.To
		break
	case nas.TxDeploy:
		break
	case nas.TxDip:
		log.Infof("tx type :%s", tx.Type)
		break
	case nas.TxProtocol:
		log.Infof("tx type :%s", tx.Type)
		break
	case nas.TxNormal:
		blocktx.Amount = decimal.NewFromBigInt(tx.Value, 0)
		blocktx.ToAddress = tx.To
		blocktx.ContractAddress = ""
		break
	default:
		return nil, fmt.Errorf("ParseTransferData input : %s, err: %v", blocktx.Data, fmt.Errorf("don't know type %s", tx.Type))
	}

	if blocktx.ContractAddress != "" { //如果是代币，检测是否为关心的token
		contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
		if err != nil {
			return nil, fmt.Errorf("dont't have care of watch contract : %s", blocktx.ContractAddress)
		}
		blocktx.CoinName = contractInfo.Name
		blocktx.Decimal = contractInfo.Decimal
	}

	if blocktx.Status != 1 {
		return nil, fmt.Errorf("block tx status : %d is failed", blocktx.Status)
	}

	return blocktx, err
}

// 解析交易信息到db
func (s *Processor) processTX(blocktx *dao.BlockTX, bestHeight int64) error {

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

	if len(tmpWatchList) == 0 {
		return fmt.Errorf("dont't have care of watch address ,from: %s, to: %s", blocktx.FromAddress, blocktx.ToAddress)
	}

	blocktx.CreateTime = time.Now()
	if blocktx.ContractAddress != "" { //如果是代币，检测是否为关心的token
		contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
		if err != nil {
			return fmt.Errorf("ont't have care of watch contract : %s", blocktx.ContractAddress)
		}

		blocktx.CoinName = contractInfo.Name
		blocktx.Decimal = contractInfo.Decimal
	}

	if blocktx.Status == 1 {
		s.processPush(blocktx, tmpWatchList, bestHeight)
	}

	return nil
}

func (s *Processor) processPush(blocktx *dao.BlockTX, tmpWatchList map[string]bool, bestHeight int64) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        int64(blocktx.BlockHeight),
		Hash:          blocktx.BlockHash,
		CoinName:      s.conf.Name,
		Token:         "",
		Confirmations: bestHeight - blocktx.BlockHeight + 1,
		Time:          blocktx.Timestamp.Unix(),
	}

	if blocktx.CoinName != s.conf.Name {
		pushBlockTx.Token = blocktx.CoinName
	}

	fee := decimal.New(blocktx.GasPrice, 0).
		Mul(decimal.New(blocktx.GasUsed, 0)).
		Shift(int32(0 - nas.WEI)).String()
	amount := blocktx.Amount.Shift(int32(0 - blocktx.Decimal)).String()
	pushBlockTx.Txs = append(pushBlockTx.Txs, bo.PushAccountTx{
		Txid:     blocktx.Txid,
		From:     blocktx.FromAddress,
		To:       blocktx.ToAddress,
		Contract: blocktx.ContractAddress,
		Fee:      fee,
		Amount:   amount,
	})

	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return err
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
		s.pusher.AddPushUserTask(int64(blockInfo.Height), pushdata)
	}

	return nil
}
