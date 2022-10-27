package chain

import (
	"dataserver/common"
	"dataserver/conf"
	"dataserver/log"
	"dataserver/models/bo"
	"dataserver/models/po"
	"dataserver/services"
	"dataserver/utils"
	"dataserver/utils/dingding"
	"dataserver/utils/eth"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Processor struct {
	*eth.RpcClient
	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		RpcClient: eth.NewRpcClient(node.Url),
		watch:     watch,
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

func (s *Processor) RepushTx(userid int64, txid string) error {
	log.Infof("RepushTx user: %d , txid : %s", userid, txid)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	bestBlockHeight, err := s.BlockNumber()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}
	blocktxs, _, err := s.getBlockTxFromNode(txid)
	if err != nil {
		if "没有发现监听的地址" == err.Error() {
			log.Infof("txId=%s 重推没有找到关心的地址，需要重新reload数据")

			if err = s.watch.Reload(); err != nil {
				log.Errorf("watch Reload error %s", err.Error())
				return fmt.Errorf("(reload error)don't get block tx %v", err)
			}
			log.Infof("txId=%s watcher reload数据完成")
			blocktxs, _, err = s.getBlockTxFromNode(txid)
		}
		if err != nil {
			log.Infof("txId=%s don't get block tx")
			return fmt.Errorf("don't get block tx %v", err)
		}
	}

	return s.processTX(blocktxs, bestBlockHeight)
	//return s.processPush(watchs, bestBlockHeight, blocktxs...)
}

func (s *Processor) Info() (string, int64, error) {
	dbheight, err := po.GetMaxBlockIndex()
	return s.conf.Name, dbheight, err
}

func (s *Processor) Init() error {
	return nil
}

func (s *Processor) Clear() {

}

func (s *Processor) CheckIrreverseBlock(hash string) error {
	cnt, err := po.GetBlockCountByHash(hash)
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
			wg.Add(1)
			go func(w *sync.WaitGroup, tx []*po.BlockTX) {
				defer w.Done()
				if err := s.processTX(tx, bestHeight); err == nil {
				}
			}(wg, tmp.([]*po.BlockTX))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.([]*po.BlockTX)
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
			wg.Add(1)
			go func(w *sync.WaitGroup, tx []*po.BlockTX) {
				defer w.Done()
				if err := s.processTX(tx, bestHeight); err == nil {
				}
			}(wg, tmp.([]*po.BlockTX))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.([]*po.BlockTX)
			if err := s.processTX(tx, bestHeight); err == nil {
			}
		}
	}
	return nil
}

func (s *Processor) ProcIrreverseBlock(b interface{}) error {
	block := b.(*po.BlockInfo)
	if _, err := po.InsertBlockInfo(block); err != nil {
		return fmt.Errorf("block %d Insert Block err : %v", block.Height, err)
	}
	return nil
}

func (s *Processor) UpdateIrreverseConfirms() {
	// 查找所有未确认的区块
	if bs, err := po.GetUnconfirmBlockInfos(s.conf.Confirmations + 6); err == nil && bs != nil && len(bs) > 0 {
		var ids []int64
		// 开始同步更新确认数
		for _, blk := range bs {
			blk.Confirmations++
			s.confirmsPush(blk)
			ids = append(ids, blk.Id)
		}
		// 批量更新订单确认数。
		if err := po.BatchUpdateConfirmations(ids, 1); err != nil {
			log.Errorf("batch update confirmations err: %v", err)
		}
	}
}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	s.confirmsPush(b.(*po.BlockInfo))
}

func (s *Processor) getBlockTxFromDB(txid string) (*po.BlockTX, error) {
	return po.SelecBlockTxByHash(txid)
}

