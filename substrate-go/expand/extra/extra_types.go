package extra

import gsrpcTypes "github.com/centrifuge/go-substrate-rpc-client/v3/types"

///centrifuge/v3/types/events.go expand from https://github.com/Phala-Network/go-substrate-rpc-client/blob/master/types/events.go

// EventElectionsMemberKicked is emitted when a member has been removed.
// This should always be followed by either `NewTerm` or `EmptyTerm`.
type EventElectionsMemberKicked struct {
	Phase  gsrpcTypes.Phase
	Member gsrpcTypes.AccountID
	Topics []gsrpcTypes.Hash
}

// EventElectionsMemberRenounced is emitted when a member has renounced their candidacy.
type EventElectionsRenounced struct {
	Phase  gsrpcTypes.Phase
	Member gsrpcTypes.AccountID
	Topics []gsrpcTypes.Hash
}

type EventElectionsCandidateSlashed struct {
	Phase  gsrpcTypes.Phase
	Member gsrpcTypes.AccountID
	Amount gsrpcTypes.U128
	Topics []gsrpcTypes.Hash
}

type EventElectionsSeatHolderSlashed struct {
	Phase  gsrpcTypes.Phase
	Member gsrpcTypes.AccountID
	Amount gsrpcTypes.U128
	Topics []gsrpcTypes.Hash
}

/*
	RandomnessCollectiveFlip: pallet_randomness_collective_flip::{Pallet, Storage} = 32

	Proposed(AccountId, ProposalIndex, Hash, MemberCount),
	Voted(AccountId, Hash, bool, MemberCount, MemberCount),
	Approved(Hash),
	Disapproved(Hash),
	Executed(Hash, DispatchResult),
	MemberExecuted(Hash, DispatchResult),
	Closed(Hash, MemberCount, MemberCount),
*/

type Proposed struct {
	Phase       gsrpcTypes.Phase
	Descriptor  CandidateDescriptor
	AccountId   gsrpcTypes.AccountID
	Index       gsrpcTypes.U32
	ProposeHash gsrpcTypes.Hash
	MemberCount gsrpcTypes.U32
	Hash        gsrpcTypes.Hash
	Topics      []gsrpcTypes.Hash
}

type Voted struct {
	Phase       gsrpcTypes.Phase
	Descriptor  CandidateDescriptor
	AccountId   gsrpcTypes.AccountID
	ProposeHash gsrpcTypes.Hash
	IsVoted     gsrpcTypes.Bool
	Yes         gsrpcTypes.U32
	No          gsrpcTypes.U32
	Hash        gsrpcTypes.Hash
	Topics      []gsrpcTypes.Hash
}

type Approved struct {
	Phase       gsrpcTypes.Phase
	Descriptor  CandidateDescriptor
	ProposeHash gsrpcTypes.Hash
	Hash        gsrpcTypes.Hash
	Topics      []gsrpcTypes.Hash
}

type Disapproved struct {
	Phase       gsrpcTypes.Phase
	Descriptor  CandidateDescriptor
	ProposeHash gsrpcTypes.Hash
	Hash        gsrpcTypes.Hash
	Topics      []gsrpcTypes.Hash
}

type Executed struct {
	Phase       gsrpcTypes.Phase
	Descriptor  CandidateDescriptor
	ProposeHash gsrpcTypes.Hash
	Result      DispatchResult
	Hash        gsrpcTypes.Hash
	Topics      []gsrpcTypes.Hash
}

type MemberExecuted struct {
	Phase       gsrpcTypes.Phase
	Descriptor  CandidateDescriptor
	ProposeHash gsrpcTypes.Hash
	Result      DispatchResult
	Hash        gsrpcTypes.Hash
	Topics      []gsrpcTypes.Hash
}

type Closed struct {
	Phase       gsrpcTypes.Phase
	Descriptor  CandidateDescriptor
	ProposeHash gsrpcTypes.Hash
	Yes         gsrpcTypes.U32
	No          gsrpcTypes.U32
	Hash        gsrpcTypes.Hash
	Topics      []gsrpcTypes.Hash
}

/*
	Gilt: pallet_gilt::{Pallet, Call, Storage, Event<T>, Config} = 38

	BidPlaced(T::AccountId, BalanceOf<T>, u32),
	BidRetracted(T::AccountId, BalanceOf<T>, u32),
	GiltIssued(ActiveIndex, T::BlockNumber, T::AccountId, BalanceOf<T>),
	GiltThawed(ActiveIndex, T::AccountId, BalanceOf<T>, BalanceOf<T>),
*/

