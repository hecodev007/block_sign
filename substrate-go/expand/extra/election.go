package extra

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

type ElectionsPhragmen struct {
	Claims_Claimed                    	[]EventClaimsClaimed
	ElectionsPhragmen_VoterReported   	[]EventElectionsPhragmenVoterReported
	ElectionsPhragmen_MemberRenounced 	[]EventElectionsPhragmenMemberRenounced
	ElectionsPhragmen_MemberKicked    	[]EventElectionsPhragmenMemberKicked
	ElectionsPhragmen_ElectionError   	[]EventElectionsPhragmenElectionError
	ElectionsPhragmen_EmptyTerm       	[]EventElectionsPhragmenEmptyTerm
	//ElectionsPhragmen_NewTerm			[]EventElectionsPhragmenNewTerm		暂不支持解析
	Democracy_Blacklisted 				[]EventDemocracyBlacklisted
}

type EventDemocracyBlacklisted struct {
	Phase  types.Phase
	Hash   types.Hash
	Topics []types.Hash
}

type EventElectionsPhragmenEmptyTerm struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventElectionsPhragmenElectionError struct {
	Phase  types.Phase
	Topics []types.Hash
}
type EventElectionsPhragmenMemberKicked struct {
	Phase     types.Phase
	AccountId types.AccountID
	Topics    []types.Hash
}
type EventElectionsPhragmenMemberRenounced struct {
	Phase     types.Phase
	AccountId types.AccountID
	Topics    []types.Hash
}
type EventElectionsPhragmenVoterReported struct {
	Phase  types.Phase
	Who1   types.AccountID
	Who2   types.AccountID
	Bool   types.Bool
	Topics []types.Hash
}

//type EventElectionsPhragmenNewTerm struct {
//	Phase    types.Phase
//	Vec
//	Topics []types.Hash
//}

type VecU8L20 struct {
	Value string
}

type EventClaimsClaimed struct {
	Phase           types.Phase
	AccountId       types.AccountID
	EthereumAddress VecU8L20
	Balance         types.U128
	Topics          []types.Hash
}
