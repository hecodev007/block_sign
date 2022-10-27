package bifrost

import "github.com/centrifuge/go-substrate-rpc-client/v3/types"

//CollatorSelection

type EventCollatorSelectionNewInvulnerables struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventCollatorSelectionNewDesiredCandidates struct {
	Phase  types.Phase
	Count  types.U32
	Topics []types.Hash
}

type EventCollatorSelectionNewCandidacyBond struct {
	Phase   types.Phase
	Balance types.U128
	Topics  []types.Hash
}

type EventCollatorSelectionCandidateAdded struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventCollatorSelectionCandidateRemoved struct {
	Phase  types.Phase
	Who    types.U128
	Topics []types.Hash
}
