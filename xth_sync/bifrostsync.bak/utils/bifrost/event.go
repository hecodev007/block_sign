package bifrost

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

//swork

type EventSworkRegisterSuccess struct {
	Phase         types.Phase
	AccountID     types.AccountID
	SworkerPubKey []byte
	Topics        []types.Hash
}

type EventSworkWorksReportSuccess struct {
	Phase         types.Phase
	Who           types.AccountID
	SworkerPubKey []byte
	Topics        []types.Hash
}

type EventSworkABUpgradeSuccess struct {
	Phase  types.Phase
	Who    types.AccountID
	A      types.Hash
	B      types.Hash
	Topics []types.Hash
}

type EventSworkSetCodeSuccess struct {
	Phase       types.Phase
	SworkerCode types.U32
	BlockNumber types.U64
	Topics      []types.Hash
}

type EventSworkJoinGroupSuccess struct {
	Phase  types.Phase
	Member types.AccountID
	Owner  types.AccountID
	Topics []types.Hash
}

type EventSworkQuitGroupSuccess struct {
	Phase  types.Phase
	Member types.AccountID
	Ower   types.AccountID
	Topics []types.Hash
}

type EventSworkCreateGroupSuccess struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventSworkKickOutSuccess struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventSworkCancelPunishmentSuccess struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventSworkAddIntoAllowlistSuccess struct {
	Phase  types.Phase
	WhoA   types.AccountID
	WhoB   types.AccountID
	Topics []types.Hash
}

type EventSworkRemoveFromAllowlistSuccesss struct {
	Phase  types.Phase
	WhoA   types.AccountID
	WhoB   types.AccountID
	Topics []types.Hash
}

type EventSworkSetPunishmentSuccess struct {
	Phase  types.Phase
	Enable types.Bool
	Topics []types.Hash
}

type EventSworkRemoveCodeSuccess struct {
	Phase       types.Phase
	SworkerCode types.U32
	Topics      []types.Hash
}

//Staking
type EventStakingReward struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventStakingSlash struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventStakingOldSlashingReportDiscarded struct {
	Phase        types.Phase
	SessionIndex types.U32
	Topics       []types.Hash
}

type EventStakingEraReward struct {
	Phase    types.Phase
	EraIndex types.U32
	ABalance types.U128
	BBalance types.U128
	Topics   []types.Hash
}
type EventStakingNotEnoughCurrency struct {
	Phase    types.Phase
	EraIndex types.U32
	ABalance types.U128
	BBalance types.U128
	Topics   []types.Hash
}

type EventStakingBonded struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventStakingUnbonded struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventStakingWithdrawn struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventStakingValidateSuccess struct {
	Phase types.Phase
	Who   types.AccountID
	ValidatorPrefs
	Topics []types.Hash
}

type ValidatorPrefs struct {
	/// Reward that validator takes up-front; only the rest is split between themselves and
	/// guarantors.

	Fee types.U32
}

type EventStakingGuaranteeSuccess struct {
	Phase   types.Phase
	A       types.AccountID
	B       types.AccountID
	Balance types.U128
	Topics  []types.Hash
}
type EventStakingCutGuaranteeSuccess struct {
	Phase   types.Phase
	A       types.AccountID
	B       types.AccountID
	Balance types.U128
	Topics  []types.Hash
}
type EventStakingChillSuccess struct {
	Phase  types.Phase
	A      types.AccountID
	B      types.AccountID
	Topics []types.Hash
}
type EventStakingUpdateStakeLimitSuccess struct {
	Phase  types.Phase
	Limit  types.U32
	Topics []types.Hash
}

//Market

type EventMarketFileSuccess struct {
	Phase      types.Phase
	Who        types.AccountID
	MerkleRoot types.Bytes
	Topics     []types.Hash
}

type EventMarketRenewFileSuccess struct {
	Phase      types.Phase
	Who        types.AccountID
	MerkleRoot types.Bytes
	Topics     []types.Hash
}
type EventMarketAddPrepaidSuccess struct {
	Phase      types.Phase
	Who        types.AccountID
	MerkleRoot types.Bytes
	Balance    types.U128
	Topics     []types.Hash
}

type EventMarketCalculateSuccess struct {
	Phase      types.Phase
	MerkleRoot types.Bytes
	Topics     []types.Hash
}

type EventMarketIllegalFileClosed struct {
	Phase      types.Phase
	MerkleRoot types.Bytes
	Topics     []types.Hash
}

type EventMarketRewardMerchantSuccess struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventMarketSetEnableMarketSuccess struct {
	Phase  types.Phase
	Enable types.Bool
	Topics []types.Hash
}

