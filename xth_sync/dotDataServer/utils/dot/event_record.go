package dot

import "github.com/yanyushr/go-substrate-rpc-client/v3/types"

type DotEventRecords struct {
	types.EventRecords

	Utility_BatchInterrupted []EventUtilityBatchInterrupted
	Utility_BatchCompleted   []EventUtilityBatchCompleted
	Utility_ItemCompleted    []EventUtilityItemCompleted
	Staking_Chilled          []EventStakingChilled
	System_Remarked          []EventSystemRemarked
	Proxy_ProxyAdded         []EventProxyProxyAdded
	Staking_PayoutStarted    []EventStakingPayoutStarted
}
