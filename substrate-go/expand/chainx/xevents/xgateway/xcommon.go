package xgateway

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

/// XGatewayCommon Type
type XGatewayCommon struct {
	XGatewayCommon_SetTrusteeProps   []EventXGatewayCommonSetTrusteeProps
	XGatewayCommon_ReferralBinded    []EventXGatewayCommonReferralBinded
	XGatewayCommon_TrusteeSetChanged []EventXGatewayCommonTrusteeSetChanged
}

type Chain uint8

const (
	ChainX   Chain = 0
	Bitcoin  Chain = 1
	Ethereum Chain = 2
	Polkadot Chain = 3
)

type GenericTrusteeIntentionProps struct {
	TrusteeIntentionProps
}

type TrusteeIntentionProps struct {
	About      []types.U8
	HotEntity  []types.U8
	ColdEntity []types.U8
}

type GenericTrusteeSessionInfo struct {
	TrusteeSessionInfo
}

type TrusteeSessionInfo struct {
	TrusteeList []types.AccountID
	Threshold   types.U16
	HotAddress  []types.U8
	ColdAddress []types.U8
}

/// A (potential) trustee set the required properties. [who, chain, trustee_props]
type EventXGatewayCommonSetTrusteeProps struct {
	Phase        types.Phase
	Who          types.AccountID
	Chain        Chain
	TrusteeProps GenericTrusteeIntentionProps
	Topics       []types.Hash
}

/// An account set its referral_account of some chain. [who, chain, referral_account]
type EventXGatewayCommonReferralBinded struct {
	Phase           types.Phase
	Who             types.AccountID
	Chain           Chain
	ReferralAccount types.AccountID
	Topics          []types.Hash
}

/// The trustee set of a chain was changed. [chain, session_number, session_info]
type EventXGatewayCommonTrusteeSetChanged struct {
	Phase         types.Phase
	Chain         Chain
	SessionNumber types.U32
	SessionInfo   GenericTrusteeSessionInfo
	Topics        []types.Hash
}
