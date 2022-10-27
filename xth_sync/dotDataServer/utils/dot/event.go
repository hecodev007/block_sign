package dot

import (
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
)

type EventUtilityBatchInterrupted struct {
	Phase  types.Phase
	Index  types.U32
	Error  types.DispatchError
	Topics []types.Hash
}

type EventUtilityBatchCompleted struct {
	Phase types.Phase

	Topics []types.Hash
}

type EventUtilityItemCompleted struct {
	Phase types.Phase

	Topics []types.Hash
}

type EventStakingChilled struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventSystemRemarked struct {
	Phase  types.Phase
	Who    types.AccountID
	Hash   types.Hash
	Topics []types.Hash
}

type EventProxyProxyAdded struct {
	Phase     types.Phase
	WhoA      types.AccountID
	WhoB      types.AccountID
	ProxyType types.ProxyType
	Number    types.BlockNumber
	Topics    []types.Hash
}

type EventStakingPayoutStarted struct {
	Phase    types.Phase
	EraIndex types.U32
	Who      types.AccountID
	Topics   []types.Hash
}
