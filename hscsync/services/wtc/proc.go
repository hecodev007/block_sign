package wtc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"hscsync/common"
	"hscsync/common/conf"
	"hscsync/common/log"
	"hscsync/models/bo"
	dao "hscsync/models/po/yotta"
	"hscsync/services"
	"hscsync/utils"
	"hscsync/utils/dingding"
	"hscsync/utils/eth"
	rpc "hscsync/utils/eth"
	"fmt"
	"github.com/onethefour/common/xutils"
	"math/big"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Processor struct {
	*rpc.RpcClient
	Watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		RpcClient: rpc.NewRpcClient(node.Url),
		Watch:     watch,
		conf:      conf.Sync,
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

//func (s *Processor) processTX(txInfo *TxInfo) (map[string]bool, []*dao.BlockTxVout, []*dao.BlockTxVout, error) {
//
//	if txInfo == nil {
//		return nil, nil, nil, fmt.Errorf("tx info don't allow nil")
//	}
//
//	var updateVins, insertVins []*dao.BlockTxVout
//	tmpWatchList := make(map[string]bool)
//
//	//starttime := time.Now()
//	amtout := decimal.Zero
//	for _, txvout := range txInfo.vouts {
//		amtout = amtout.Add(txvout.Value)
//
//		if s.Watch.IsWatchAddressExist(txvout.Address) {
//			tmpWatchList[txvout.Address] = true
//		}
//	}
//
//	amtin := decimal.Zero
//	for _, txvin := range txInfo.vins {
//
//		if txvin.Txid == "coinbase" {
//			continue
//		}
//
//		if vout, err := dao.SelectBlockTXVout(txvin.Txid, txvin.Voutn); err == nil {
//			txvin.Id = vout.Id
//			txvin.Value = vout.Value
//			txvin.Address = vout.Address
//
//			if s.conf.FullBackup {
//				updateVins = append(updateVins, txvin)
//			} else {
//				if s.Watch.IsWatchAddressExist(txvin.Address) {
//					updateVins = append(updateVins, txvin)
//				}
//			}
//		} else {
//			if err != gorm.ErrRecordNotFound {
//				log.Printf("processTX SelectBlockTXVout txid: %s, n: %d, err: %v", txvin.Txid, txvin.Voutn, err)
//				continue
//			}
//			tx, err := s.GetRawTransaction(txvin.Txid)
//			if err != nil {
//				log.Println(err.Error())
//				return nil, nil, nil, fmt.Errorf("processTX : GetRawTransaction %s", txvin.Txid)
//			}
//			txvin.BlockHash = tx.BlockHash
//			txvin.Value = decimal.NewFromFloat(tx.Vout[txvin.Voutn].Value)
//			txvin.Timestamp = time.Unix(tx.Time, 0)
//			txvin.CreateTime = time.Now()
//			address, err := tx.Vout[txvin.Voutn].ScriptPubkey.GetAddress()
//			if err == nil {
//				txvin.Address = address[0]
//			}
//			data, _ := json.Marshal(tx.Vout[txvin.Voutn].ScriptPubkey)
//			txvin.ScriptPubKey = string(data)
//
//			if s.conf.FullBackup {
//				insertVins = append(insertVins, txvin)
//			} else {
//				if s.Watch.IsWatchAddressExist(txvin.Address) {
//					insertVins = append(insertVins, txvin)
//				}
//			}
//		}
//
//		amtin = amtin.Add(txvin.Value)
//		if s.Watch.IsWatchAddressExist(txvin.Address) {
//			tmpWatchList[txvin.Address] = true
//		}
//	}
//
//	txInfo.tx.Fee = amtin.Sub(amtout)
//	if txInfo.tx.Fee.IsNegative() && s.conf.CoinName != "doge" {
//		return nil, nil, nil, fmt.Errorf("tx fee don't allow negative,vin:%v, vout:%v", amtin, amtout)
//	}
//
//	//log.Printf("processTX : %s ,vins : %d , vouts : %d , used time : %f 's", blocktx.Txid, len(txvins), len(insertVouts), time.Since(starttime).Seconds())
//	return tmpWatchList, updateVins, insertVins, nil
//}
func (s *Processor) getBlockTxFromNode(txid string) ([]*dao.BlockTx, error) {
	tx, err := s.GetTransactionByHash(txid)
	if err != nil || tx == nil {
		return nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}
	block, err := s.GetBlockByNumber(tx.BlockNumber, false)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}
	bestHeight, err := s.BlockNumber()
	if err != nil {
		return nil, fmt.Errorf("BlockNumber err: %v ", err)
	}
	daoBlock := &dao.BlockInfo{
		Height:            block.Number,
		Hash:              block.Hash,
		Previousblockhash: block.ParentHash,
		Timestamp:         time.Unix(int64(block.Timestamp), 0),
		Transactions:      len(block.Transactions),
		Confirmations:     bestHeight - block.Number + 1,
		Createtime:        time.Now(),
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
func (s *Processor) parseBlockTX(tx *rpc.Transaction, block *dao.BlockInfo) (blocktxs []*dao.BlockTx, err error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	//Watch := s.Watch
	//rpc := s.RpcClient
	//var blocktxs = make([]*dao.BlockTx, 0)
reGetTransactionReceipt:
	txReceipt, err := s.RpcClient.GetTransactionReceipt(tx.Hash)
	if err != nil {
		log.Info(err.Error())
		time.Sleep(time.Second * 2)
		goto reGetTransactionReceipt
	}

	if txReceipt == nil {
		blocktx := dao.BlockTx{
			CoinName:        conf.Cfg.Sync.CoinName,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: "",
			FromAddress:     tx.From,
			ToAddress:       tx.To,
			Amount:          decimal.NewFromBigInt(tx.Value, -18),
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(21000)), -18),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return
	}
	//eth 合约转账交易
	if tx.Value.Cmp(big.NewInt(0)) > 0 && (s.Watch.IsWatchAddressExist(tx.From) || s.Watch.IsWatchAddressExist(tx.To)) {

		blocktx := dao.BlockTx{
			CoinName:        conf.Cfg.Sync.CoinName,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: "",
			FromAddress:     tx.From,
			ToAddress:       tx.To,
			Amount:          decimal.NewFromBigInt(tx.Value, -18),
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return
	}
	//失败的提现交易把手续费推送过去
	if txReceipt.Status != "0x1" && s.Watch.IsWatchAddressExist(tx.From) {
		to :=tx.To
		contractAddress:=""
		if s.Watch.IsContractExist(tx.To){
			tokenTo, _, err := rpc.ERC20{}.ParseTransferData(tx.Input)
			if err != nil {
				to=tokenTo
				contractAddress=tx.To
			}
		}

		blocktx := dao.BlockTx{
			CoinName:        conf.Cfg.Sync.CoinName,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: contractAddress,
			FromAddress:     tx.From,
			ToAddress:       to,
			Amount:          decimal.Zero,
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return
	}
	//其他失败的交易不处理
	if txReceipt.Status != "0x1" {
		return nil, errors.New("失败的交易")
	}

	if len(txReceipt.Logs) == 0 {
		return nil, errors.New("交易不含监控地址")
	}
	if s.Watch.IsContractExist(tx.To) {
		_, _, erc20err := rpc.ERC20{}.ParseTransferData(tx.Input)
		if erc20err == nil { //erc20 包括erc20销毁币
			for _, txlog := range txReceipt.Logs {
				contractInfo, err := s.Watch.GetContract(txlog.Address)
				if err != nil {
					//log.Info(err.Error())
					continue
				}
				if len(txlog.Topics) != 3 || txlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
					continue
				}
				from := "0x" + txlog.Topics[1][26:66]
				to := "0x" + txlog.Topics[2][26:66]
				data, _ := hex.DecodeString(strings.TrimPrefix(txlog.Data, "0x"))
				amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(data), 0-int32(contractInfo.Decimal))
				blocktx := &dao.BlockTx{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					FromAddress:     from,
					CoinName:        conf.Cfg.Sync.CoinName,
					Timestamp:       block.Timestamp,
					Amount:          amount,
					ToAddress:       to,
					Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
					Status:          "success",
					ContractAddress: txlog.Address,
				}
				blocktxs = append(blocktxs, blocktx)
			}
			return
		}

		//if tx.To == "0x1c2349acbb7f83d07577692c75b6d7654899bf10" {
		//	for _, lg := range txReceipt.Logs {
		//		if lg.Address != "0x1c2349acbb7f83d07577692c75b6d7654899bf10" {
		//			continue
		//		}
		//
		//		if len(lg.Topics) == 0 || lg.Topics[0] != "0x3efc190d59645f005a5974aa84aa94401ad997938870e7b2aa74a45138ad679b" {
		//			continue
		//		}
		//		log.Info(lg.Topics[0])
		//		blocktx := &dao.BlockTx{
		//			BlockHeight:     tx.BlockNumber,
		//			BlockHash:       tx.BlockHash,
		//			Txid:            tx.Hash,
		//			CoinName:        conf.Cfg.Sync.CoinName,
		//			Timestamp:       block.Timestamp,
		//			ContractAddress: lg.Address,
		//		}
		//
		//		haTopicsHash := make([]ethcommon.Hash, 0)
		//		for _, vt := range lg.Topics {
		//			haTopicsHash = append(haTopicsHash, ethcommon.HexToHash(vt))
		//		}
		//		vLog := types.Log{
		//			Address:     ethcommon.HexToAddress(lg.Address),
		//			Topics:      haTopicsHash,
		//			Data:        ethcommon.FromHex(lg.Data),
		//			BlockNumber: uint64(tx.BlockNumber),
		//			TxHash:      ethcommon.HexToHash(txReceipt.TransactionHash),
		//			TxIndex:     uint(txReceipt.TransactionIndex),
		//			BlockHash:   ethcommon.HexToHash(txReceipt.BlockHash),
		//			Index:       uint(txReceipt.LogIndex),
		//			Removed:     txReceipt.Removed,
		//		}
		//		sender, receiver, amountFloatStr, txid, err := rpc.MyKeyProcessTransferLogic(vLog)
		//		log.Infof("mykey 解析数据解析：sender：%s,receiver:%s,am:%s", sender, receiver, amountFloatStr)
		//		if err != nil {
		//			log.Infof("mykey 解析异常 err:%s", err.Error())
		//			continue
		//		}
		//		if txid != tx.Hash {
		//			log.Infof("mykey txid 不一致")
		//			continue
		//		}
		//
		//		am, _ := decimal.NewFromString(amountFloatStr)
		//		if sender == "" || receiver == "" || am.IsZero() {
		//			log.Infof("mykey txid 解析数据解析不全：sender：%s,receiver:%s,am:%s", sender, receiver, am.String())
		//			continue
		//		}
		//		blocktx.Amount = am //扩大18变成int
		//		blocktx.FromAddress = sender
		//		blocktx.ToAddress = receiver
		//		blocktx.ContractAddress = "" //合约清空
		//		blocktxs = append(blocktxs, blocktx)
		//		log.Infof("添加mykey eth交易，txid：%s", txid)
		//	}
		//	return
		//}
		//if erc20err != nil {
		//	return nil, errors.New("不支持的转账交易")
		//}
	}
	if conf.Cfg.Sync.EnableInternal {
		internoTx, err := s.InternalTxns(tx, block.Timestamp.Unix())
		if err == nil && internoTx != nil {
			blocktxs = append(blocktxs, internoTx)
			return blocktxs,nil
		}
	}
	//log.Info(err,s.Watch.IsContractExist(tx.To))
	if tx.BlockNumber> conf.Cfg.Sync.InitHeight { //防止以前已经退币的交易
		for _, txlog := range txReceipt.Logs {
			contractInfo, err := s.Watch.GetContract(txlog.Address)
			if err != nil {
				log.Info(err.Error())
				continue
			}
			if len(txlog.Topics) == 0 || txlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
				continue
			}
			//log.Info("a")
			from := "0x" + txlog.Topics[1][26:66]
			to := "0x" + txlog.Topics[2][26:66]
			data, _ := hex.DecodeString(strings.TrimPrefix(txlog.Data, "0x"))
			value := decimal.NewFromBigInt(big.NewInt(0).SetBytes(data), 0-int32(contractInfo.Decimal))
			if !s.Watch.IsWatchAddressExist(to){
				continue
			}

			blocktx := &dao.BlockTx{
				BlockHeight:     tx.BlockNumber,
				BlockHash:       tx.BlockHash,
				Txid:            tx.Hash,
				FromAddress:     from,
				CoinName:        conf.Cfg.Sync.CoinName,
				Timestamp:       block.Timestamp,
				Amount:          value,
				ToAddress:       to,
				Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
				Status:          "success",
				ContractAddress: txlog.Address,
			}
			blocktxs = append(blocktxs, blocktx)
			err = nil
		}
	} else {
		return nil,fmt.Errorf("高度低于新版本数据服务开始高度:%v,需要开发调低initheight配置!",conf.Cfg.Sync.InitHeight)
	}

	return
}

