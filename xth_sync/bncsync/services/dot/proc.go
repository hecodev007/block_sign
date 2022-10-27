package dot

import (
	"bncsync/common"
	"bncsync/common/conf"
	"bncsync/common/log"
	"bncsync/models/bo"
	dao "bncsync/models/po/dot"
	"bncsync/services"
	rpc "bncsync/utils/dot"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Processor struct {
	*rpc.RpcClient
	watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	rpcClient,err := rpc.NewRpcClient(node.Node, node.ScanApi,node.ScanKey)
	if err != nil {
		panic(err.Error())
	}
	return &Processor{
		RpcClient: rpcClient,
		watch:     watch,
		conf:      conf.Sync,
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
	//log.Info(string(tj))
	watchaddrses, txinfosmap := p.parseContractTxs(task.TxInfos)
	for txhash, watchaddrs := range watchaddrses {
		p.processPushs(txinfosmap[txhash], watchaddrs, task.BestHeight)
	}
	for _, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)
		if p.conf.FullBackup {
			dao.InsertTx(txInfo)
		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo)
		}
	}
	dao.InsertBlock(task.Block)
	return nil
}

func (s *Processor) processPush(blocktx *dao.BlockTx, tmpWatchList map[string]bool, bestHeight int64) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktx.Height,
		Hash:          blocktx.Hash,
		CoinName:      s.conf.Name,
		Token:         "", //blocktx.CoinName
		Confirmations: bestHeight - blocktx.Height + 1,
		Time:          time.Now().Unix(),
	}

	pushBlockTx.Txs = append(pushBlockTx.Txs, &bo.PushAccountTx{
		Name:     s.conf.Name,
		Txid:     blocktx.Txid,
		From:     blocktx.Fromaccount,
		To:       blocktx.Toaccount,
		Contract: blocktx.Contractaddress,
		Fee:      blocktx.SysFee,
		Amount:   blocktx.Amount,
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
	log.Info(string(pusdata))
	return nil
}

func (s *Processor) processPushs(blocktxs []*dao.BlockTx, tmpWatchList map[string]bool, bestHeight int64) error {
	if len(blocktxs) < 1 {
		return nil
	}
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktxs[0].Height,
		Hash:          blocktxs[0].Hash,
		CoinName:      s.conf.Name,
		Token:         "", //blocktx.CoinName
		Confirmations: bestHeight - blocktxs[0].Height + 1,
		Time:          time.Now().Unix(),
	}
	for _, blocktx := range blocktxs {
		pushBlockTx.Txs = append(pushBlockTx.Txs, &bo.PushAccountTx{
			Name:     s.conf.Name,
			Txid:     blocktx.Txid,
			From:     blocktx.Fromaccount,
			To:       blocktx.Toaccount,
			Contract: blocktx.Contractaddress,
			Fee:      blocktx.SysFee,
			Amount:   blocktx.Amount,
			Memo:     blocktx.Memo,
		})
	}
	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return fmt.Errorf("marshal err %v", err)
	}
	log.Info(string(pusdata))

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
	if tx.Contractaddress != "" && !p.watch.IsContractExist(tx.Contractaddress) {
		return
	}

	if p.watch.IsWatchAddressExist(tx.Fromaccount) {
		watchaddrs[tx.Fromaccount] = true
	}
	if p.watch.IsWatchAddressExist(tx.Toaccount) {
		watchaddrs[tx.Toaccount] = true
	}
	return
}
func (p *Processor) parseContractTxs(txs []*dao.BlockTx) (watchaddrs map[string]map[string]bool, watchtxs map[string][]*dao.BlockTx) {
	//txj, _ := json.Marshal(txs)
	//log.Info(string(txj))
	watchaddrs = make(map[string]map[string]bool)
	watchtxs = make(map[string][]*dao.BlockTx)
	for _, tx := range txs {
		if tx.Contractaddress != "" && !p.watch.IsContractExist(tx.Contractaddress) {
			return
		}

		if p.watch.IsWatchAddressExist(tx.Fromaccount) {
			if _, ok := watchaddrs[tx.Txid]; !ok {
				watchaddrs[tx.Txid] = make(map[string]bool)
			}
			watchaddrs[tx.Txid][tx.Fromaccount] = true
		}
		if p.watch.IsWatchAddressExist(tx.Toaccount) {
			if _, ok := watchaddrs[tx.Txid]; !ok {
				watchaddrs[tx.Txid] = make(map[string]bool)
			}
			watchaddrs[tx.Txid][tx.Toaccount] = true
		}
		if p.watch.IsWatchAddressExist(tx.Fromaccount) || p.watch.IsWatchAddressExist(tx.Toaccount) {
			watchtxs[tx.Txid] = append(watchtxs[tx.Txid], tx)
		}
	}

	return
}
func (p *Processor) UpdateAmount(addr string) error {

	return nil
}

