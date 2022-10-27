package fio

import (
	"encoding/json"
	"errors"
	"fmt"
	gofio "github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"github.com/shopspring/decimal"
	"rsksync/common"
	"rsksync/conf"
	dao "rsksync/models/po/fio"
	"strings"
	"sync"
	"time"
)

type Scanner struct {
	baseUrl string
	client  *gofio.API
	lock    *sync.Mutex
	conf    conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	client, _, _ := gofio.NewConnection(nil, node.Url)
	return &Scanner{
		baseUrl: node.Url,
		client:  client,
		lock:    &sync.Mutex{},
		conf:    conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.DeleteBlockInfo(height)
	dao.DeleteBlockTX(height)
}

//爬数据
func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

func (s *Scanner) reConnect() error {
	if s.client == nil {
		client, _, err := gofio.NewConnection(nil, s.baseUrl)
		if err != nil {
			return err
		}
		s.client = client
	}
	return nil
}
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	err := s.reConnect()
	if err != nil {
		return 0, err
	}
	info, err := s.client.GetInfo()
	if err != nil || info == nil {
		return 0, nil
	}
	return int64(info.LastIrreversibleBlockNum), nil
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.GetMaxBlockIndex()
}

//批量扫描多个区块
func (s *Scanner) BatchScanIrreverseBlocks(startHeight, endHeight, bestHeight int64) *sync.Map {
	starttime := time.Now()
	count := endHeight - startHeight
	taskmap := &sync.Map{}
	wg := &sync.WaitGroup{}

	wg.Add(int(count))
	for i := int64(0); i < count; i++ {
		height := startHeight + i
		go func(w *sync.WaitGroup) {
			if task, err := s.ScanIrreverseBlock(height, bestHeight); err == nil {
				taskmap.Store(height, task)
			}
			w.Done()
		}(wg)
	}
	wg.Wait()
	log.Debugf("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
	return taskmap
}

func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()
	log.Infof("scanBlock %d ", height)
	block, err := s.client.GetBlockByNum(uint32(height))
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := dao.GetBlockCountByHash(block.ID.String())
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &ProcTask{
		bestHeight: bestHeight,
		block: &dao.BlockInfo{
			Height:         int64(block.BlockNum),
			Hash:           block.ID.String(),
			FrontBlockHash: block.Previous.String(),
			Timestamp:      block.Timestamp.Time,
			Transactions:   len(block.Transactions),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	log.Infof("tx len %d", len(block.Transactions))
	//处理区块内的交易
	if len(block.Transactions) > 0 {
		if s.conf.EnableGoroutine {
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Transactions))
			for _, tmp := range block.Transactions {
				tx := tmp
				go s.batchParseTx(&tx, task, wg)
			}
			wg.Wait()
		} else {
			for _, tx := range block.Transactions {
				blockTx, err := s.parseBlockTX(&tx, task.block)
				if err == nil {
					task.txInfos = append(task.txInfos, blockTx)
				} else {
					log.Errorf("process tx error,Err=[%v]", err)
				}
			}
		}
	}
	d, _ := json.Marshal(task)
	fmt.Println(string(d))
	log.Infof("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *eos.TransactionReceipt, task *ProcTask, w *sync.WaitGroup) {
	defer w.Done()

	blockTx, err := s.parseBlockTX(tx, task.block)
	if err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
	if err != nil {
		log.Errorf("process tx error,err=[%v]", err)
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *eos.TransactionReceipt, block *dao.BlockInfo) (*dao.BlockTX, error) {
	log.Infof("start处理交易，txid=[%s]", tx.Transaction.ID.String())
	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}

	if tx.Transaction.Packed == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed is null")
	}

	if tx.Transaction.Packed.PackedTransaction == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed.Transaction is null")
	}

	blocktx := &dao.BlockTX{
		CoinName:    s.conf.Name,
		Txid:        tx.Transaction.ID.String(),
		BlockHeight: block.Height,
		BlockHash:   block.Hash,
		Status:      tx.Status.String(),
		Timestamp:   block.Timestamp,
		CreateTime:  time.Now(),
	}
	//处理action
	// 1。 根据txid查找交易
	out, err := s.client.GetTransaction(tx.Transaction.ID)
	if err != nil || out == nil {
		return nil, fmt.Errorf("get tx by txid=[%s] error,err=[%v]", tx.Transaction.ID.String(), err)
	}
	if out.ID == nil {
		d, _ := json.Marshal(out)
		log.Debug(string(d))
		return nil, fmt.Errorf("get tx error,txid=[%s]", tx.Transaction.ID.String())
	}
	if &out.Transaction == nil {
		return nil, fmt.Errorf("tx1 is null,txid=[%s]", tx.Transaction.ID.String())
	}
	if &out.Transaction.Transaction == nil {
		return nil, fmt.Errorf("tx2 is null,txid=[%s]", tx.Transaction.ID.String())
	}
	if out.Transaction.Transaction.Transaction == nil {
		return nil, fmt.Errorf("tx3 is null,txid=[%s]", tx.Transaction.ID.String())
	}
	if len(out.Transaction.Transaction.Actions) <= 0 {
		log.Infof("do not find any action,txid=[%s]", tx.Transaction.ID.String())
		return nil, fmt.Errorf("do not find any action,txid=[%s]", tx.Transaction.ID.String())
	}
	for _, action := range out.Transaction.Transaction.Actions {
		if err := parseActionForBlocktx(s.client, blocktx, action); err != nil {
			log.Errorf("parse action error,Err=[%v]", err)
			continue
		}
	}

	if blocktx.FromAddress == "" || blocktx.ToAddress == "" {
		return nil, fmt.Errorf("tx. from : %s , to :%s", blocktx.FromAddress, blocktx.ToAddress)
	}

	data, _ := json.Marshal(tx)
	if len(data) < 8000 {
		blocktx.TxJson = string(data)
	}
	log.Infof("end处理交易，txid=[%s]", tx.Transaction.ID.String())
	return blocktx, nil
}