type BidPlaced struct {
	Phase    gsrpcTypes.Phase
	Who      gsrpcTypes.AccountID
	Amount   gsrpcTypes.U128
	Duration gsrpcTypes.U32
	Topics   []gsrpcTypes.Hash
}

type BidRetracted struct {
	Phase    gsrpcTypes.Phase
	Who      gsrpcTypes.AccountID
	Amount   gsrpcTypes.U128
	Duration gsrpcTypes.U32
	Topics   []gsrpcTypes.Hash
}

type GiltIssued struct {
	Phase  gsrpcTypes.Phase
	Index  gsrpcTypes.U32
	Expiry gsrpcTypes.U32
	Who    gsrpcTypes.AccountID
	Amount gsrpcTypes.U128
	Topics []gsrpcTypes.Hash
}

type GiltThawed struct {
	Phase      gsrpcTypes.Phase
	Index      gsrpcTypes.U32
	Who        gsrpcTypes.AccountID
	Original   gsrpcTypes.U128
	Additional gsrpcTypes.U128
	Topics     []gsrpcTypes.Hash
}

/*
    ParasInclusion: parachains_inclusion::{Pallet, Call, Storage, Event<T>} = 53

	/// A candidate was backed. [candidate, head_data]
	CandidateBacked(CandidateReceipt<Hash>, HeadData, CoreIndex, GroupIndex),
	/// A candidate was included. [candidate, head_data]
	CandidateIncluded(CandidateReceipt<Hash>, HeadData, CoreIndex, GroupIndex),
	/// A candidate timed out. [candidate, head_data]
	CandidateTimedOut(CandidateReceipt<Hash>, HeadData, CoreIndex),
*/

type CandidateDescriptor struct {
	ParaId      gsrpcTypes.U32
	RelayParent gsrpcTypes.Hash
	Collator    gsrpcTypes.Hash
	PvdHash     gsrpcTypes.Hash
	PovHash     gsrpcTypes.Hash
	EnsureRoot  gsrpcTypes.Hash
	Signature   gsrpcTypes.Signature
	ParaHead    gsrpcTypes.Hash
	CodeHash    gsrpcTypes.Hash
}

type CandidateReceipt struct {
	Descriptor      CandidateDescriptor
	CommitmentsHash gsrpcTypes.Hash
}

type CandidateIncluded struct {
	Phase      gsrpcTypes.Phase
	Receipt    CandidateReceipt
	Head       gsrpcTypes.Bytes
	CoreIndex  gsrpcTypes.U32
	GroupIndex gsrpcTypes.U32
	Topics     []gsrpcTypes.Hash
}

type CandidateBacked struct {
	Phase      gsrpcTypes.Phase
	Receipt    CandidateReceipt
	Head       gsrpcTypes.Bytes
	CoreIndex  gsrpcTypes.U32
	GroupIndex gsrpcTypes.U32
	Topics     []gsrpcTypes.Hash
}

type CandidateTimedOut struct {
	Phase     gsrpcTypes.Phase
	Receipt   CandidateReceipt
	Head      gsrpcTypes.Bytes
	CoreIndex gsrpcTypes.U32
	Topics    []gsrpcTypes.Hash
}

type Remarked struct {
	Phase   gsrpcTypes.Phase
	Account gsrpcTypes.AccountID
	Hash    gsrpcTypes.Hash
	Topics  []gsrpcTypes.Hash
}

type SchedulerDispatched struct {
	Phase    gsrpcTypes.Phase
	BlockNum gsrpcTypes.U32
	Index    gsrpcTypes.U32
	ID       gsrpcTypes.OptionBytes
	Result   DispatchResult
	Topics   []gsrpcTypes.Hash
}

