package tmpl

import (
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"sync"
)

/*
模版
*/

type TmplService struct {
	cfg          conf.Config
	nodeCfg      conf.NodeConfig
	latestHeight int64
	url          string
	lock         sync.RWMutex
	// add custom filed

}

func (ts *TmplService) GetLatestBlockHeight() (int64, error) {
	panic("implement me")
}

func (ts *TmplService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	panic("implement me")
}

func (ts *TmplService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {
	panic("implement me")
}

func (ts *TmplService) GetHeightByTxid(txid string) (int64, error) {
	return 0, nil
}

func (ts *TmplService) GetTxIsExist(height int64, txid string) bool {
	return true
}

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	ts := new(TmplService)
	ts.cfg = cfg
	ts.nodeCfg = nodeCfg
	ts.url = nodeCfg.Url
	ts.lock = sync.RWMutex{}
	//todo init other filed
	return ts
}
