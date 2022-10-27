package chainx

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/rjman-ljm/substrate-go/expand/base"
	"github.com/rjman-ljm/substrate-go/expand/bridge"
	"github.com/rjman-ljm/substrate-go/expand/chainx/pallets"
	"github.com/rjman-ljm/substrate-go/expand/extra"
)

type ChainXEventRecords struct {
	//types.EventRecords
	base.BaseEventRecords
	XPallets
	pallets.Swap
	bridge.BridgeEvents
}

func (p ChainXEventRecords) GetMultisigNewMultisig() []types.EventMultisigNewMultisig {
	return p.Multisig_NewMultisig
}

func (p ChainXEventRecords) GetMultisigApproval() []types.EventMultisigApproval {
	return p.Multisig_MultisigApproval
}

func (p ChainXEventRecords) GetMultisigExecuted() []types.EventMultisigExecuted {
	return p.Multisig_MultisigExecuted
}

func (p ChainXEventRecords) GetMultisigCancelled() []types.EventMultisigCancelled {
	return p.Multisig_MultisigCancelled
}

func (p ChainXEventRecords) GetUtilityBatchCompleted() []types.EventUtilityBatchCompleted {
	return p.Utility_BatchCompleted
}

func (p ChainXEventRecords) GetBalancesTransfer() []types.EventBalancesTransfer {
	return p.Balances_Transfer
}

func (p ChainXEventRecords) GetSystemExtrinsicSuccess() []types.EventSystemExtrinsicSuccess {
	return p.System_ExtrinsicSuccess
}

func (p ChainXEventRecords) GetSystemExtrinsicFailed() []extra.EventSystemExtrinsicFailed {
	return p.System_ExtrinsicFailed
}
