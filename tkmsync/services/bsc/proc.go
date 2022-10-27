package bsc

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"rsksync/common"
	"rsksync/conf"
	"rsksync/models/bo"
	dao "rsksync/models/po/bsc"
	"rsksync/services"
	"rsksync/utils"
	"rsksync/utils/bsc"
	"strings"
	"sync"
	"time"
)

type Processor struct {
	*bsc.RpcClient
	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		RpcClient: bsc.NewRpcClient(node.Url),
		watch:     watch,
		conf:      conf.Sync,
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
	log.Printf("RepushTx user: %d , txid : %s \n", userid, txid)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	bestBlockHeight, err := s.BlockNumber()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}
	//if blocktx, err := s.getBlockTxFromDB(txid); err == nil {
	//	return s.processTX([]*dao.BlockTX{blocktx}, bestBlockHeight)
	//} else {
	//	blocktxs, watchs, err := s.getBlockTxFromNode(txid)
	//	if err != nil {
	//		return fmt.Errorf("don't get block tx %v", err)
	//	}
	//	return s.processPush(watchs, bestBlockHeight, blocktxs...)
	//}

	blocktxs, watchs, err := s.getBlockTxFromNode(txid)
	if err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}
	//临时版本改变时间
	for i, _ := range blocktxs {
		//9天钱数据的话改变为现在时间
		oldTime := time.Now().AddDate(0, 0, -9)
		if blocktxs[i].Timestamp.Before(oldTime) {
			blocktxs[i].Timestamp = time.Now()
		}
	}
	return s.processPush(watchs, bestBlockHeight, blocktxs...)
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
		for _, tmp := range tmps {
			go func(w *sync.WaitGroup, tx []*dao.BlockTX) {
				w.Add(1)
				defer w.Done()
				if err := s.processTX(tx, bestHeight); err == nil {
				}
			}(wg, tmp.([]*dao.BlockTX))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.([]*dao.BlockTX)
			if err := s.processTX(tx, bestHeight); err == nil {
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
		for _, tmp := range tmps {
			go func(w *sync.WaitGroup, tx []*dao.BlockTX) {
				w.Add(1)
				defer w.Done()
				if err := s.processTX(tx, bestHeight); err == nil {
				}
			}(wg, tmp.([]*dao.BlockTX))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.([]*dao.BlockTX)
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
	if bs, err := dao.GetUnconfirmBlockInfos(s.conf.Confirmations); err == nil && bs != nil && len(bs) > 0 {
		var ids []int64
		//开始同步更新确认数
		for _, blk := range bs {
			blk.Confirmations++
			s.confirmsPush(blk)
			ids = append(ids, blk.Id)
		}
		//批量更新订单确认数。
		if err := dao.BatchUpdateConfirmations(ids, 1); err != nil {
			log.Printf("batch update confirmations err: %s \n", err.Error())
		}
	}
}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	s.confirmsPush(b.(*dao.BlockInfo))
}

func (s *Processor) getBlockTxFromDB(txid string) (*dao.BlockTX, error) {
	return dao.SelecBlockTxByHash(txid)
}

func (s *Processor) getBlockTxFromNode(txid string) ([]*dao.BlockTX, map[string]bool, error) {
	tx, err := s.GetTransactionByHash(txid)
	if err != nil || tx == nil {
		return nil, nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}
	block, err := s.GetBlockByNumber(tx.BlockNumber, false)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}
	txReceipt, err := s.GetTransactionReceipt(tx.Hash)
	if err != nil {
		return nil, nil, err
	}
	status, _ := utils.ParseInt(txReceipt.Status)
	if status != 1 {
		return nil, nil, fmt.Errorf("this contract tx status err : %d ", status)
	}
	//p,_:=json.Marshal(txReceipt)
	//log.Infof("txReceipt is : %s",string(p))
	//if txReceipt.Removed {
	//	return nil, nil, fmt.Errorf("this contract tx Removed err : %t ", txReceipt.Removed)
	//}

	blocktxs := make([]*dao.BlockTX, 0)
	watchLists := make(map[string]bool)
	if !s.IsContractTx(tx) {
		blocktx := &dao.BlockTX{
			BlockHeight: tx.BlockNumber,
			BlockHash:   tx.BlockHash,
			Txid:        tx.Hash,
			FromAddress: tx.From,
			Nonce:       tx.Nonce,
			GasPrice:    tx.GasPrice.Int64(),
			Input:       tx.Input,
			CoinName:    s.conf.Name,
			Decimal:     bsc.WEI,
			Timestamp:   time.Unix(block.Timestamp, 0),
			Amount:      decimal.NewFromBigInt(tx.Value, 0),
			ToAddress:   tx.To,
			GasUsed:     txReceipt.GasUsed,
			CreateTime:  time.Now(),
		}
		blocktxs = append(blocktxs, blocktx)
		if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
			watchLists[blocktx.FromAddress] = true
		}
		if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
			watchLists[blocktx.ToAddress] = true
		}
		log.Printf("dont't have care of watch address ,from: %s, to: %s \n", blocktx.FromAddress, blocktx.ToAddress)
	} else {
		log.Println("合约交易")
		//由于内层销毁的金额无法获取，因此需要外层-内层的金额
		isDetroyTx := false
		outAmount := decimal.Zero
		rawTx, err := s.GetTransactionByHash(tx.Hash)
		if err != nil {
			return nil, nil, err
		}
		//差异金额的标识
		if len(rawTx.Input) == 138 && len(txReceipt.Logs) == 1 {

			if strings.HasPrefix(rawTx.Input, "0xa9059cbb000000000000000000000000") {
				isDetroyTx = true
				//有可能是销毁币种
				am, _ := new(big.Int).SetString(rawTx.Input[74:], 16)
				outAmount = decimal.NewFromBigInt(am, 0)
				if outAmount.IsZero() {
					return nil, nil, fmt.Errorf("txid:[%s]解析原始金额错误,可能非交易类型", tx.Hash)
				}
			}
		}
		inAmount := decimal.Zero
		for _, lg := range txReceipt.Logs {
			blocktx := &dao.BlockTX{
				BlockHeight:     tx.BlockNumber,
				BlockHash:       tx.BlockHash,
				Txid:            tx.Hash,
				Nonce:           tx.Nonce,
				GasUsed:         txReceipt.GasUsed,
				GasPrice:        tx.GasPrice.Int64(),
				Input:           tx.Input,
				CoinName:        s.conf.Name,
				Decimal:         bsc.WEI,
				Timestamp:       time.Unix(block.Timestamp, 0),
				ContractAddress: lg.Address,
				CreateTime:      time.Now(),
			}

			contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
			if err != nil {
				log.Printf("dont't have care of watch contract : %s \n", blocktx.ContractAddress)
				continue
			}
			blocktx.CoinName = contractInfo.Name
			blocktx.Decimal = contractInfo.Decimal
			if lg.Data == "" || len(lg.Data) < 3 {
				continue
			}
			tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
			blocktx.Amount = decimal.NewFromBigInt(tmp, 0)
			inAmount = inAmount.Add(decimal.NewFromBigInt(tmp, 0))
			if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
				continue
			}
			if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
				blocktx.FromAddress = "0x" + lg.Topics[1][26:66]
				blocktx.ToAddress = "0x" + lg.Topics[2][26:66]
			} else {
				continue
			}

			if !s.watch.IsWatchAddressExist(blocktx.FromAddress) && !s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				log.Printf("dont't have care of watch address ,from: %s, to: %s \n", blocktx.FromAddress, blocktx.ToAddress)
				continue
			}
			if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
				watchLists[blocktx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				watchLists[blocktx.ToAddress] = true
			}
			log.Println("推送结果：", blocktx.ToAddress, "  amm:", blocktx.Amount)
			blocktxs = append(blocktxs, blocktx)
		}
		if isDetroyTx && !inAmount.IsZero() {

			if outAmount.GreaterThan(inAmount) {
				contractInfo, err := s.watch.GetContract(rawTx.To)
				if err != nil {
					log.Printf("dont't have care of watch contract : %s \n", rawTx.To)
					return nil, nil, err
				}

				blockTx := &dao.BlockTX{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					Nonce:           tx.Nonce,
					GasUsed:         txReceipt.GasUsed,
					GasPrice:        tx.GasPrice.Int64(),
					Input:           tx.Input,
					CoinName:        contractInfo.Name,
					Decimal:         contractInfo.Decimal, //
					Timestamp:       time.Unix(block.Timestamp, 0),
					ContractAddress: rawTx.To,
					CreateTime:      time.Now(),
					Status:          1,
					Amount:          outAmount.Sub(inAmount),
					FromAddress:     rawTx.From,
					ToAddress:       "0x0000000000000000000000000000000000000000",
				}
				if s.watch.IsWatchAddressExist(blockTx.FromAddress) {
					watchLists[blockTx.FromAddress] = true
				}
				if s.watch.IsWatchAddressExist(blockTx.ToAddress) {
					watchLists[blockTx.ToAddress] = true
				}
				blocktxs = append(blocktxs, blockTx)
				dd, _ := json.Marshal(blockTx)
				log.Printf("【%s】销毁币种，添加数据推送 【%s】", string(blockTx.Txid), string(dd))
			}
		}

	}
	if len(blocktxs) == 0 {
		return nil, nil, fmt.Errorf("dont't have care of tx")
	}
	return blocktxs, watchLists, nil
}

