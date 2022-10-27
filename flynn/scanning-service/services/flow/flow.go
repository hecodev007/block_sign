package flow

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"strings"
	"sync"
)

type FlowService struct {
	cfg          conf.Config
	nodeCfg      conf.NodeConfig
	latestHeight int64
	url          string
	lock         sync.RWMutex
	// add custom filed
	flowClient *client.Client
	ctx        context.Context
}

const (
	Deposit   = "A.1654653399040a61.FlowToken.TokensDeposited"
	Withdrawn = "A.1654653399040a61.FlowToken.TokensWithdrawn"
)

func (ts *FlowService) GetLatestBlockHeight() (int64, error) {
	if ts.flowClient == nil {
		c, err := client.New(ts.url, grpc.WithInsecure())
		if err != nil {
			panic(fmt.Sprintf("init flow client error: %v", err))
		}
		ts.flowClient = c
	}
	header, err := ts.flowClient.GetLatestBlockHeader(ts.ctx, true)
	if err != nil {
		return 0, err
	}
	if header == nil {
		return 0, errors.New("resp header is nil")
	}
	ts.latestHeight = int64(header.Height)
	return int64(header.Height), nil

}

func (ts *FlowService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	block, err := ts.flowClient.GetBlockByHeight(ts.ctx, uint64(height))
	if err != nil || block == nil {
		return nil, fmt.Errorf("get block by height error: %v,height: %d", err, height)
	}
	bd := new(common.BlockData)
	bd.Height = int64(block.BlockHeader.Height)
	bd.Hash = block.BlockHeader.ID.Hex()
	bd.PrevHash = block.BlockHeader.ParentID.Hex()
	bd.Timestamp = block.Timestamp.Unix()
	bd.Confirmation = ts.latestHeight - height
	//获取交易的数量
	if len(block.CollectionGuarantees) == 0 {
		bd.TxNums = 0
		return bd, nil
	}
	for _, cg := range block.CollectionGuarantees {
		collection, err := ts.flowClient.GetCollection(ts.ctx, cg.CollectionID)
		if err != nil {
			log.Errorf("height %d get collection [%s] error: %v",
				height, cg.CollectionID.Hex(), err)
			continue
		}
		if len(collection.TransactionIDs) == 0 {
			continue
		}
		for _, txid := range collection.TransactionIDs {
			bd.TxIds = append(bd.TxIds, txid.Hex())
		}
	}
	return bd, nil
}

func (ts *FlowService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {
	txid = strings.TrimPrefix(txid, "0x")
	id := ts.txidToFlowIdentifier(txid)
	tx, err := ts.flowClient.GetTransactionResult(ts.ctx, id)
	if err != nil || tx == nil {
		return nil, fmt.Errorf("get transaction result error: %v", err)
	}
	if tx.Error != nil {
		return nil, fmt.Errorf("get transaction result have error: %v", tx.Error)
	}
	td := new(common.TxData)
	td.IsFakeTx = true
	if tx.Status == flow.TransactionStatusSealed {
		td.IsFakeTx = false
	}
	//处理event
	var (
		from, to, fAmount, tAmount string
	)
	for _, event := range tx.Events {
		value := event.Value
		//fmt.Println(event.Type)
		switch event.Type {
		case Withdrawn:
			// 这里面获取from地址
			from, fAmount, err = ts.parseValue(&value)

		case Deposit:
			// 这里面获取to地址
			to, tAmount, err = ts.parseValue(&value)
		}
		if err != nil {
			log.Errorf("parse value error: %v", err)
			continue
		}
	}
	//fmt.Printf("from: %s,to: %s,famount: %s,toAmt: %s\n",from,to,fAmount,tAmount)
	if ts.validAddress(from) && ts.validAddress(to) && from != to {
		// 判断是否是监听的地址
		if isWatchAddress(from) || isWatchAddress(to) {
			td.IsContainTx = true
			if fAmount != tAmount {
				//如果from的amount和to的amount不相等，这里就有问题了，有可能是解析出错了，或者只有里面的一个事件，
				// 这里我就把他当作假充值去处理
				td.IsFakeTx = true
				return td, nil
			}
			td.FromAddr = from
			td.ToAddr = to
			td.Amount = fAmount
			memo, err := ts.parseAmount(fAmount)
			if err != nil {
				return nil, err
			}
			if isWatchAddress(to) {
				td.Memo = memo
			}
		}
	}
	td.Txid = txid

	return td, nil
}

func (ts *FlowService) GetHeightByTxid(txid string) (int64, error) {
	return 0, errors.New("unsupport get height by txid")
}

func (ts *FlowService) GetTxIsExist(height int64, txid string) bool {
	txid = strings.TrimPrefix(txid, "0x")
	id := ts.txidToFlowIdentifier(txid)
	tx, err := ts.flowClient.GetTransactionResult(ts.ctx, id)
	if err != nil || tx == nil {
		return false
	}
	if tx.Error != nil {
		return false
	}
	if tx.Status == flow.TransactionStatusSealed {
		return true
	}
	return false
}

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	ts := new(FlowService)
	ts.cfg = cfg
	ts.nodeCfg = nodeCfg
	ts.lock = sync.RWMutex{}
	//todo init other filed
	ts.url = strings.TrimPrefix(nodeCfg.Url, "http://")
	ts.flowClient, _ = client.New(ts.url, grpc.WithInsecure())
	ts.ctx = context.Background()
	return ts
}

func (ts *FlowService) txidToFlowIdentifier(txid string) flow.Identifier {
	txid = strings.TrimPrefix(txid, "0x")
	return flow.HexToID(txid)
}

func (ts *FlowService) parseValue(value *cadence.Event) (address, amount string, err error) {
	// 正常的field里面只包含address和amount
	if len(value.Fields) != 2 {
		err = fmt.Errorf("value field的长度不等于2，无法确定里面含有什么数据：%d", len(value.Fields))
		return "", "", err
	}
	var hasErr = false
	for _, field := range value.Fields {
		if field.Type().ID() == "UFix64" {
			amt := field.ToGoValue().(uint64)
			amount = decimal.NewFromInt(int64(amt)).Shift(-8).String()
			//amount = field.String()

		} else {
			//这里其实有点问题，应该判断field.Type().ID()=="Address"，但是这个字段解析出来是{}?，不等于Address,所以无法判断
			if ts.validAddress(field.String()) {
				address = field.String()
			} else {
				addrBytes, ok := field.ToGoValue().([8]byte)
				if !ok {
					//地址解析出错了
					hasErr = true
					address = field.String() //为了表示是哪一个地址出错
				} else {
					address = "0x" + hex.EncodeToString(addrBytes[:])

				}

			}

		}
	}
	if hasErr {
		return "", "", fmt.Errorf("解析地址出错： %s", address)
	}
	return address, amount, nil
}

func (ts *FlowService) validAddress(address string) bool {
	if len(address) != 18 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}

	return true
}

/*
解析amount和memo
*/
func (ts *FlowService) parseAmount(amount string) (string, error) {
	if !strings.Contains(amount, ".") {
		return "", nil
	}
	ss := strings.Split(amount, ".")
	if len(ss) != 2 {
		return "", fmt.Errorf("parse amount error: %s", amount)
	}
	return ss[1], nil
}
