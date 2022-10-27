package dot

import (
	//"github.com/stafiprotocol/go-substrate-rpc-client/types"
	"github.com/JFJun/go-substrate-rpc-client/v3/types"
)

type KarEventRecords struct {
	types.EventRecords
	Treasury_DepositRing []types.EventTreasurySpending
	Currencies_Deposited []EventCurrencisDeposited
	//System_ExtrinsicSuccess []EventSystemExtrinsicSuccess //nolint:stylecheck,golint
}
type EventCurrencisDeposited struct {
	Phase   types.Phase
	AssetID types.U32
	From    types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

func (p KarEventRecords) GetBalancesTransfer() []types.EventBalancesTransfer {
	return p.Balances_Transfer
}

func (p KarEventRecords) GetSystemExtrinsicSuccess() []types.EventSystemExtrinsicSuccess {
	//var ret []types.EventSystemExtrinsicSuccess
	//for _, v := range p.System_ExtrinsicSuccess {
	//	tmpv := types.EventSystemExtrinsicSuccess{}
	//	tmpv.DispatchInfo = v.DispatchInfo
	//	ret = append(ret, tmpv)
	//}
	//return ret
	return p.System_ExtrinsicSuccess
}

func (p KarEventRecords) GetSystemExtrinsicFailed() []types.EventSystemExtrinsicFailed {
	return p.System_ExtrinsicFailed
}
