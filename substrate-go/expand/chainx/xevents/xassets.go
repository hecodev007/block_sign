package xevents

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

/// XAssets Type
type XAssets struct {
	XAssets_Moved      []EventXAssetsMoved
	XAssets_Issued     []EventXAssetsIssued
	XAssets_Destroyed  []EventXAssetsDestroyed
	XAssets_BalanceSet []EventXAssetsBalanceSet
}

type AssetId types.U32
type AssetType uint8

const (
	Usable             AssetType = 0
	Locked             AssetType = 1
	Reserved           AssetType = 2
	ReservedWithdrawal AssetType = 3
	ReservedDexSpot    AssetType = 4
)

/// Some balances of an asset was moved from one to another. [asset_id, from, from_type, to, to_type, amount]
type EventXAssetsMoved struct {
	Phase    types.Phase
	AssetId  AssetId
	From     types.AccountID
	FromType AssetType
	To       types.AccountID
	ToType   AssetType
	Balance  types.U128
	Topics   []types.Hash
}

/// New balances of an asset were issued. [asset_id, receiver, amount]
type EventXAssetsIssued struct {
	Phase    types.Phase
	AssetId  AssetId
	Receiver types.AccountID
	Amount   types.U128
	Topics   []types.Hash
}

/// Some balances of an asset were destoryed. [asset_id, who, amount]
type EventXAssetsDestroyed struct {
	Phase   types.Phase
	AssetId AssetId
	Who     types.AccountID
	Amount  types.U128
	Topics  []types.Hash
}

/// Set asset balance of an account by root. [asset_id, who, asset_type, amount]
type EventXAssetsBalanceSet struct {
	Phase     types.Phase
	AssetId   AssetId
	Who       types.AccountID
	AssetType AssetType
	Balance   types.U128
	Topics    []types.Hash
}
