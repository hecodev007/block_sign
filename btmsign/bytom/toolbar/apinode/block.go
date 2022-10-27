package apinode

import (
	"encoding/json"

	"btmSign/bytom/api"
	"btmSign/bytom/errors"
	"btmSign/bytom/protocol/bc/types"
)

func (n *Node) GetBlockByHash(hash string) (*types.Block, error) {
	return n.getRawBlock(&getRawBlockReq{BlockHash: hash})
}

func (n *Node) GetBlockByHeight(height uint64) (*types.Block, error) {
	return n.getRawBlock(&getRawBlockReq{BlockHeight: height})
}

type getRawBlockReq struct {
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
}

func (n *Node) getRawBlock(req *getRawBlockReq) (*types.Block, error) {
	url := "/get-raw-block"
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal")
	}
	resp := &api.GetRawBlockResp{}
	return resp.RawBlock, n.request(url, payload, resp)
}

// bytomChainStatusResp is the response of bytom chain status
type bytomChainStatusResp struct {
	FinalizedHeight uint64 `json:"finalized_height"`
}

// GetFinalizedHeight return the finalized block height of connected node
func (n *Node) GetFinalizedHeight() (uint64, error) {
	url := "/chain-status"
	res := &bytomChainStatusResp{}
	return res.FinalizedHeight, n.request(url, nil, res)
}