func (s *Processor) getBlockTxFromNode(txid string) ([]*po.BlockTX, map[string]bool, error) {
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

	blocktxs := make([]*po.BlockTX, 0)
	watchLists := make(map[string]bool)
	if !s.IsContractTx(tx) {
		blocktx := &po.BlockTX{
			BlockHeight: tx.BlockNumber,
			BlockHash:   tx.BlockHash,
			Txid:        tx.Hash,
			FromAddress: tx.From,
			Nonce:       tx.Nonce,
			GasPrice:    tx.GasPrice.Int64(),
			Input:       tx.Input,
			CoinName:    s.conf.Name,
			Decimal:     eth.WEI,
			Timestamp:   time.Unix(block.Timestamp, 0),
			//Timestamp:  time.Now(),
			Amount:     decimal.NewFromBigInt(tx.Value, 0),
			ToAddress:  tx.To,
			GasUsed:    txReceipt.GasUsed,
			CreateTime: time.Now(),
			Status:     status,
		}
		blocktxs = append(blocktxs, blocktx)
		if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
			watchLists[blocktx.FromAddress] = true
		}
		if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
			watchLists[blocktx.ToAddress] = true
		}
		log.Errorf("dont't have care of watch address ,from: %s, to: %s", blocktx.FromAddress, blocktx.ToAddress)
	} else {
		if errParseTx := s.parseTxReceipt(txReceipt, tx, block.Timestamp, status, watchLists, &blocktxs); errParseTx != nil {
			blocktx, InternalTxnsWatchLists, InternalTxnsErr := s.InternalTxns(tx, block.Timestamp) //添加内部交易判断
			if InternalTxnsErr == nil {                                                             //有内部数据
				blocktxs = append(blocktxs, blocktx)
				return blocktxs, InternalTxnsWatchLists, nil
			}
			return nil, nil, errParseTx
		}
	}
	if len(blocktxs) == 0 {
		blocktx, InternalTxnsWatchLists, InternalTxnsErr := s.InternalTxns(tx, block.Timestamp) //添加内部交易判断
		if InternalTxnsErr == nil {                                                             //有内部数据
			blocktxs = append(blocktxs, blocktx)
			return blocktxs, InternalTxnsWatchLists, nil
		}
		return nil, nil, errors.New("没有发现监听的地址")
	}
	return blocktxs, watchLists, nil
}