type EventMarketSetBaseFeeSuccess struct {
	Phase   types.Phase
	Balance types.U128
	Topics  []types.Hash
}

//Benefits
type FundsType struct {
	IsSwork  bool
	AsSwork  uint
	IsMarket bool
	AsMarket uint
}

type EventBenefitsAddBenefitFundsSuccess struct {
	Phase     types.Phase
	Who       types.AccountID
	Balance   types.U128
	FundsType FundsType
	Topics    []types.Hash
}

type EventBenefitsCutBenefitFundsSuccess struct {
	Phase     types.Phase
	Who       types.AccountID
	Balance   types.U128
	FundsType FundsType
	Topics    []types.Hash
}

type EventBenefitsRebondBenefitFundsSuccess struct {
	Phase     types.Phase
	Who       types.AccountID
	Balance   types.U128
	FundsType FundsType
	Topics    []types.Hash
}

type EventBenefitsWithdrawBenefitFundsSuccess struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

//ChainBridge
type EventChainBridgeRelayerThresholdChanged struct {
	Phase  types.Phase
	A      types.U32
	Topics []types.Hash
}

type EventChainBridgeChainWhitelisted struct {
	Phase  types.Phase
	A      types.U8
	Topics []types.Hash
}
type EventChainBridgeRelayerAdded struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}
type EventChainBridgeRelayerRemoved struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}
type EventChainBridgeFungibleTransfer struct {
	Phase      types.Phase
	A          types.U8
	B          types.U64
	ResourceId [32]types.U8
	C          types.U256
	D          []byte
	Topics     []types.Hash
}
type EventChainBridgeNonFungibleTransfer struct {
	Phase      types.Phase
	A          types.U8
	B          types.U64
	ResourceId [32]types.U8
	C          types.Bytes
	D          types.Bytes
	E          types.Bytes
	Topics     []types.Hash
}

type EventChainBridgeGenericTransfer struct {
	Phase      types.Phase
	A          types.U8
	B          types.U64
	ResourceId [32]types.U8
	C          types.Bytes
	Topics     []types.Hash
}
type EventChainBridgeVoteFor struct {
	Phase  types.Phase
	A      types.U8
	B      types.U64
	Who    types.AccountID
	Topics []types.Hash
}
type EventChainBridgeVoteAgainst struct {
	Phase  types.Phase
	A      types.U8
	B      types.U64
	Who    types.AccountID
	Topics []types.Hash
}

type EventChainBridgeProposalApproved struct {
	Phase  types.Phase
	A      types.U8
	B      types.U64
	Topics []types.Hash
}

type EventChainBridgeProposalRejected struct {
	Phase  types.Phase
	A      types.U8
	B      types.U64
	Topics []types.Hash
}

type EventChainBridgeProposalSucceeded struct {
	Phase  types.Phase
	A      types.U8
	B      types.U64
	Topics []types.Hash
}

type EventChainBridgeProposalFailed struct {
	Phase  types.Phase
	A      types.U8
	B      types.U64
	Topics []types.Hash
}

//BridgeTransfer

type EventBridgeTransferFeeUpdated struct {
	Phase    types.Phase
	ChainID  types.U8
	MinFee   types.U128
	FeeScale types.U32
	Topics   []types.Hash
}

//Claims

type EventClaimsInitPot struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventClaimsSuperiorChanged struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventClaimsMinerChanged struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}
type EventClaimsSetLimitSuccess struct {
	Phase  types.Phase
	Limit  types.U128
	Topics []types.Hash
}

//
type EthereumTxHash struct {
	A [32]types.U8
}

type EthereumAddress struct {
	A [20]types.U8
}

type EventClaimsMintSuccess struct {
	Phase           types.Phase
	EthereumTxHash  [32]types.U8
	EthereumAddress [20]types.U8
	Balance         types.U128
	Topics          []types.Hash
}

type EventClaimsClaimed struct {
	Phase           types.Phase
	Who             types.AccountID
	EthereumAddress [20]types.U8
	Amount          types.U128
	Topics          []types.Hash
}

type EventClaimsBondEthSuccess struct {
	Phase           types.Phase
	Who             types.AccountID
	EthereumAddress [20]types.U8
	Topics          []types.Hash
}

//locks
type EventLocksUnlockStartedFrom struct {
	Phase       types.Phase
	BlockNumber types.U64
	Topics      []types.Hash
}

type EventLocksUnlockSuccess struct {
	Phase       types.Phase
	Who         types.AccountID
	BlockNumber types.U64
	Topics      []types.Hash
}

