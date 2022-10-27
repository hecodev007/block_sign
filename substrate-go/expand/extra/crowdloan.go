package extra

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
	gsrpcTypes "github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

/// Crowdloan
type Created struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type Contributed struct {
	Phase     gsrpcTypes.Phase
	AccountId gsrpcTypes.AccountID
	ParaId    gsrpcTypes.U32
	Value     gsrpcTypes.U128
	Topics    []gsrpcTypes.Hash
}

type Withdrew struct {
	Phase     gsrpcTypes.Phase
	AccountId gsrpcTypes.AccountID
	ParaId    gsrpcTypes.U32
	Value     gsrpcTypes.U128
	Topics    []gsrpcTypes.Hash
}

type PartiallyRefunded struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type AllRefunded struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type Dissolved struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type HandleBidResult struct {
	Phase          gsrpcTypes.Phase
	ParaId         gsrpcTypes.U32
	DispatchResult DispatchResult
	Topics         []gsrpcTypes.Hash
}

type Edited struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

type MemoUpdated struct {
	Phase     gsrpcTypes.Phase
	AccountId gsrpcTypes.AccountID
	ParaId    gsrpcTypes.U32
	Memo      gsrpcTypes.Bytes
	Topics    []gsrpcTypes.Hash
}

type AddedToNewRaise struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Topics []gsrpcTypes.Hash
}

/// Auction
type AuctionStarted struct {
	Phase        gsrpcTypes.Phase
	AuctionIndex gsrpcTypes.U32
	PeriodBegin  gsrpcTypes.U32
	BlockNumber  gsrpcTypes.U32
	Topics       []gsrpcTypes.Hash
}

type AuctionClosed struct {
	Phase        gsrpcTypes.Phase
	AuctionIndex gsrpcTypes.U32
	Topics       []gsrpcTypes.Hash
}

type Reserved struct {
	Phase         gsrpcTypes.Phase
	Bidder        gsrpcTypes.AccountID
	ExtraReserved gsrpcTypes.U128
	TotalAmount   gsrpcTypes.U128
	Topics        []gsrpcTypes.Hash
}

type Unreserved struct {
	Phase  gsrpcTypes.Phase
	Bidder gsrpcTypes.AccountID
	Amount gsrpcTypes.U128
	Topics []gsrpcTypes.Hash
}

type ReserveConfiscated struct {
	Phase  gsrpcTypes.Phase
	ParaId gsrpcTypes.U32
	Leaser gsrpcTypes.AccountID
	Amount gsrpcTypes.U128
	Topics []gsrpcTypes.Hash
}

type BidAccepted struct {
	Phase     gsrpcTypes.Phase
	Bidder    gsrpcTypes.AccountID
	ParaId    gsrpcTypes.U32
	Amount    gsrpcTypes.U128
	FirstSlot gsrpcTypes.U32
	LastSlot  gsrpcTypes.U32
	Topics    []gsrpcTypes.Hash
}

type WinningOffset struct {
	Phase        gsrpcTypes.Phase
	AuctionIndex gsrpcTypes.U32
	BlockNumber  gsrpcTypes.U32
	Topics       []gsrpcTypes.Hash
}

/// Slot

type NewLeasePeriod struct {
	Phase       gsrpcTypes.Phase
	PeriodBegin gsrpcTypes.U32
	Topics      []gsrpcTypes.Hash
}

type Leased struct {
	Phase        gsrpcTypes.Phase
	ParaId       gsrpcTypes.U32
	Leaser       gsrpcTypes.AccountID
	PeriodBegin  gsrpcTypes.U32
	PeriodCount  gsrpcTypes.U32
	ExtraReseved gsrpcTypes.U128
	TotalAmount  gsrpcTypes.U128
	Topics       []gsrpcTypes.Hash
}

type TokenError struct {
	IsNoFunds      bool
	IsWouldDie     bool
	IsBelowMinimum bool
	IsCannotCreate bool
	IsUnknownAsset bool
	IsFrozen       bool
	IsUnsupported  bool
}

func (e *TokenError) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		e.IsNoFunds = true
	case 1:
		e.IsWouldDie = true
	case 2:
		e.IsBelowMinimum = true
	case 3:
		e.IsCannotCreate = true
	case 4:
		e.IsUnknownAsset = true
	case 5:
		e.IsFrozen = true
	case 6:
		e.IsUnsupported = true
	}
	return nil
}

type ArithmeticError struct {
	IsUnderflow      bool
	IsOverflow       bool
	IsDivisionByZero bool
}

func (e *ArithmeticError) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		e.IsUnderflow = true
	case 1:
		e.IsOverflow = true
	case 2:
		e.IsDivisionByZero = true
	}

	return nil
}

type DispatchError struct {
	IsOther        bool
	IsCannotLookup bool
	IsBadOrigin    bool
	IsModule       bool
	AsModule       struct {
		Index gsrpcTypes.U8
		Error gsrpcTypes.U8
	}
	IsConsumerRemaining bool
	IsNoProviders       bool
	IsToken             bool
	AsToken             TokenError
	IsArithmetic        bool
	AsArithmetic        ArithmeticError
}

func (e *DispatchError) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		e.IsOther = true
	case 1:
		e.IsCannotLookup = true
	case 2:
		e.IsBadOrigin = true
	case 3:
		e.IsModule = true
		err = decoder.Decode(&e.AsModule)
		if err != nil {
			return err
		}
	case 4:
		e.IsConsumerRemaining = true
	case 5:
		e.IsNoProviders = true
	case 6:
		e.IsToken = true
		err = decoder.Decode(&e.AsToken)
		if err != nil {
			return err
		}
	case 7:
		e.IsArithmetic = true
		err = decoder.Decode(&e.AsArithmetic)
		if err != nil {
			return err
		}
	}

	return nil
}

// DispatchResult can be returned from dispatchable functions
type DispatchResult struct {
	IsOk    bool
	IsError bool
	AsError DispatchError
}

func (r *DispatchResult) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		r.IsOk = true
	case 1:
		r.IsError = true
		err = decoder.Decode(&r.AsError)
		if err != nil {
			return err
		}
	}

	return nil
}
