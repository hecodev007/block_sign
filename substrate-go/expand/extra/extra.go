package extra

import gsrpcTypes "github.com/centrifuge/go-substrate-rpc-client/v3/types"

type ExtraEvents struct {
	ExtraEventRecord
	//ElectionsPhragmen
}

type ExtraEventRecord struct {
	// Fix some events
	Balances_ReserveRepatriated []gsrpcTypes.EventBalancesReserveRepatriated
	System_Remarked             []Remarked
	Scheduler_Dispatched        []SchedulerDispatched
	Scheduler_Scheduled         []SchedulerScheduled
	Scheduler_Canceled          []SchedulerCanceled
	System_ExtrinsicFailed      []EventSystemExtrinsicFailed
	Proxy_ProxyExecuted         []ProxyProxyExecuted

	// slot auction
	Crowdloan_Created           []Created
	Crowdloan_Contributed       []Contributed
	Crowdloan_Withdrew          []Withdrew
	Crowdloan_PartiallyRefunded []PartiallyRefunded
	Crowdloan_AllRefunded       []PartiallyRefunded
	Crowdloan_Dissolved         []Dissolved
	Crowdloan_HandleBidResult   []HandleBidResult
	Crowdloan_Edited            []Edited
	Crowdloan_MemoUpdated       []MemoUpdated
	Crowdloan_AddedToNewRaise   []AddedToNewRaise

	Auctions_AuctionStarted     []AuctionStarted
	Auctions_AuctionClosed      []AuctionClosed
	Auctions_Reserved           []Reserved
	Auctions_Unreserved         []Unreserved
	Auctions_ReserveConfiscated []ReserveConfiscated
	Auctions_BidAccepted        []BidAccepted
	Auctions_WinningOffset      []WinningOffset

	Slots_NewLeasePeriod []NewLeasePeriod
	Slots_Leased         []Leased

	// rococo
	Inclusion_CandidateIncluded []CandidateIncluded
	Inclusion_CandidateBacked   []CandidateBacked
	Inclusion_CandidateTimedOut []CandidateTimedOut

	// kusama
	ParaInclusion_CandidateIncluded []CandidateIncluded
	ParaInclusion_CandidateBacked   []CandidateBacked
	ParaInclusion_CandidateTimedOut []CandidateTimedOut

	RandomnessCollectiveFlip_Proposed       []Proposed
	RandomnessCollectiveFlip_Voted          []Voted
	RandomnessCollectiveFlip_Approved       []Approved
	RandomnessCollectiveFlip_Disapproved    []Disapproved
	RandomnessCollectiveFlip_Executed       []Executed
	RandomnessCollectiveFlip_MemberExecuted []MemberExecuted
	RandomnessCollectiveFlip_Closed         []Closed

	PhragmenElection_NewTerm           []gsrpcTypes.EventElectionsNewTerm
	PhragmenElection_EmptyTerm         []gsrpcTypes.EventElectionsEmptyTerm
	PhragmenElection_ElectionError     []gsrpcTypes.EventElectionsElectionError
	PhragmenElection_MemberKicked      []gsrpcTypes.EventElectionsMemberKicked
	PhragmenElection_Renounced         []gsrpcTypes.EventElectionsRenounced
	PhragmenElection_CandidateSlashed  []gsrpcTypes.EventElectionsCandidateSlashed
	PhragmenElection_SeatHolderSlashed []gsrpcTypes.EventElectionsSeatHolderSlashed

	Gilt_BidPlaced    []BidPlaced
	Gilt_BidRetracted []BidRetracted
	Gilt_GiltIssued   []GiltIssued
	Gilt_GiltThawed   []GiltThawed

	XcmPallet_Attempted []Attempted
	XcmPallet_Sent      []Sent

	Paras_CurrentCodeUpdated   []CurrentCodeUpdated
	Paras_CurrentHeadUpdated   []CurrentHeadUpdated
	Paras_CodeUpgradeScheduled []CodeUpgradeScheduled
	Paras_NewHeadNoted         []NewHeadNoted
	Paras_ActionQueued         []ActionQueued

	ParasUmp_InvalidFormat          []InvalidFormat
	ParasUmp_UnsupportedVersion     []UnsupportedVersion
	ParasUmp_ExecutedUpward         []ExecutedUpward
	ParasUmp_WeightExhausted        []WeightExhausted
	ParasUmp_UpwardMessagesReceived []UpwardMessagesReceived

	ParasHrmp_OpenChannelRequested []OpenChannelRequested
	ParasHrmp_OpenChannelAccepted  []OpenChannelAccepted
	ParasHrmp_ChannelClosed        []ChannelClosed

	Registrar_Registered   []Registered
	Registrar_Deregistered []Deregistered
	Registrar_Reserved     []ParaReserved

	// Update Origin
	Utility_ItemCompleted []EventUtilityItemCompleted //nolint:stylecheck,golint
}