func parseActionForBlocktx(client *gofio.API, blocktx *dao.BlockTX, action *eos.Action) error {
	if err := action.MapToRegisteredAction(); err != nil {
		//log.Infof("action MapToRegistered err : %v", err)
		return fmt.Errorf("action MapToRegistered err : %v", err)
	}

	if action.Name != "trnsfiopubky" {
		return fmt.Errorf("don't support action name : %v", action.Name)
	}

	blocktx.ContractAddress = string(action.Account)

	for k, v := range action.Data.(map[string]interface{}) {
		switch k {
		case "actor":
			fromAccount, ok := v.(string)
			if !ok {
				return fmt.Errorf("from address type err: %T", v)
			}
			//根据账户名获取from地址
			from, err := getAddressbyAccountName(client, fromAccount)
			if err != nil {
				return err
			}
			blocktx.FromAddress = from
		case "payee_public_key":
			to, ok := v.(string)
			if !ok {
				return fmt.Errorf("to address type err: %T", v)
			}
			blocktx.ToAddress = to
		case "max_fee":
			fee, ok := v.(float64)
			if !ok {
				return fmt.Errorf("fee type err: %T", v)
			}

			blocktx.Fee = decimal.NewFromFloat(fee).Shift(-9)
		case "amount":
			amountF, ok := v.(float64)
			if !ok {
				return fmt.Errorf("fee type err: %T", v)
			}
			amount := decimal.NewFromFloat(amountF).Shift(-9)
			blocktx.Amount = amount
		}
	}
	return nil
}

func getAddressbyAccountName(client *gofio.API, accountName string) (string, error) {
	if strings.HasPrefix(accountName, "FIO") {
		return accountName, nil
	}
	if len(accountName) != 12 {
		return "", errors.New("account name length is not equal 12")
	}
	out, err := client.GetFioAccount(accountName)
	if err != nil || out == nil {
		return "", fmt.Errorf("get fio account info error,err=[%v]", err)
	}
	var pubKey string
	for _, p := range out.Permissions {
		if p.PermName == "active" {
			if len(p.RequiredAuth.Keys) == 0 {
				if len(p.RequiredAuth.Accounts) > 0 {
					return "", fmt.Errorf("[%s] is a multi account", accountName)
				}
				return "", fmt.Errorf("parse [%s] account error", accountName)
			}
			pubKey = p.RequiredAuth.Keys[0].PublicKey.String()
			break
		}
	}
	return pubKey, nil
}