func (s *Processor) InternalTxns(tx *eth.Transaction, blockTimestamp int64) (*dao.BlockTx, error) {
	debugTraceTransactionInfo, debugTraceTransactionInfoErr := s.GetTraceTransaction(tx.Hash)
	if debugTraceTransactionInfoErr != nil {
		return nil, fmt.Errorf("%s未获取到内部交易,err:%s", tx.Hash, debugTraceTransactionInfoErr.Error())
	}
	if debugTraceTransactionInfo.Calls != nil {
		return s.InternalTxnsRecursion(debugTraceTransactionInfo.Calls, tx, blockTimestamp)
	}
	return nil, fmt.Errorf("%s未获取到内部Calls交易数据,calls为0", tx.Hash)
}

func (s *Processor) InternalTxnsRecursion(callsArr []eth.TraceTransactionInfoCalls, tx *eth.Transaction, blockTimestamp int64) (*dao.BlockTx, error) {
	for _, v := range callsArr {
		Amount, _ := utils.ParseBigInt(v.Value)
		if Amount.Sign() == 1 && v.Type == "CALL" && v.Input == "0x" && v.Output == "0x" {
			watchLists := make(map[string]bool)
			if s.Watch.IsWatchAddressExist(v.From) {
				watchLists[v.From] = true
			}
			if s.Watch.IsWatchAddressExist(v.To) {
				watchLists[v.To] = true
			}
			if len(watchLists) > 0 {
				//GasUsed, _ := strconv.ParseInt(v.GasUsed, 0, 64)
				blocktx := &dao.BlockTx{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					CoinName:        conf.Cfg.Sync.CoinName,
					Timestamp:       time.Unix(blockTimestamp, 0),
					ContractAddress: "", //这个没填
				}
				blocktx.FromAddress = v.From
				blocktx.ToAddress = v.To

				blocktx.Amount = decimal.NewFromBigInt(Amount, -18)
				dingding.NotifyError(fmt.Sprintf("内部交易监控到txId: %s\nform：%s\namount:%s\nto:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(eth.WEI).String(), blocktx.ToAddress))
				log.Infof("内部交易监控到地址%s form：%s,amount:%s,to:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(-eth.WEI).String(), blocktx.ToAddress)
				return blocktx, nil
			}
		}
		Internalblocktx, InternalErr := s.InternalTxnsRecursion(v.Calls, tx, blockTimestamp)
		if InternalErr == nil {
			return Internalblocktx, InternalErr
		}
	}
	return nil, fmt.Errorf("%s未获取到内部交易", tx.Hash)
}