func (s *Processor) parseTxReceipt(txReceipt *eth.TransactionReceipt, tx *eth.Transaction, blockTimestamp int64, status int, watchLists map[string]bool, blocktxs *[]*po.BlockTX) error {
	log.Info("合约交易")
	inAmount := decimal.Zero
	for _, lg := range txReceipt.Logs {
		blocktx := &po.BlockTX{
			BlockHeight:     tx.BlockNumber,
			BlockHash:       tx.BlockHash,
			Txid:            tx.Hash,
			Nonce:           tx.Nonce,
			GasUsed:         txReceipt.GasUsed,
			GasPrice:        tx.GasPrice.Int64(),
			Input:           tx.Input,
			CoinName:        s.conf.Name,
			Decimal:         eth.WEI,
			Timestamp:       time.Now(),
			ContractAddress: lg.Address,
			CreateTime:      time.Now(),
			Status:          status,
		}

		contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
		if err != nil {
			log.Errorf("dont't have care of watch contract : %s", blocktx.ContractAddress)
			continue
		}
		blocktx.CoinName = contractInfo.Name
		blocktx.Decimal = contractInfo.Decimal
		// 没有输出日志数据，认为是非合法的交易
		// status=2 表示失败
		if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
			blocktx.Status = 2
		}
		if lg.Data == "" || len(lg.Data) < 3 {
			continue
		}
		tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
		blocktx.Amount = decimal.NewFromBigInt(tmp, 0)

		// 内部金额叠加
		inAmount = inAmount.Add(decimal.NewFromBigInt(tmp, 0))

		if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
			continue
		}
		if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			blocktx.FromAddress = "0x" + lg.Topics[1][26:66]
			blocktx.ToAddress = "0x" + lg.Topics[2][26:66]

			if lg.Address == GDR_Contract && blocktx.ToAddress == GDR_Owner && lg.Data == ZERO {
				log.Infof("%s 是GDR币种的特殊交易，不需要推送", tx.Hash)
				continue
			}

		} else {
			continue
		}

		if !s.watch.IsWatchAddressExist(blocktx.FromAddress) && !s.watch.IsWatchAddressExist(blocktx.ToAddress) {
			log.Errorf("dont't have care of watch address ,from: %s, to: %s", blocktx.FromAddress, blocktx.ToAddress)
			continue
		}
		if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
			watchLists[blocktx.FromAddress] = true
		}
		if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
			watchLists[blocktx.ToAddress] = true
		}
		if blocktx.Status == 1 {
			log.Info("推送结果：", blocktx.ToAddress, "  amm:", blocktx.Amount)
			*blocktxs = append(*blocktxs, blocktx)
		}
	}

	if len(*blocktxs) > 0 {
		// log.Println(fmt.Sprintf("【%s】，存在交易，开始解析数据是否销毁", tx.Hash, tx.From, tx.To))
		// 由于内层销毁的金额无法获取，因此需要外层-内层的金额
		outAmount := decimal.Zero
		rawTx := tx
		if len(rawTx.Input) == 0 || len(txReceipt.Logs) == 0 {
			// log.Printf("%s 没有可以解析的交易", txReceipt.TransactionHash)
			return fmt.Errorf("%s 没有可以解析的交易", txReceipt.TransactionHash)
		}
		// 差异金额的标识
		if len(rawTx.Input) == 138 && len(txReceipt.Logs) == 1 && !inAmount.IsZero() {
			if strings.HasPrefix(rawTx.Input, "0xa9059cbb000000000000000000000000") {
				// 有可能是销毁币种
				am, _ := new(big.Int).SetString(rawTx.Input[74:], 16)
				outAmount = decimal.NewFromBigInt(am, 0)
				if outAmount.IsZero() {
					return fmt.Errorf("txid:[%s]解析原始金额错误,可能非交易类型", tx.Hash)
				}
				if outAmount.GreaterThan(inAmount) {
					detroyTx := &po.BlockTX{
						BlockHeight:     tx.BlockNumber,
						BlockHash:       tx.BlockHash,
						Txid:            tx.Hash,
						Nonce:           tx.Nonce,
						GasUsed:         txReceipt.GasUsed,
						GasPrice:        tx.GasPrice.Int64(),
						Input:           tx.Input,
						CoinName:        s.conf.Name,
						Decimal:         eth.WEI, //
						Timestamp:       time.Unix(blockTimestamp, 0),
						ContractAddress: rawTx.To,
						CreateTime:      time.Now(),
						Status:          1,
						Amount:          outAmount.Sub(inAmount),
						FromAddress:     rawTx.From,
						ToAddress:       "0x0000000000000000000000000000000000000000",
					}
					if detroyTx.ContractAddress == "0x18ff245c134d9daa6fed977617654490ba4da526" {
						log.Infof("【%s】销毁币种 MASKDOGE，跳过推送 ", detroyTx.Txid)
					} else {
						*blocktxs = append(*blocktxs, detroyTx)
						dd, _ := json.Marshal(detroyTx)
						log.Infof("【%s】销毁币种，添加数据 【%s】", detroyTx.Txid, string(dd))
					}
				}
			}
		}
	}
	return nil
}
func (s *Processor) InternalTxns(tx *eth.Transaction, blockTimestamp int64) (*po.BlockTX, map[string]bool, error) {
	debugTraceTransactionInfo, debugTraceTransactionInfoErr := s.GetTraceTransaction(tx.Hash)
	if debugTraceTransactionInfoErr != nil {
		return nil, map[string]bool{}, fmt.Errorf("%s未获取到内部交易数据,err:%s", tx.Hash, debugTraceTransactionInfoErr.Error())
	}
	if debugTraceTransactionInfo.Calls != nil {
		return s.InternalTxnsRecursion(debugTraceTransactionInfo.Calls, tx, blockTimestamp)
	}
	return nil, map[string]bool{}, fmt.Errorf("%s未获取到内部Calls交易数据,err:%s", tx.Hash, debugTraceTransactionInfoErr.Error())
}

