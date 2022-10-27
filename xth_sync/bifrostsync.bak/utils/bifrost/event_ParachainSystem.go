package bifrost

import "github.com/centrifuge/go-substrate-rpc-client/v3/types"

//ParachainSystem
type EventParachainSystemValidationFunctionStored struct {
	Phase                    types.Phase
	ValidationFunctionStored types.BlockNumber
	Topics                   []types.Hash
}

type EventParachainSystemValidationFunctionApplied struct {
	Phase                    types.Phase
	ValidationFunctionStored types.BlockNumber
	Topics                   []types.Hash
}

type EventParachainSystemUpgradeAuthorized struct {
	Phase  types.Phase
	Hash   types.Hash
	Topics []types.Hash
}

type EventParachainSystemDownwardMessagesReceived struct {
	Phase  types.Phase
	Count  types.U32
	Topics []types.Hash
}

type EventParachainSystemDownwardMessagesProcessed struct {
	Phase  types.Phase
	Weight types.Weight
	Hash   types.Hash
	Topics []types.Hash
}
