package tkm

import (
	"errors"
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/providers"
)

type TkmService struct {
	cfg     conf.Config
	nodeCfg conf.NodeConfig
	client  *web3.Web3
}

func (ts *TkmService) GetHeightByTxid(txid string) (int64, error) {
	return 0, errors.New("unsupport it")
}

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	cs := new(TkmService)
	cs.cfg = cfg
	cs.nodeCfg = nodeCfg
	cs.client = web3.NewWeb3(providers.NewHTTPProvider(nodeCfg.Url, 60, false))
	return cs
}
func (ts *TkmService) GetTxIsExist(height int64, txid string) bool {
	panic("implement me")
}

func (cs *TkmService) GetLatestBlockHeight() (int64, error) {
	panic("implement me")
}

func (ts *TkmService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	panic("implement me")
}

func (ts *TkmService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {
	panic("implement me")
}