func (s *Processor) InternalTxnsRecursion(callsArr []eth.TraceTransactionInfoCalls, tx *eth.Transaction, blockTimestamp int64) (*po.BlockTX, map[string]bool, error) {
	for _, v := range callsArr {
		Amount, _ := utils.ParseBigInt(v.Value)
		if Amount.Sign() == 1 && v.Type == "CALL" && v.Input == "0x" && v.Output == "0x" {
			watchLists := make(map[string]bool)
			if s.watch.IsWatchAddressExist(v.From) {
				watchLists[v.From] = true
			}
			if s.watch.IsWatchAddressExist(v.To) {
				watchLists[v.To] = true
			}
			if len(watchLists) > 0 {
				GasUsed, _ := strconv.ParseInt(v.GasUsed, 0, 64)
				blocktx := &po.BlockTX{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					Nonce:           tx.Nonce,
					GasUsed:         GasUsed,
					GasPrice:        tx.GasPrice.Int64(),
					Input:           tx.Input,
					CoinName:        s.conf.Name,
					Decimal:         eth.WEI,
					Timestamp:       time.Unix(blockTimestamp, 0),
					ContractAddress: "", //这个没填
					Status:          1,
				}
				blocktx.FromAddress = v.From
				blocktx.ToAddress = v.To
				blocktx.Amount = decimal.NewFromBigInt(Amount, 0)
				dingding.NotifyError(fmt.Sprintf("内部交易监控到地址%s, form：%s,amount:%s,to:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(-eth.WEI).String(), blocktx.ToAddress))
				log.Infof("内部交易监控到地址%s, form：%s,amount:%s,to:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(-eth.WEI).String(), blocktx.ToAddress)
				return blocktx, watchLists, nil
			}
		}
		InternalTensBlock, InternalWatchLists, InternalErr := s.InternalTxnsRecursion(v.Calls, tx, blockTimestamp)
		if InternalErr == nil {
			return InternalTensBlock, InternalWatchLists, InternalErr
		}
	}
	return nil, map[string]bool{}, fmt.Errorf("%s未获取到内部交易", tx.Hash)
}

// 解析交易信息到db
func (s *Processor) processTX(blocktxs []*po.BlockTX, bestHeight int64) error {

	if len(blocktxs) == 0 {
		return fmt.Errorf("tx is null")
	}
	// txReceipt, err := s.GetTransactionReceipt(blocktxs[0].Txid)
	// if err != nil {
	// 	return err
	// }
	// status, _ := utils.ParseInt(txReceipt.Status)

	// 检测是否为关心的地址
	results := make([]*po.BlockTX, 0)
	tmpWatchList := make(map[string]bool)
	for _, blockTx := range blocktxs {
		if blockTx.ContractAddress == "" {
			if !s.watch.IsWatchAddressExist(blockTx.FromAddress) && !s.watch.IsWatchAddressExist(blockTx.ToAddress) {
				log.Debugf("主链币交易txId %s 不是我们关心的出入账地址 from: %s to: %s", blockTx.Txid, blockTx.FromAddress, blockTx.ToAddress)
				continue
			}
			blockTx.CreateTime = time.Now()
			blockTx.Decimal = eth.WEI
			if s.watch.IsWatchAddressExist(blockTx.FromAddress) {
				tmpWatchList[blockTx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blockTx.ToAddress) {
				tmpWatchList[blockTx.ToAddress] = true
			}
			if num, err := po.InsertBlockTX(blockTx); num <= 0 || err != nil {
				log.Errorf("block tx insert err: %v", err)
			}
			if blockTx.Status == 1 {
				results = append(results, blockTx)
			}
		} else {
			// 合约交易
			contractInfo, err := s.watch.GetContract(blockTx.ContractAddress)
			if err != nil {
				continue
			}
			if !s.watch.IsWatchAddressExist(blockTx.FromAddress) && !s.watch.IsWatchAddressExist(blockTx.ToAddress) {
				log.Debugf("合约交易txId %s 不是我们关心的出入账地址 from: %s to: %s", blockTx.Txid, blockTx.FromAddress, blockTx.ToAddress)
				continue
			}
			blockTx.CoinName = contractInfo.Name
			blockTx.Decimal = contractInfo.Decimal
			// if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
			// 	blockTx.Status = 2
			// }
			// if txReceipt.Logs != nil {
			// 	btys, _ := json.Marshal(txReceipt.Logs)
			// 	blockTx.Logs = string(btys)
			// }
			if s.watch.IsWatchAddressExist(blockTx.FromAddress) {
				tmpWatchList[blockTx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blockTx.ToAddress) {
				tmpWatchList[blockTx.ToAddress] = true
			}
			if num, err := po.InsertBlockTX(blockTx); num <= 0 || err != nil {
				log.Errorf("block tx insert err: %v", err)
			}
			if blockTx.Status == 1 {
				results = append(results, blockTx)
			}
		}
	}
	if len(tmpWatchList) == 0 || len(results) == 0 {
		return fmt.Errorf("dont't have care of watch address ")
	}
	go s.processPush(tmpWatchList, bestHeight, results...)
	return nil
}

func (s *Processor) processPush(tmpWatchList map[string]bool, bestHeight int64, blocktxs ...*po.BlockTX) error {
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
		Shift(int32(0 - eth.WEI)).String()
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
		log.Infof("推送结构：%s", string(pusdata))
		s.pusher.AddPushTask(pushBlockTx.Height, blocktxs[0].Txid, tmpWatchList, pusdata)
	}
	return nil
}

func (s *Processor) confirmsPush(blockInfo *po.BlockInfo) error {
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
