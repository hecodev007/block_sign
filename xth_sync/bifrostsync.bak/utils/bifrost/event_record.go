package bifrost

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

type CRustEventRecords struct {
	types.EventRecords

	//swork
	Swork_RegisterSuccess             []EventSworkRegisterSuccess
	Swork_WorksReportSuccess          []EventSworkWorksReportSuccess
	Swork_ABUpgradeSuccess            []EventSworkABUpgradeSuccess
	Swork_SetCodeSuccess              []EventSworkSetCodeSuccess
	Swork_JoinGroupSuccess            []EventSworkJoinGroupSuccess
	Swork_QuitGroupSuccess            []EventSworkQuitGroupSuccess
	Swork_CreateGroupSuccess          []EventSworkCreateGroupSuccess
	Swork_KickOutSuccess              []EventSworkKickOutSuccess
	Swork_CancelPunishmentSuccess     []EventSworkCancelPunishmentSuccess
	Swork_AddIntoAllowlistSuccess     []EventSworkAddIntoAllowlistSuccess
	Swork_RemoveFromAllowlistSuccesss []EventSworkRemoveFromAllowlistSuccesss
	Swork_SetPunishmentSuccess        []EventSworkSetPunishmentSuccess
	Swork_RemoveCodeSuccess           []EventSworkRemoveCodeSuccess

	//Staking
	Staking_Reward                     []EventStakingReward
	Staking_Slash                      []EventStakingSlash
	Staking_OldSlashingReportDiscarded []EventStakingOldSlashingReportDiscarded
	Staking_EraReward                  []EventStakingEraReward
	Staking_gNotEnoughCurrency         []EventStakingNotEnoughCurrency
	Staking_Bonded                     []EventStakingBonded
	Staking_Unbonded                   []EventStakingUnbonded
	Staking_Withdrawn                  []EventStakingWithdrawn
	Staking_ValidateSuccess            []EventStakingValidateSuccess
	Staking_GuaranteeSuccess           []EventStakingGuaranteeSuccess
	Staking_CutGuaranteeSuccess        []EventStakingCutGuaranteeSuccess
	Staking_ChillSuccess               []EventStakingChillSuccess
	Staking_UpdateStakeLimitSuccess    []EventStakingUpdateStakeLimitSuccess

	//Market
	Market_FileSuccess            []EventMarketFileSuccess
	Market_RenewFileSuccess       []EventMarketRenewFileSuccess
	Market_AddPrepaidSuccess      []EventMarketAddPrepaidSuccess
	Market_CalculateSuccess       []EventMarketCalculateSuccess
	Market_IllegalFileClosed      []EventMarketIllegalFileClosed
	Market_RewardMerchantSuccess  []EventMarketRewardMerchantSuccess
	Market_SetEnableMarketSuccess []EventMarketSetEnableMarketSuccess
	Market_SetBaseFeeSuccess      []EventMarketSetBaseFeeSuccess

	//Benefits
	Benefits_AddBenefitFundsSuccess      []EventBenefitsAddBenefitFundsSuccess
	Benefits_CutBenefitFundsSuccess      []EventBenefitsCutBenefitFundsSuccess
	Benefits_RebondBenefitFundsSuccess   []EventBenefitsRebondBenefitFundsSuccess
	Benefits_WithdrawBenefitFundsSuccess []EventBenefitsWithdrawBenefitFundsSuccess

	//ChainBridge
	ChainBridge_RelayerThresholdChanged []EventChainBridgeRelayerThresholdChanged
	ChainBridge_ChainWhitelisted        []EventChainBridgeChainWhitelisted
	ChainBridge_RelayerAdded            []EventChainBridgeRelayerAdded
	ChainBridge_RelayerRemoved          []EventChainBridgeRelayerRemoved
	ChainBridge_FungibleTransfer        []EventChainBridgeFungibleTransfer
	ChainBridge_NonFungibleTransfer     []EventChainBridgeNonFungibleTransfer
	ChainBridge_GenericTransfer         []EventChainBridgeGenericTransfer
	ChainBridge_VoteFor                 []EventChainBridgeVoteFor
	ChainBridge_VoteAgainst             []EventChainBridgeVoteAgainst
	ChainBridge_ProposalApproved        []EventChainBridgeProposalApproved
	ChainBridge_roposalRejected         []EventChainBridgeProposalRejected
	ChainBridge_ProposalSucceeded       []EventChainBridgeProposalSucceeded
	ChainBridge_ProposalFailed          []EventChainBridgeProposalFailed

	//BridgeTransfer
	Bridge_TransferFeeUpdated []EventBridgeTransferFeeUpdated

	//Claims
	Claims_InitPot         []EventClaimsInitPot
	Claims_SuperiorChanged []EventClaimsSuperiorChanged
	Claims_MinerChanged    []EventClaimsMinerChanged
	Claims_SetLimitSuccess []EventClaimsSetLimitSuccess
	Claims_MintSuccess     []EventClaimsMintSuccess
	Claims_Claimed         []EventClaimsClaimed
	Claims_BondEthSuccess  []EventClaimsBondEthSuccess

	//Locks
	Locks_UnlockStartedFrom []EventLocksUnlockStartedFrom
	Locks_UnlockSuccess     []EventLocksUnlockSuccess

	//bifrost
	Democracy_Blacklisted                     []EventDemocracyBlacklisted
	Elections_ElectionError                   []EventElectionsElectionError
	Elections_CandidateSlashed                []EventElectionsCandidateSlashed
	Elections_SeatHolderSlashed               []EventElectionsSeatHolderSlashed
	Assets_Created                            []EventAssetsCreated
	Assets_AccountAssetCreated                []EventAssetsAccountAssetCreated
	Assets_AccountAssetDestroy                []EventAssetsAccountAssetDestroy
	Assets_UnlockedAsset                      []EventAssetsUnlockedAsset
	Convert_UpdateConvertSuccess              []EventConvertUpdateConvertSuccess
	Convert_UpdateRatePerBlockSuccess         []EventConvertUpdateRatePerBlockSuccess
	Convert_ConvertTokenToVTokenSuccess       []EventConvertConvertTokenToVTokenSuccess
	Convert_ConvertVTokenToTokenSuccess       []EventConvertConvertVTokenToTokenSuccess
	Convert_RedeemedPointsSuccess             []EventConvertRedeemedPointsSuccess
	Convert_UpdateConvertPoolSuccess          []EventConvertUpdateConvertPoolSuccess
	BridgeEos_InitSchedule                    []EventBridgeEosInitSchedule
	BridgeEos_ChangeSchedule                  []EventBridgeEosChangeSchedule
	BridgeEos_ProveAction                     []EventBridgeEosProveAction
	BridgeEos_RelayBlock                      []EventBridgeEosRelayBlock
	BridgeEos_Deposit                         []EventBridgeEosDeposit
	BridgeEos_DepositFail                     []EventBridgeEosDepositFail
	BridgeEos_Withdraw                        []EventBridgeEosWithdraw
	BridgeEos_WithdrawFail                    []EventBridgeEosWithdrawFail
	BridgeEos_SentCrossChainTransaction       []EventBridgeEosSentCrossChainTransaction
	BridgeEos_FailToSendCrossChainTransaction []EventBridgeEosFailToSendCrossChainTransaction
	BridgeEos_GrantedCrossChainPrivilege      []EventBridgeEosGrantedCrossChainPrivilege
	BridgeEos_RemovedCrossChainPrivilege      []EventBridgeEosRemovedCrossChainPrivilege
	BridgeEos_UnsignedTrx                     []EventBridgeEosUnsignedTrx
	BridgeIost_InitSchedule                   []EventBridgeIostInitSchedule
	BridgeIost_ChangeSchedule                 []EventBridgeIostChangeSchedule
	BridgeIost_ProveAction                    []EventBridgeIostProveAction
	BridgeIost_RelayBlock                     []EventBridgeIostRelayBlock
	BridgeIost_Deposit                        []EventBridgeIostDeposit
	BridgeIost_DepositFail                    []EventBridgeIostDepositFail
	BridgeIost_Withdraw                       []EventBridgeIostWithdraw
	BridgeIost_WithdrawFail                   []EventBridgeIostWithdrawFail
	BridgeIost_SendTransactionSuccess         []EventBridgeIostSendTransactionSuccess
	BridgeIost_SendTransactionFailure         []EventBridgeIostSendTransactionFailure
	BridgeIost_GrantedCrossChainPrivilege     []EventBridgeIostGrantedCrossChainPrivilege
	BridgeIost_RemovedCrossChainPrivilege     []EventBridgeIostRemovedCrossChainPrivilege
	Swap_AddLiquiditySuccess                  []EventSwapAddLiquiditySuccess
	Swap_RemoveLiquiditySuccess               []EventSwapRemoveLiquiditySuccess
	Swap_AddSingleLiquiditySuccess            []EventSwapAddSingleLiquiditySuccess
	Swap_RemoveSingleLiquiditySuccess         []EventSwapRemoveSingleLiquiditySuccess
	Swap_SwapTokenSuccess                     []EventSwapSwapTokenSuccess
	Swap_CreatePoolSuccess                    []EventSwapCreatePoolSuccess
	Voucher_IssuedVoucher                     []EventVoucherIssuedVoucher
	Voucher_DestroyedVoucher                  []EventVoucherDestroyedVoucher
}
