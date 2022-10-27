package bifrost

import "github.com/centrifuge/go-substrate-rpc-client/v3/types"

//
type EventDemocracyProposed struct {
	Phase     types.Phase
	PropIndex types.U32
	Deposit   Balance
	Topics    []types.Hash
}

type EventDemocracyTabled struct {
	Phase     types.Phase
	PropIndex types.U32
	Deposit   Balance
	Depositor types.AccountID
	Topics    []types.Hash
}

type EventDemocracyExternalTabled struct {
	Phase  types.Phase
	Topics []types.Hash
}

type EventDemocracyStarted struct {
	Phase     types.Phase
	RefIndex  types.U32
	Threshold types.VoteThreshold
	Topics    []types.Hash
}

type EventDemocracyPassed struct {
	Phase    types.Phase
	RefIndex types.U32
	Topics   []types.Hash
}

type EventDemocracyNotPassed struct {
	Phase    types.Phase
	RefIndex types.U32
	Topics   []types.Hash
}

type EventDemocracyCancelled struct {
	Phase    types.Phase
	RefIndex types.U32
	Topics   []types.Hash
}

type EventDemocracyExecuted struct {
	Phase    types.Phase
	RefIndex types.U32
	Result   types.DispatchResult
	Topics   []types.Hash
}
