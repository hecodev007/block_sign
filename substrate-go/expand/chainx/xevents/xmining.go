package xevents

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

///XMining Type
type XMiningAsset struct {
	XMiningAsset_Claimed []EventXMiningAssetClaimed
	XMiningAsset_Minted  []EventXMiningAssetMinted
}

type XStaking struct {
	XStaking_Minted            []EventXStakingMinted
	XStaking_Slashed           []EventXStakingSlashed
	XStaking_Bonded            []EventXStakingBonded
	XStaking_Rebonded          []EventXStakingRebonded
	XStaking_Unbonded          []EventXStakingUnbonded
	XStaking_Claimed           []EventXStakingClaimed
	XStaking_Withdrawn         []EventXStakingWithdrawn
	XStaking_ForceChilled      []EventXStakingForceChilled
	XStaking_ForceAllWithdrawn []EventXStakingForceAllWithdrawn
}

type SessionIndex uint32

/// An asset miner claimed the mining reward. [claimer, asset_id, amount]
type EventXMiningAssetClaimed struct {
	Phase   types.Phase
	Claimer types.AccountID
	AssetId types.U32
	Amount  types.U128
	Topics  []types.Hash
}

/// Issue new balance to the reward pot. [reward_pot_account, amount]
type EventXMiningAssetMinted struct {
	Phase            types.Phase
	RewardPotAccount types.AccountID
	Amount           types.U128
	Topics           []types.Hash
}

/// Issue new balance to this account. [account, reward_amount]
type EventXStakingMinted struct {
	Phase        types.Phase
	Account      types.AccountID
	RewardAmount types.U128
	Topics       []types.Hash
}

/// A validator (and its reward pot) was slashed. [validator, slashed_amount]
type EventXStakingSlashed struct {
	Phase         types.Phase
	Validator     types.AccountID
	SlashedAmount types.U128
	Topics        []types.Hash
}

/// A nominator bonded to the validator this amount. [nominator, validator, amount]
type EventXStakingBonded struct {
	Phase     types.Phase
	Nominator types.AccountID
	Validator types.AccountID
	Amount    types.U128
	Topics    []types.Hash
}

/// A nominator switched the vote from one validator to another. [nominator, from, to, amount]
type EventXStakingRebonded struct {
	Phase     types.Phase
	Nominator types.AccountID
	From      types.AccountID
	To        types.AccountID
	Amount    types.U128
	Topics    []types.Hash
}

/// A nominator unbonded this amount. [nominator, validator, amount]
type EventXStakingUnbonded struct {
	Phase     types.Phase
	Nominator types.AccountID
	Validator types.AccountID
	Amount    types.U128
	Topics    []types.Hash
}

/// A nominator claimed the staking dividend. [nominator, validator, dividend]
type EventXStakingClaimed struct {
	Phase     types.Phase
	Nominator types.AccountID
	Validator types.AccountID
	Dividend  types.U128
	Topics    []types.Hash
}

/// The nominator withdrew the locked balance from the unlocking queue. [nominator, amount]
type EventXStakingWithdrawn struct {
	Phase     types.Phase
	Nominator types.AccountID
	Amount    types.U128
	Topics    []types.Hash
}

/// Offenders were forcibly to be chilled due to insufficient reward pot balance. [session_index, chilled_validators]
type EventXStakingForceChilled struct {
	Phase             types.Phase
	SessionIndex      SessionIndex
	ChilledValidators []types.AccountID
	Topics            []types.Hash
}

/// Unlock the unbonded withdrawal by force. [account]
type EventXStakingForceAllWithdrawn struct {
	Phase   types.Phase
	Account types.AccountID
	Topics  []types.Hash
}
