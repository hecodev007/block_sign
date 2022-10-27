package ksm

import (
	"errors"
	"github.com/centrifuge/go-substrate-rpc-client/v3/client"
	"github.com/centrifuge/go-substrate-rpc-client/v3/rpc/chain"
)

type DotNodeRpc struct {
	Url    string
	Chain  *chain.Chain
	Client client.Client
}

func NewDotNodeRpc(url string) (*DotNodeRpc, error) {
	cl, err := client.Connect(url)
	if err != nil {
		return nil, err
	}
	chain := chain.NewChain(cl)
	return &DotNodeRpc{
		Url:    url,
		Chain:  chain,
		Client: cl,
	}, nil
}

func (node *DotNodeRpc) GetBlockHash(height int64) (blockhash string, err error) {
	hash, err := node.Chain.GetBlockHash(uint64(height))
	if err != nil {
		return "", err
	}
	if hash.Hex() == "" {
		return "", errors.New("获取hash失败")
	}
	return hash.Hex(), nil
}
func (node *DotNodeRpc) LatestBlock() (height int64, err error) {
	block, err := node.Chain.GetBlockLatest()
	if err != nil {
		return 0, err
	}
	return int64(block.Block.Header.Number), nil
}
