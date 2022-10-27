package common

import "github.com/group-coldwallet/scanning-service/models/po"

// 判断是否是合约交易
type IsContractTx func(contractAddress string) (*po.ContractInfo, bool)

type IsWatchAddress func(address string) bool

type IScanner interface {
	GetLatestBlockHeight() (int64, error)
	GetBlockByHeight(height int64) (*BlockData, error)
	GetTxData(blockData *BlockData, txid string, isWatchAddress IsWatchAddress, isContractTx IsContractTx) (*TxData, error)
	GetTxIsExist(height int64, txid string) bool
	GetHeightByTxid(txid string) (int64, error)
}

type BlockData struct {
	Height       int64
	Hash         string
	PrevHash     string
	NextHash     string
	Timestamp    int64
	TxNums       int
	Confirmation int64
	TxIds        []string
	TxDatas      []*TxData //如果可以根据高度把所有的交易解析出来，而不需要一个txid去查询，就放在这里处理
}
type TxData struct {
	Height          int64
	Txid            string
	IsFakeTx        bool   //是否是假充值
	FromAddr        string //下标对应地址
	ToAddr          string
	Amount          string
	Fee             string
	Memo            string
	ContractAddress string
	IsContainTx     bool
	MainDecimal     int32 // 主币的精度，  ***注意不是代币的精度，仅是主币的精度
}