func (s *Processor) RepushTx(userid int64, height int64, txid string) error {
	var (
		err     error
		txinfos *dao.BlockTx
	)
	log.Infof("RepushTx user: %d , txid : %s", userid, txid)

	if txid == "" || height == 0 {
		return fmt.Errorf("txid,height 不能为空")
	}

	if txinfos, err = s.getBlockTxFromNode(height, txid); err != nil {
		return err
	}

	log.Info(String(txinfos))
	bestBlockHeight, err := s.GetBestHeight()
	if err != nil {
		return err
	}
	err = fmt.Errorf("txid %s don't have care of address", txid)
	watchaddrses, txinfosmap := s.parseContractTxs([]*dao.BlockTx{txinfos})
	for txhash, watchaddrs := range watchaddrses {
		err = s.processPushs(txinfosmap[txhash], watchaddrs, bestBlockHeight)
	}
	return err
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
//		if s.watch.IsWatchAddressExist(txvout.Address) {
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
//				if s.watch.IsWatchAddressExist(txvin.Address) {
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
//				if s.watch.IsWatchAddressExist(txvin.Address) {
//					insertVins = append(insertVins, txvin)
//				}
//			}
//		}
//
//		amtin = amtin.Add(txvin.Value)
//		if s.watch.IsWatchAddressExist(txvin.Address) {
//			tmpWatchList[txvin.Address] = true
//		}
//	}
//
//	txInfo.tx.Fee = amtin.Sub(amtout)
//	if txInfo.tx.Fee.IsNegative() && s.conf.Name != "doge" {
//		return nil, nil, nil, fmt.Errorf("tx fee don't allow negative,vin:%v, vout:%v", amtin, amtout)
//	}
//
//	//log.Printf("processTX : %s ,vins : %d , vouts : %d , used time : %f 's", blocktx.Txid, len(txvins), len(insertVouts), time.Since(starttime).Seconds())
//	return tmpWatchList, updateVins, insertVins, nil
//}
func (s *Processor) getBlockTxFromNode(height int64, txid string) (*dao.BlockTx, error) {
	blockhash,err := s.GetBlockHash(height)
	if err != nil {
		return nil,err
	}
	//log.Info(blockhash)
	block,err :=s.Block(blockhash)
	if err != nil {
		return nil,err
	}
	txIndex := -1
	for k,tmptx := range block.Extrinsics{
		if strings.ToLower(tmptx.ExtrinsicHash) == strings.ToLower(txid){
			txIndex = k
			break
		}
	}
	if txIndex < 0 {
		return nil,errors.New("交易没找到")
	}
	tx,err := block.ToTransaction(txIndex)
	if err != nil{
		return nil,err
	}
	if tx == nil {
		return nil,errors.New("不支持的交易类型")
	}
	if !tx.Status{
		return nil,errors.New("失败的交易")
	}

	tmpDaoTx :=  &dao.BlockTx{
		Height: tx.BlockHeight,
		Hash:tx.BlockHash,
		Txid:tx.Txid,
		Fromaccount: tx.From,
		Toaccount: tx.To,
		Amount:tx.Value,
		SysFee :tx.Fee,
	}
	return tmpDaoTx,nil
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
		Time:          task.Block.Time.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}

//解析交易