type SchedulerScheduled struct {
	Phase  gsrpcTypes.Phase
	When   gsrpcTypes.U32
	Index  gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type SchedulerCanceled struct {
	Phase  gsrpcTypes.Phase
	When   gsrpcTypes.U32
	Index  gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

/*
	Paras: parachains_paras::{Pallet, Call, Storage, Event, Config<T>} = 56

	CurrentCodeUpdated(ParaId),
	CurrentHeadUpdated(ParaId),
	CodeUpgradeScheduled(ParaId),
	NewHeadNoted(ParaId),
	ActionQueued(ParaId, SessionIndex),
*/

type CurrentCodeUpdated struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type CurrentHeadUpdated struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type CodeUpgradeScheduled struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type NewHeadNoted struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type ActionQueued struct {
	Phase        gsrpcTypes.Phase
	ParaId       gsrpcTypes.U32
	SessionIndex gsrpcTypes.U32
	Topics       []gsrpcTypes.Hash
}

/*
	ParasUmp: parachains_ump::{Pallet, Call, Storage, Event} = 59

	InvalidFormat(MessageId),
	UnsupportedVersion(MessageId),
	ExecutedUpward(MessageId, Outcome),
	WeightExhausted(MessageId, Weight, Weight),
	UpwardMessagesReceived(ParaId, u32, u32),
*/

type InvalidFormat struct {
	Phase     gsrpcTypes.Phase
	MessageId gsrpcTypes.Bytes32
	Topics    []gsrpcTypes.Hash
}

type UnsupportedVersion struct {
	Phase     gsrpcTypes.Phase
	MessageId gsrpcTypes.Bytes32
	Topics    []gsrpcTypes.Hash
}

type ExecutedUpward struct {
	Phase     gsrpcTypes.Phase
	MessageId gsrpcTypes.Bytes32
	OutCome   OutCome
	Topics    []gsrpcTypes.Hash
}

type WeightExhausted struct {
	Phase     gsrpcTypes.Phase
	MessageId gsrpcTypes.Bytes32
	Remaining gsrpcTypes.U64
	Required  gsrpcTypes.U64
	Topics    []gsrpcTypes.Hash
}

type UpwardMessagesReceived struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Count  gsrpcTypes.U32
	Size   gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

/*
	ParasHrmp: parachains_hrmp::{Pallet, Call, Storage, Event} = 60

	OpenChannelRequested(ParaId, ParaId, u32, u32),
	OpenChannelAccepted(ParaId, ParaId),
	ChannelClosed(ParaId, HrmpChannelId),
*/

type OpenChannelRequested struct {
	Phase        gsrpcTypes.Phase
	Sender       gsrpcTypes.U32
	Recipient    gsrpcTypes.U32
	Capacity     gsrpcTypes.U32
	MaxSize      gsrpcTypes.U32
	SessionIndex gsrpcTypes.U32
	Topics       []gsrpcTypes.Hash
}

type OpenChannelAccepted struct {
	Phase     gsrpcTypes.Phase
	Sender    gsrpcTypes.U32
	Recipient gsrpcTypes.U32
	Topics    []gsrpcTypes.Hash
}

type ChannelClosed struct {
	Phase     gsrpcTypes.Phase
	Chain     gsrpcTypes.U32
	Sender    gsrpcTypes.U32
	Recipient gsrpcTypes.U32
	Topics    []gsrpcTypes.Hash
}

/*
   	Registrar: paras_registrar::{Pallet, Call, Storage, Event<T>} = 70

	Registered(ParaId, AccountId),
	Deregistered(ParaId),
	Reserved(ParaId, AccountId),
*/

type Registered struct {
	Phase   gsrpcTypes.Phase
	ParaId  gsrpcTypes.U32
	Account gsrpcTypes.AccountID
	Topics  []gsrpcTypes.Hash
}

type Deregistered struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type ParaReserved struct {
	Phase   gsrpcTypes.Phase
	ParaId  gsrpcTypes.U32
	Account gsrpcTypes.AccountID
	Topics  []gsrpcTypes.Hash
}

/*
	XcmPallet: pallet_xcm::{Pallet, Call, Storage, Event<T>} = 99

	Attempted(xcm::v0::Outcome),
	Sent(MultiLocation, MultiLocation, Xcm<()>),
*/
type Attempted struct {
	Phase   gsrpcTypes.Phase
	OutCome OutCome
	Topics  []gsrpcTypes.Hash
}

type Sent struct {
	Phase  gsrpcTypes.Phase
	Origin MultiLocation
	Target MultiLocation
	Xcm    Xcm
	Topics []gsrpcTypes.Hash
}

type EventSystemExtrinsicFailed struct {
	Phase         gsrpcTypes.Phase
	DispatchError DispatchError
	DispatchInfo  gsrpcTypes.DispatchInfo
	Topics        []gsrpcTypes.Hash
}

type ProxyProxyExecuted struct {
	Phase         gsrpcTypes.Phase
	DispatchError DispatchResult
	Topics        []gsrpcTypes.Hash
}

// EventUtilityItemCompleted is emitted when a item of batch dispatch completed with no error.
type EventUtilityItemCompleted struct {
	Phase  gsrpcTypes.Phase
	Topics []gsrpcTypes.Hash
}