// 解析交易信息到db
func (s *Processor) processTX(blocktxs []*dao.BlockTX, bestHeight int64) error {

	if len(blocktxs) == 0 {
		return fmt.Errorf("tx is null")
	}
	txReceipt, err := s.GetTransactionReceipt(blocktxs[0].Txid)
	if err != nil {
		return err
	}
	status, _ := utils.ParseInt(txReceipt.Status)
	if s.conf.Name == "etc" {
		status = 1
	}
	//检测是否为关心的地址
	results := make([]*dao.BlockTX, 0)
	tmpWatchList := make(map[string]bool)
	for _, blocktx := range blocktxs {
		if blocktx.ContractAddress == "" {
			if !s.watch.IsWatchAddressExist(blocktx.FromAddress) && !s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				continue
			}
			blocktx.CreateTime = time.Now()
			blocktx.Decimal = bsc.WEI
			blocktx.GasUsed = txReceipt.GasUsed
			blocktx.Status = status
			if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
				tmpWatchList[blocktx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				tmpWatchList[blocktx.ToAddress] = true
			}
			if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
				log.Printf("block tx insert err: %v \n", err)
			}
			if blocktx.Status == 1 {
				results = append(results, blocktx)
			}
		} else {
			//合约交易
			contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
			if err != nil {
				continue
			}
			if !s.watch.IsWatchAddressExist(blocktx.FromAddress) && !s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				continue
			}
			blocktx.CoinName = contractInfo.Name
			blocktx.Decimal = contractInfo.Decimal
			blocktx.GasUsed = txReceipt.GasUsed
			blocktx.Status = status
			if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
				blocktx.Status = 2
			}
			if txReceipt.Logs != nil {
				btys, _ := json.Marshal(txReceipt.Logs)
				blocktx.Logs = string(btys)
			}
			if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
				tmpWatchList[blocktx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				tmpWatchList[blocktx.ToAddress] = true
			}
			if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
				log.Printf("block tx insert err: %v \n", err)
			}
			if blocktx.Status == 1 {
				results = append(results, blocktx)
			}
		}
	}
	if len(tmpWatchList) == 0 || len(results) == 0 {
		return fmt.Errorf("dont't have care of watch address ")
	}
	if status == 1 {
		s.processPush(tmpWatchList, bestHeight, results...)
	} else {
		log.Printf("block tx status : %d is failed \n", status)
	}
	return nil
}