// bifrost
type EventVoucherDestroyedVoucher struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}
type EventVoucherIssuedVoucher struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}
type EventSwapCreatePoolSuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventSwapSwapTokenSuccess struct {
	Phase    types.Phase
	Balance1 types.U128
	Balance2 types.U128
	Topics   []types.Hash
}
type EventSwapRemoveSingleLiquiditySuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventSwapAddSingleLiquiditySuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventSwapRemoveLiquiditySuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventSwapAddLiquiditySuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeIostRemovedCrossChainPrivilege struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}
type EventBridgeIostGrantedCrossChainPrivilege struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}
type EventBridgeIostSendTransactionFailure struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeIostSendTransactionSuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeIostWithdrawFail struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeIostWithdraw struct {
	Phase  types.Phase
	Who    types.AccountID
	Data   types.Bytes
	Topics []types.Hash
}
type EventBridgeIostDepositFail struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeIostDeposit struct {
	Phase  types.Phase
	Data   types.Bytes
	Who    types.AccountID
	Topics []types.Hash
}

type EventBridgeIostRelayBlock struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeIostProveAction struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeIostChangeSchedule struct {
	Phase      types.Phase
	VersionId1 VersionId
	VersionId2 VersionId
	Topics     []types.Hash
}
type EventBridgeIostInitSchedule struct {
	Phase     types.Phase
	VersionId VersionId
	Topics    []types.Hash
}
type EventBridgeEosUnsignedTrx struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeEosRemovedCrossChainPrivilege struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}
type EventBridgeEosGrantedCrossChainPrivilege struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}
type EventBridgeEosFailToSendCrossChainTransaction struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeEosSentCrossChainTransaction struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeEosWithdrawFail struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeEosWithdraw struct {
	Phase  types.Phase
	Who    types.AccountID
	Data   types.Bytes
	Topics []types.Hash
}
type EventBridgeEosDepositFail struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeEosDeposit struct {
	Phase  types.Phase
	Data   types.Bytes
	Who    types.AccountID
	Topics []types.Hash
}
type EventBridgeEosRelayBlock struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeEosProveAction struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventBridgeEosChangeSchedule struct {
	Phase      types.Phase
	VersionId1 VersionId
	VersionId2 VersionId
	Topics     []types.Hash
}
type EventBridgeEosInitSchedule struct {
	Phase     types.Phase
	VersionId VersionId
	Topics    []types.Hash
}
type EventConvertUpdateConvertPoolSuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventConvertRedeemedPointsSuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventConvertConvertVTokenToTokenSuccess struct {
	Phase types.Phase

	Topics []types.Hash
}

type EventConvertConvertTokenToVTokenSuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventConvertUpdateRatePerBlockSuccess struct {
	Phase types.Phase

	Topics []types.Hash
}
type EventConvertUpdateConvertSuccess struct {
	Phase  types.Phase
	Topics []types.Hash
}
type EventAssetsUnlockedAsset struct {
	Phase       types.Phase
	Who         types.AccountID
	TokenSymbol TokenSymbol
	Balance     types.U128
	Topics      []types.Hash
}
type EventAssetsAccountAssetDestroy struct {
	Phase   types.Phase
	Who     types.AccountID
	AssetId AssetId
	Topics  []types.Hash
}
type EventAssetsAccountAssetCreated struct {
	Phase   types.Phase
	Who     types.AccountID
	AssetId AssetId
	Topics  []types.Hash
}
type EventAssetsCreated struct {
	Phase        types.Phase
	AssetId      AssetId
	TokenBalance types.U128
	Topics       []types.Hash
}
type EventElectionsSeatHolderSlashed struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}
type EventElectionsCandidateSlashed struct {
	Phase   types.Phase
	Who     types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type EventDemocracyBlacklisted struct {
	Phase  types.Phase
	Hash   types.Hash
	Topics []types.Hash
}
type EventElectionsElectionError struct {
	Phase types.Phase

	Topics []types.Hash
}

/*
Bifrost types: https://github.com/bifrost-finance/bifrost/blob/multiaddress-parachain-doc/docs/developer_setting.json
*/

type VersionId types.U32

type TokenSymbol struct {
	enum  []string
	Value string
}

func (d *TokenSymbol) Decode(decoder scale.Decoder) error {
	d.enum = []string{"aUSD", "DOT", "vDOT", "KSM", "vKSM", "EOS", "vEOS", "IOST", "vIOST"}
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}
	if int(b) > len(d.enum) || int(b) < 0 {
		return fmt.Errorf("types=[TokenSymbol]  don not have this enum: %d", b)
	}
	d.Value = d.enum[int(b)]
	return nil
}

type AssetId types.U32

type Balance types.U128
