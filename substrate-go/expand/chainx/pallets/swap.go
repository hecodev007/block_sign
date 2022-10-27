package pallets

import "github.com/centrifuge/go-substrate-rpc-client/v3/types"

type AssetId types.U32
type Balance types.U128

/// XAssets Type
type Swap struct {
	Swap_PairCreated		[]EventPairCreated
	Swap_LiquidityAdded		[]EventLiquidityAdded
	Swap_LiquidityRemoved	[]EventLiquidityRemoved
	Swap_TokenSwap			[]EventTokenSwap
}

/// Create a trading pair. \[creator, asset_id, asset_id\]
type EventPairCreated struct {
	Phase    	types.Phase
	Creator     types.AccountID
	AssetA  	AssetId
	AssetB  	AssetId
	Topics   	[]types.Hash
}

/// Add liquidity. \[owner, asset_id, asset_id\]
type EventLiquidityAdded struct {
	Phase    	types.Phase
	Owner     	types.AccountID
	AssetA  	AssetId
	AssetB  	AssetId
	Topics   	[]types.Hash
}

/// Remove liquidity. \[owner, recipient, asset_id, asset_id, amount\]
type EventLiquidityRemoved struct {
	Phase    	types.Phase
	Owner     	types.AccountID
	Recipient   types.AccountID
	AssetA  	AssetId
	AssetB  	AssetId
	Amount 		Balance
	Topics   	[]types.Hash
}

/// Transact in trading \[owner, recipient, swap_path\]
type EventTokenSwap struct {
	Phase    	types.Phase
	Owner     	types.AccountID
	Recipient   types.AccountID
	SwapPath  	[]AssetId
	Topics   	[]types.Hash
}