func (s *Processor) processPush(tmpWatchList map[string]bool, bestHeight int64, blocktxs ...*dao.BlockTX) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktxs[0].BlockHeight,
		Hash:          blocktxs[0].BlockHash,
		CoinName:      s.conf.Name,
		Token:         "",
		Confirmations: bestHeight - blocktxs[0].BlockHeight + 1,
		Time:          blocktxs[0].Timestamp.Unix(),
	}
	if blocktxs[0].CoinName != s.conf.Name {
		pushBlockTx.Token = blocktxs[0].CoinName
	}
	fee := decimal.New(blocktxs[0].GasPrice, 0).
		Mul(decimal.New(blocktxs[0].GasUsed, 0)).
		Shift(int32(0 - bsc.WEI)).String()
	for _, blocktx := range blocktxs {
		amount := blocktx.Amount.Shift(int32(0 - blocktx.Decimal)).String()
		pushBlockTx.Txs = append(pushBlockTx.Txs, bo.PushAccountTx{
			Txid:     blocktx.Txid,
			From:     blocktx.FromAddress,
			To:       blocktx.ToAddress,
			Contract: blocktx.ContractAddress,
			Fee:      fee,
			Amount:   amount,
		})
	}
	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return fmt.Errorf("marshal err %v", err)
	}
	if s.pusher != nil {
		log.Printf("推送结构：%s \n", string(pusdata))
		s.pusher.AddPushTask(pushBlockTx.Height, blocktxs[0].Txid, tmpWatchList, pusdata)
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
