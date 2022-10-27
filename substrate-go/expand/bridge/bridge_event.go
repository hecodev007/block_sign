package bridge

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	events "github.com/hacpy/chainbridge-substrate-events"
)

type BridgeEvents struct {
	events.Events
	AssetEvents
}

type AssetEvents struct {
	Erc721_Minted                    []EventErc721Minted                   //nolint:stylecheck,golint
	Erc721_Transferred               []EventErc721Transferred              //nolint:stylecheck,golint
	Erc721_Burned                    []EventErc721Burned                   //nolint:stylecheck,golint
	Example_Remark                   []EventExampleRemark                  //nolint:stylecheck,golint
	Nfts_DepositAsset                []EventNFTDeposited                   //nolint:stylecheck,golint
	Council_Proposed                 []types.EventCollectiveProposed       //nolint:stylecheck,golint
	Council_Voted                    []types.EventCollectiveVoted          //nolint:stylecheck,golint
	Council_Approved                 []types.EventCollectiveApproved       //nolint:stylecheck,golint
	Council_Disapproved              []types.EventCollectiveDisapproved    //nolint:stylecheck,golint
	Council_Executed                 []types.EventCollectiveExecuted       //nolint:stylecheck,golint
	Council_MemberExecuted           []types.EventCollectiveMemberExecuted //nolint:stylecheck,golint
	Council_Closed                   []types.EventCollectiveClosed         //nolint:stylecheck,golint
	Fees_FeeChanged                  []EventFeeChanged                     //nolint:stylecheck,golint
	MultiAccount_NewMultiAccount     []EventNewMultiAccount                //nolint:stylecheck,golint
	MultiAccount_MultiAccountUpdated []EventMultiAccountUpdated            //nolint:stylecheck,golint
	MultiAccount_MultiAccountRemoved []EventMultiAccountRemoved            //nolint:stylecheck,golint
	MultiAccount_NewMultisig         []EventNewMultisig                    //nolint:stylecheck,golint
	MultiAccount_MultisigApproval    []EventMultisigApproval               //nolint:stylecheck,golint
	MultiAccount_MultisigExecuted    []EventMultisigExecuted               //nolint:stylecheck,golint
	MultiAccount_MultisigCancelled   []EventMultisigCancelled              //nolint:stylecheck,golint
	TreasuryReward_TreasuryMinting   []EventTreasuryMinting                //nolint:stylecheck,golint
	Nft_Transferred                  []EventNftTransferred                 //nolint:stylecheck,golint
	RadClaims_Claimed                []EventRadClaimsClaimed               //nolint:stylecheck,golint
	RadClaims_RootHashStored         []EventRadClaimsRootHashStored        //nolint:stylecheck,golint
	Registry_Mint                    []EventRegistryMint                   //nolint:stylecheck,golint
	Registry_RegistryCreated         []EventRegistryRegistryCreated        //nolint:stylecheck,golint
	Registry_RegistryTmp             []EventRegistryTmp                    //nolint:stylecheck,golint
}

