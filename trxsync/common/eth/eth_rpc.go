package eth

import (
	"encoding/hex"
	"fmt"
	"github.com/group-coldwallet/trxsync/common"
	"golang.org/x/crypto/sha3"
	"time"
)

const (
	ContractTransfer     = "transfer(address,uint256)"
	ContractBalanceOf    = "balanceOf(address)"
	ContractDecimals     = "decimals()"
	ContractAllowance    = "allowance(address,address)"
	ContractSymbol       = "symbol()"
	ContractTotalSupply  = "totalSupply()"
	ContractName         = "name()"
	ContractApprove      = "approve(address,uint256)"
	ContractTransferFrom = "transferFrom(address,address,uint256)"
)

type EthRpcClient struct {
	url, user, psd string
	c              *common.Client
}

func NewEthRpcClient(url, user, password string) *EthRpcClient {
	e := new(EthRpcClient)
	e.user = user
	e.url = url
	e.psd = password
	c, _ := common.NewRpcClient(url, user, password)
	e.c = c
	return e
}

func (e *EthRpcClient) initEthRpcClient() error {
	var err error
	if e.c == nil {
		e.c, err = common.NewRpcClient(e.url, e.user, e.psd)
		if err != nil {
			// 避免接口出错了循环的调用，所以休眠3秒
			time.Sleep(3 * time.Second)
			return fmt.Errorf("init rpc client error: %v", err)
		}
	}
	return nil
}
func (e *EthRpcClient) GetLatestBlockHeight() (int64, error) {
	err := e.initEthRpcClient()
	if err != nil {
		return 0, err
	}
	var heightResp string
	err = e.c.Post("eth_blockNumber", &heightResp, nil)
	if err != nil || heightResp == "" {
		return 0, fmt.Errorf("rpc get latest height error: %v", err)
	}
	var height int64
	height, err = common.ParseInt64(heightResp)
	if err != nil {
		return 0, fmt.Errorf("parse height hex error: %v", err)
	}
	return height, nil
}
func (e *EthRpcClient) GetBlockByNumber(height int64, resp interface{}, withTx bool) error {
	err := e.initEthRpcClient()
	if err != nil {
		return err
	}
	var (
		params []interface{}
	)
	params = append(params, common.Int64ToHex(height))
	params = append(params, withTx)

	err = e.c.Post("eth_getBlockByNumber", resp, params)
	if err != nil {
		return fmt.Errorf("rpc get block by height(%d) error: %v", height, err)
	}
	if resp == nil {
		return fmt.Errorf("get block by height(%d) is null", height)
	}
	return nil
}

func (e *EthRpcClient) GetTransactionReceipt(txid string, resp interface{}) error {
	err := e.initEthRpcClient()
	if err != nil {
		return err
	}
	err = e.c.Post("eth_getTransactionReceipt", resp, []interface{}{txid})
	if err != nil {
		return fmt.Errorf("rpc get transaction receipt error: %v", err)
	}
	return nil
}

func (e *EthRpcClient) GetTransactionByHash(txid string, resp interface{}) error {
	err := e.initEthRpcClient()
	if err != nil {
		return err
	}
	err = e.c.Post("eth_getTransactionByHash", resp, []interface{}{txid})
	if err != nil {
		return fmt.Errorf("rpc get transaction by hash error: %v", err)
	}
	return nil
}

func (e *EthRpcClient) EthCall(req, resp interface{}) error {
	err := e.initEthRpcClient()
	if err != nil {
		return err
	}
	err = e.c.Post("eth_call", resp, req)
	if err != nil {
		return fmt.Errorf("rpc eth_call error: %v", err)
	}
	return nil
}

func (e *EthRpcClient) GetMethodID(method string) string {
	data := e.keccak256([]byte(method))
	return "0x" + hex.EncodeToString(data[:4])
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func (e *EthRpcClient) keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func (e *EthRpcClient) SendRawTransaction(rawTx string, resp interface{}) error {
	err := e.initEthRpcClient()
	if err != nil {
		return err
	}
	err = e.c.Post("eth_sendRawTransaction", resp, []interface{}{rawTx})
	if err != nil {
		return fmt.Errorf("rpc send raw transaction error: %v", err)
	}
	return nil
}

func (e *EthRpcClient) GetCode(address string, resp interface{}) error {
	err := e.initEthRpcClient()
	if err != nil {
		return err
	}
	err = e.c.Post("eth_getCode", resp, []interface{}{address, "latest"})
	if err != nil {
		return fmt.Errorf("rpc get code error: %v", err)
	}
	return nil
}
