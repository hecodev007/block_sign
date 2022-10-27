package xgateway

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

/// XGatewayRecords Type
type XGatewayRecords struct {
	XGatewayRecords_Deposited           []EventXGatewayRecordsDeposited
	XGatewayRecords_WithdrawalCreated   []EventXGatewayRecordsWithdrawalCreated
	XGatewayRecords_WithdrawalProcessed []EventXGatewayRecordsWithdrawalProcessed
	XGatewayRecords_WithdrawalRecovered []EventXGatewayRecordsWithdrawalRecovered
	XGatewayRecords_WithdrawalCanceled  []EventXGatewayRecordsWithdrawalCanceled
	XGatewayRecords_WithdrawalFinished  []EventXGatewayRecordsWithdrawalFinished
}

type WithdrawalRecordId uint32
type AddrStr []uint8
type Memo struct {
	Memo []uint8
}

type OptionU128 struct {
	HasValue bool
	Balance  types.U128
}

type WithdrawalRecord struct {
	AssetId   types.U32
	Applicant types.AccountID
	Balance   types.U128
	Addr      AddrStr
	Ext       Memo
	Height    types.BlockNumber
}

type WithdrawalState uint8

const (
	Applying     WithdrawalState = 0
	Processing   WithdrawalState = 1
	NormalFinish WithdrawalState = 2
	RootFinish   WithdrawalState = 3
	NormalCancel WithdrawalState = 4
	RootCancel   WithdrawalState = 5
)

/// An account deposited some asset. [who, asset_id, amount]
type EventXGatewayRecordsDeposited struct {
	Phase   types.Phase
	Who     types.AccountID
	AssetId types.AccountID
	Amount  Balance
	Topics  []types.Hash
}

/// A withdrawal application was created. [withdrawal_id, record_info]
type EventXGatewayRecordsWithdrawalCreated struct {
	Phase        types.Phase
	WithdrawalId WithdrawalRecordId
	RecordInfo   WithdrawalRecord
	Topics       []types.Hash
}

/// A withdrawal proposal was processed. [withdrawal_id]
type EventXGatewayRecordsWithdrawalProcessed struct {
	Phase        types.Phase
	WithdrawalId WithdrawalRecordId
	Topics       []types.Hash
}

/// A withdrawal proposal was recovered. [withdrawal_id]
type EventXGatewayRecordsWithdrawalRecovered struct {
	Phase        types.Phase
	WithdrawalId WithdrawalRecordId
	Topics       []types.Hash
}

/// A withdrawal proposal was canceled. [withdrawal_id, withdrawal_state]
type EventXGatewayRecordsWithdrawalCanceled struct {
	Phase           types.Phase
	WithdrawalId    WithdrawalRecordId
	WithdrawalState WithdrawalState
	Topics          []types.Hash
}

/// A withdrawal proposal was canceled. [withdrawal_id, withdrawal_state]
type EventXGatewayRecordsWithdrawalFinished struct {
	Phase           types.Phase
	WithdrawalId    WithdrawalRecordId
	WithdrawalState WithdrawalState
	Topics          []types.Hash
}
