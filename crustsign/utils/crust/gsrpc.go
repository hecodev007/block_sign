package crust

import (
	"fmt"
	"strings"

	gsrc "github.com/yanyushr/go-substrate-rpc-client/v3"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
)

type Client struct {
	Api *gsrc.SubstrateAPI

	Meta               *types.Metadata
	prefix             []byte //币种的前缀
	ChainName          string //链名字
	SpecVersion        int
	TransactionVersion int
	genesisHash        types.Hash
	Url                string
}

func NewClient(url string) (*Client, error) {
	c := new(Client)
	c.Url = url
	var err error

	// 初始化rpc客户端
	c.Api, err = gsrc.NewSubstrateAPI(url)
	if err != nil {
		return nil, err
	}
	//检查当前链运行的版本
	err = c.checkRuntimeVersion()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) GetAccountInfo(address string, meta *types.Metadata) (types.AccountInfo, error) {
	var accountInfo types.AccountInfo

	pubKey := GetPublicFromAddr(address, CRustPrefix)

	key, err := types.CreateStorageKey(meta, "System", "Account", pubKey, nil)
	if err != nil {
		return accountInfo, err
	}

	_, err = c.Api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return accountInfo, err
	}
	return accountInfo, nil
}

func (c *Client) GetGenesisHash() types.Hash {
	if c.genesisHash.Hex() != "0x0000000000000000000000000000000000000000000000000000000000000000" {
		return c.genesisHash
	}
	hash, err := c.Api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return types.Hash{}
	}
	c.genesisHash = hash
	return hash
}

func (c *Client) reConnect() error {
	api, err := gsrc.NewSubstrateAPI(c.Url)
	if err != nil {
		return err
	}
	c.Api = api
	return nil
}

func (c *Client) checkRuntimeVersion() error {
	v, err := c.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		if !strings.Contains(err.Error(), "tls: use of closed connection") {
			return fmt.Errorf("init runtime version error,err=%v", err)
		}
		//	重连处理，这是因为第三方包的问题，所以只能这样处理了了
		err := c.reConnect()
		if err != nil {
			return fmt.Errorf("reconnect error: %v", err)
		}

		v, err = c.Api.RPC.State.GetRuntimeVersionLatest()
		if err != nil {
			return fmt.Errorf("init runtime version error,aleady reconnect,err: %v", err)
		}
	}
	c.TransactionVersion = int(v.TransactionVersion)
	c.ChainName = v.SpecName
	specVersion := int(v.SpecVersion)
	//检查metadata数据是否有升级
	if specVersion != c.SpecVersion {
		c.Meta, err = c.Api.RPC.State.GetMetadataLatest()
		if err != nil {
			return fmt.Errorf("init metadata error: %v", err)
		}
		c.SpecVersion = specVersion
	}
	return nil
}

func (c *Client) SetGenesisHash(hash types.Hash) {
	c.genesisHash = hash
}
