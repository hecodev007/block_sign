package extra

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
	gsrpcTypes "github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

// https://github.com/paritytech/polkadot/blob/master/xcm/src/v0/traits.rs#L25
type XCMError struct {
	IsUndefined                 bool
	IsOverflow                  bool
	IsUnimplemented             bool
	IsUnhandledXcmVersion       bool
	IsUnhandledXcmMessage       bool
	IsUnhandledEffect           bool
	IsEscalationOfPrivilege     bool
	IsUntrustedReserveLocation  bool
	IsUntrustedTeleportLocation bool
	IsDestinationBufferOverflow bool
	IsSendFailed                bool
	IsCannotReachDestination    bool
	AsCannotReachDestination    struct {
		Dest MultiLocation
		Xcm  Xcm
	}
	IsMultiLocationFull     bool
	IsFailedToDecode        bool
	IsBadOrigin             bool
	IsExceedsMaxMessageSize bool
	IsFailedToTransactAsset bool
	IsWeightLimitReached    bool
	IsWildcard              bool
	IsTooMuchWeightRequired bool
	IsNotHoldingFees        bool
	IsWeightNotComputable   bool
	IsBarrier               bool
	IsNotWithdrawable       bool
	IsLocationCannotHold    bool
	IsTooExpensive          bool
	IsAssetNotFound         bool
}

func (x *XCMError) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		x.IsUndefined = true
	case 1:
		x.IsOverflow = true
	case 2:
		x.IsUnimplemented = true
	case 3:
		x.IsUnhandledXcmVersion = true
	case 4:
		x.IsUnhandledXcmMessage = true
	case 5:
		x.IsUnhandledEffect = true
	case 6:
		x.IsEscalationOfPrivilege = true
	case 7:
		x.IsUntrustedReserveLocation = true
	case 8:
		x.IsUntrustedTeleportLocation = true
	case 9:
		x.IsDestinationBufferOverflow = true
	case 10:
		x.IsSendFailed = true
	case 11:
		x.IsCannotReachDestination = true
		err = decoder.Decode(&x.AsCannotReachDestination)
		if err != nil {
			return err
		}
	case 12:
		x.IsMultiLocationFull = true
	case 13:
		x.IsFailedToDecode = true
	case 14:
		x.IsBadOrigin = true
	case 15:
		x.IsExceedsMaxMessageSize = true
	case 16:
		x.IsFailedToTransactAsset = true
	case 17:
		x.IsWeightLimitReached = true
	case 18:
		x.IsWildcard = true
	case 19:
		x.IsTooMuchWeightRequired = true
	case 20:
		x.IsNotHoldingFees = true
	case 21:
		x.IsWeightNotComputable = true
	case 22:
		x.IsBarrier = true
	case 23:
		x.IsNotWithdrawable = true
	case 24:
		x.IsLocationCannotHold = true
	case 25:
		x.IsTooExpensive = true
	case 26:
		x.IsAssetNotFound = true
	}

	return nil
}

func (x XCMError) Encode(encoder scale.Encoder) error {
	var err error

	switch {
	case x.IsUndefined:
		err = encoder.PushByte(0)
	case x.IsOverflow:
		err = encoder.PushByte(1)
	case x.IsUnimplemented:
		err = encoder.PushByte(2)
	case x.IsUnhandledXcmVersion:
		err = encoder.PushByte(3)
	case x.IsUnhandledXcmMessage:
		err = encoder.PushByte(4)
	case x.IsUnhandledEffect:
		err = encoder.PushByte(5)
	case x.IsEscalationOfPrivilege:
		err = encoder.PushByte(6)
	case x.IsUntrustedReserveLocation:
		err = encoder.PushByte(7)
	case x.IsUntrustedTeleportLocation:
		err = encoder.PushByte(8)
	case x.IsDestinationBufferOverflow:
		err = encoder.PushByte(9)
	case x.IsSendFailed:
		err = encoder.PushByte(10)
	case x.IsCannotReachDestination:
		err = encoder.PushByte(11)
	case x.IsMultiLocationFull:
		err = encoder.PushByte(12)
	case x.IsFailedToDecode:
		err = encoder.PushByte(13)
	case x.IsBadOrigin:
		err = encoder.PushByte(14)
	case x.IsExceedsMaxMessageSize:
		err = encoder.PushByte(15)
	case x.IsFailedToTransactAsset:
		err = encoder.PushByte(16)
	case x.IsWeightLimitReached:
		err = encoder.PushByte(17)
	case x.IsWildcard:
		err = encoder.PushByte(18)
	case x.IsTooMuchWeightRequired:
		err = encoder.PushByte(19)
	case x.IsNotHoldingFees:
		err = encoder.PushByte(20)
	case x.IsWeightNotComputable:
		err = encoder.PushByte(21)
	case x.IsBarrier:
		err = encoder.PushByte(22)
	case x.IsNotWithdrawable:
		err = encoder.PushByte(23)
	case x.IsLocationCannotHold:
		err = encoder.PushByte(24)
	case x.IsTooExpensive:
		err = encoder.PushByte(25)
	case x.IsAssetNotFound:
		err = encoder.PushByte(26)
	}

	if err != nil {
		return err
	}

	return nil
}

/*
	/// Execution completed successfully; given weight was used.
	Complete(Weight),
	/// Execution started, but did not complete successfully due to the given error; given weight was used.
	Incomplete(Weight, Error),
	/// Execution did not start due to the given error.
	Error(Error),
*/
type Incomplete struct {
	Weight gsrpcTypes.U64
	Error  XCMError
}

type OutCome struct {
	IsComplete   bool
	AsComplete   gsrpcTypes.U64
	IsIncomplete bool
	AsIncomplete Incomplete
	IsError      bool
	AsError      XCMError
}

func (o *OutCome) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		o.IsComplete = true
		err = decoder.Decode(&o.AsComplete)
		if err != nil {
			return err
		}
		return nil
	case 1:
		o.IsIncomplete = true
		err = decoder.Decode(&o.AsIncomplete)
		if err != nil {
			return err
		}
		return nil
	default:
		o.IsError = true
		err = decoder.Decode(&o.AsError)
		if err != nil {
			return err
		}
		return nil
	}
}

type NetworkId struct {
	IsAny      bool
	IsNamed    bool
	AsNamed    gsrpcTypes.Bytes
	IsPolkadot bool
	IsKusama   bool
}

func (n *NetworkId) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		n.IsAny = true
	case 1:
		n.IsNamed = true
		err = decoder.Decode(&n.AsNamed)
		if err != nil {
			return err
		}
	case 2:
		n.IsPolkadot = true
	case 3:
		n.IsKusama = true
	}

	return nil
}

type AccountId32 struct {
	NetworkId NetworkId
	Id        gsrpcTypes.Bytes32
}

type AccountIndex64 struct {
	NetworkId NetworkId
	Index     gsrpcTypes.UCompact
}

type AccountKey20 struct {
	NetworkId NetworkId
	Key       [20]byte
}

type BodyId struct {
	IsUnit        bool
	IsNamed       bool
	AsNamed       gsrpcTypes.Bytes
	IsIndex       bool
	AsIndex       gsrpcTypes.UCompact
	IsExecutive   bool
	IsTechnical   bool
	IsLegislative bool
	IsJudicial    bool
}

func (body *BodyId) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		body.IsUnit = true
	case 1:
		body.IsNamed = true
		err = decoder.Decode(&body.AsNamed)
		if err != nil {
			return err
		}
	case 2:
		body.IsIndex = true
		err = decoder.Decode(&body.AsIndex)
		if err != nil {
			return err
		}
	case 3:
		body.IsExecutive = true
	case 4:
		body.IsTechnical = true
	case 5:
		body.IsLegislative = true
	case 6:
		body.IsJudicial = true
	}
	return nil
}

type BodyPart struct {
	IsVoice    bool
	IsMembers  bool
	AsMembers  gsrpcTypes.UCompact
	IsFraction bool
	AsFraction struct {
		Nom   gsrpcTypes.UCompact
		Denom gsrpcTypes.UCompact
	}
	IsAtLeastProportion bool
	AsAtLeastProportion struct {
		Nom   gsrpcTypes.UCompact
		Denom gsrpcTypes.UCompact
	}
	IsMoreThanProportion bool
	AsMoreThanProportion struct {
		Nom   gsrpcTypes.UCompact
		Denom gsrpcTypes.UCompact
	}
}

func (body *BodyPart) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		body.IsVoice = true
	case 1:
		body.IsMembers = true
		err = decoder.Decode(&body.AsMembers)
		if err != nil {
			return err
		}
	case 2:
		body.IsFraction = true
		err = decoder.Decode(&body.AsFraction)
		if err != nil {
			return err
		}
	case 3:
		body.IsAtLeastProportion = true
		err = decoder.Decode(&body.AsAtLeastProportion)
		if err != nil {
			return err
		}
	case 4:
		body.IsMoreThanProportion = true
		err = decoder.Decode(&body.AsMoreThanProportion)
		if err != nil {
			return err
		}
	}

	return nil
}

type Plurality struct {
	Id   BodyId
	Part BodyPart
}

type Junction struct {
	IsParent         bool
	IsParachain      bool
	AsParachain      gsrpcTypes.UCompact
	IsAccountId32    bool
	AsAccountId32    AccountId32
	IsAccountIndex64 bool
	AsAccountIndex64 AccountIndex64
	IsAccountKey20   bool
	AsAccountKey20   AccountKey20
	IsPalletInstance bool
	AsPalletInstance gsrpcTypes.U8
	IsGeneralIndex   bool
	AsGeneralIndex   gsrpcTypes.UCompact
	IsGeneralKey     bool
	AsGeneralKey     gsrpcTypes.Bytes
	IsOnlyChild      bool
	IsPlurality      bool
	AsPlurality      Plurality
}

func (j *Junction) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		j.IsParent = true
	case 1:
		j.IsParachain = true
		err = decoder.Decode(&j.AsParachain)
		if err != nil {
			return err
		}
	case 2:
		j.IsAccountId32 = true
		err = decoder.Decode(&j.AsAccountId32)
		if err != nil {
			return err
		}
	case 3:
		j.IsAccountIndex64 = true
		err = decoder.Decode(&j.AsAccountIndex64)
		if err != nil {
			return err
		}
	case 4:
		j.IsAccountKey20 = true
		err = decoder.Decode(&j.AsAccountKey20)
		if err != nil {
			return err
		}
	case 5:
		j.IsPalletInstance = true
		err = decoder.Decode(&j.AsPalletInstance)
		if err != nil {
			return err
		}
	case 6:
		j.IsGeneralIndex = true
		err = decoder.Decode(&j.AsGeneralIndex)
		if err != nil {
			return err
		}
	case 7:
		j.IsGeneralKey = true
		err = decoder.Decode(&j.AsGeneralKey)
		if err != nil {
			return err
		}
	case 8:
		j.IsOnlyChild = true
	case 9:
		j.IsPlurality = true
		err = decoder.Decode(&j.AsPlurality)
		if err != nil {
			return err
		}
	}

	return nil
}

type MultiLocation struct {
	IsNull bool
	IsX1   bool
	AsX1   Junction
	IsX2   bool
	AsX2   struct {
		X1 Junction
		X2 Junction
	}
	IsX3 bool
	AsX3 struct {
		X1 Junction
		X2 Junction
		X3 Junction
	}
	IsX4 bool
	AsX4 struct {
		X1 Junction
		X2 Junction
		X3 Junction
		X4 Junction
	}
	IsX5 bool
	AsX5 struct {
		X1 Junction
		X2 Junction
		X3 Junction
		X4 Junction
		X5 Junction
	}
	IsX6 bool
	AsX6 struct {
		X1 Junction
		X2 Junction
		X3 Junction
		X4 Junction
		X5 Junction
		X6 Junction
	}
	IsX7 bool
	AsX7 struct {
		X1 Junction
		X2 Junction
		X3 Junction
		X4 Junction
		X5 Junction
		X6 Junction
		X7 Junction
	}
	IsX8 bool
	AsX8 struct {
		X1 Junction
		X2 Junction
		X3 Junction
		X4 Junction
		X5 Junction
		X6 Junction
		X7 Junction
		X8 Junction
	}
}

func (m *MultiLocation) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		m.IsNull = true
	case 1:
		m.IsX1 = true
		err = decoder.Decode(&m.AsX1)
		if err != nil {
			return err
		}
	case 2:
		m.IsX2 = true
		err = decoder.Decode(&m.AsX2)
		if err != nil {
			return err
		}
	case 3:
		m.IsX3 = true
		err = decoder.Decode(&m.AsX3)
		if err != nil {
			return err
		}
	case 4:
		m.IsX4 = true
		err = decoder.Decode(&m.AsX4)
		if err != nil {
			return err
		}
	case 5:
		m.IsX5 = true
		err = decoder.Decode(&m.AsX5)
		if err != nil {
			return err
		}
	case 6:
		m.IsX6 = true
		err = decoder.Decode(&m.AsX6)
		if err != nil {
			return err
		}
	case 7:
		m.IsX7 = true
		err = decoder.Decode(&m.AsX7)
		if err != nil {
			return err
		}
	}

	return nil
}

type AssetInstance struct {
	IsUndefined bool
	IsIndex     bool
	AsIndex     gsrpcTypes.UCompact
	IsArray4    bool
	AsArray4    [4]byte
	IsArray8    bool
	AsArray8    [8]byte
	IsArray16   bool
	AsArray16   [16]byte
	IsArray32   bool
	AsArray32   [32]byte
	IsBlob      bool
	AsBlob      gsrpcTypes.Bytes
}

func (a *AssetInstance) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		a.IsUndefined = true
	case 1:
		a.IsIndex = true
		err = decoder.Decode(&a.AsIndex)
		if err != nil {
			return err
		}
	case 2:
		a.IsArray4 = true
		err = decoder.Decode(&a.AsArray4)
		if err != nil {
			return err
		}
	case 3:
		a.IsArray8 = true
		err = decoder.Decode(&a.AsArray8)
		if err != nil {
			return err
		}
	case 4:
		a.IsArray16 = true
		err = decoder.Decode(&a.AsArray16)
		if err != nil {
			return err
		}
	case 5:
		a.IsArray32 = true
		err = decoder.Decode(&a.AsArray32)
		if err != nil {
			return err
		}
	case 6:
		a.IsBlob = true
		err = decoder.Decode(&a.AsBlob)
		if err != nil {
			return err
		}
	}

	return nil
}

type MultiAsset struct {
	IsNone                   bool
	IsAll                    bool
	IsAllFungible            bool
	IsAllNonFungible         bool
	IsAllAbstractFungible    bool
	AsAllAbstractFungible    gsrpcTypes.Bytes
	IsAllAbstractNonFungible bool
	AsAllAbstractNonFungible gsrpcTypes.Bytes
	IsAllConcreteFungible    bool
	AsAllConcreteFungible    MultiLocation
	IsAllConcreteNonFungible bool
	AsAllConcreteNonFungible MultiLocation
	IsAbstractFungible       bool
	AsAbstractFungible       struct {
		Id     gsrpcTypes.Bytes
		Amount gsrpcTypes.UCompact
	}
	IsAbstractNonFungible bool
	AsAbstractNonFungible struct {
		Class    gsrpcTypes.Bytes
		Instance AssetInstance
	}
	IsConcreteFungible bool
	AsConcreteFungible struct {
		Id     MultiLocation
		Amount gsrpcTypes.UCompact
	}
	IsConcreteNonFungible bool
	AsConcreteNonFungible struct {
		Class    MultiLocation
		Instance AssetInstance
	}
}

func (m *MultiAsset) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		m.IsNone = true
	case 1:
		m.IsAll = true
	case 2:
		m.IsAllFungible = true
	case 3:
		m.IsAllNonFungible = true
	case 4:
		m.IsAllAbstractFungible = true
		err = decoder.Decode(&m.AsAllAbstractFungible)
		if err != nil {
			return err
		}
	case 5:
		m.IsAllAbstractNonFungible = true
		err = decoder.Decode(&m.AsAllAbstractNonFungible)
		if err != nil {
			return err
		}
	case 6:
		m.IsAllConcreteFungible = true
		err = decoder.Decode(&m.AsAllConcreteFungible)
		if err != nil {
			return err
		}
	case 7:
		m.IsAllConcreteNonFungible = true
		err = decoder.Decode(&m.AsAllConcreteNonFungible)
		if err != nil {
			return err
		}
	case 8:
		m.IsAbstractFungible = true
		err = decoder.Decode(&m.AsAbstractFungible)
		if err != nil {
			return err
		}
	case 9:
		m.IsAbstractNonFungible = true
		err = decoder.Decode(&m.AsAbstractNonFungible)
		if err != nil {
			return err
		}
	case 10:
		m.IsConcreteFungible = true
		err = decoder.Decode(&m.AsConcreteFungible)
		if err != nil {
			return err
		}
	case 11:
		m.IsConcreteNonFungible = true
		err = decoder.Decode(&m.AsConcreteNonFungible)
		if err != nil {
			return err
		}
	}

	return nil
}

type OriginKind struct {
	IsNative           bool
	IsSovereignAccount bool
	IsSuperuser        bool
	IsXcm              bool
}

func (o *OriginKind) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		o.IsNative = true
	case 1:
		o.IsSovereignAccount = true
	case 2:
		o.IsSuperuser = true
	case 3:
		o.IsXcm = true
	}

	return nil
}

/*
	Null,
	DepositAsset { assets: Vec<MultiAsset>, dest: MultiLocation },
	DepositReserveAsset { assets: Vec<MultiAsset>, dest: MultiLocation, effects: Vec<Order<()>> },
	ExchangeAsset { give: Vec<MultiAsset>, receive: Vec<MultiAsset> },
	InitiateReserveWithdraw { assets: Vec<MultiAsset>, reserve: MultiLocation, effects: Vec<Order<()>> },
	InitiateTeleport { assets: Vec<MultiAsset>, dest: MultiLocation, effects: Vec<Order<()>> },
	QueryHolding { #[codec(compact)] query_id: u64, dest: MultiLocation, assets: Vec<MultiAsset> },
	BuyExecution { fees: MultiAsset, weight: u64, debt: u64, halt_on_error: bool, xcm: Vec<Xcm<Call>> },
*/
type Order struct {
	IsNull         bool
	IsDepositAsset bool
	AsDepositAsset struct {
		Assets []MultiAsset
		Dest   MultiLocation
	}
	IsDepositReserveAsset bool
	AsDepositReserveAsset struct {
		Assets  []MultiAsset
		Dest    MultiLocation
		Effects []Order
	}
	IsExchangeAsset bool
	AsExchangeAsset struct {
		Give    []MultiAsset
		Receive []MultiAsset
	}
	IsInitiateReserveWithdraw bool
	AsInitiateReserveWithdraw struct {
		Assets  []MultiAsset
		Reserve MultiLocation
		Effects []Order
	}
	IsInitiateTeleport bool
	AsInitiateTeleport struct {
		Assets  []MultiAsset
		Dest    MultiLocation
		Effects []Order
	}
	IsQueryHolding bool
	AsQueryHolding struct {
		QueryId gsrpcTypes.UCompact
		Dest    MultiLocation
		Assets  []MultiAsset
	}
	IsBuyExecution bool
	AsBuyExecution struct {
		Fees        MultiAsset
		Weight      gsrpcTypes.U64
		Debt        gsrpcTypes.U64
		HaltOnError gsrpcTypes.Bool
		Xcm         []Xcm
	}
}

/*
	WithdrawAsset { assets: Vec<MultiAsset>, effects: Vec<Order<Call>> },
	ReserveAssetDeposit { assets: Vec<MultiAsset>, effects: Vec<Order<Call>> },
	TeleportAsset { assets: Vec<MultiAsset>, effects: Vec<Order<Call>> },
	QueryResponse { #[codec(compact)] query_id: u64, response: Response },
	TransferAsset { assets: Vec<MultiAsset>, dest: MultiLocation },
	TransferReserveAsset { assets: Vec<MultiAsset>, dest: MultiLocation, effects: Vec<Order<()>> }
	Transact { origin_type: OriginKind, require_weight_at_most: u64, call: DoubleEncoded<Call> },
	HrmpNewChannelOpenRequest {
		#[codec(compact)] sender: u32,
		#[codec(compact)] max_message_size: u32,
		#[codec(compact)] max_capacity: u32,
	},
	HrmpChannelAccepted {
		#[codec(compact)] recipient: u32,
	},
	HrmpChannelClosing {
		#[codec(compact)] initiator: u32,
		#[codec(compact)] sender: u32,
		#[codec(compact)] recipient: u32,
	},
	RelayedFrom {
		who: MultiLocation,
		message: alloc::boxed::Box<Xcm<Call>>,
	},
*/
type Xcm struct {
	IsWithdrawAsset bool
	AsWithdrawAsset struct {
		Assets  []MultiAsset
		Effects []Order
	}
	IsReserveAssetDeposit bool
	AsReserveAssetDeposit struct {
		Assets  []MultiAsset
		Effects []Order
	}
	IsTeleportAsset bool
	AsTeleportAsset struct {
		Assets  []MultiAsset
		Effects []Order
	}
	IsQueryResponse bool
	AsQueryResponse struct {
		QueryId  gsrpcTypes.UCompact
		Response []MultiAsset
	}
	IsTransferAsset bool
	AsTransferAsset struct {
		Assets []MultiAsset
		Dest   MultiLocation
	}
	IsTransferReserveAsset bool
	AsTransferReserveAsset struct {
		Assets  []MultiAsset
		Dest    MultiLocation
		Effects []Order
	}
	IsTransact bool
	AsTransact struct {
		OriginType OriginKind
		Weight     gsrpcTypes.U64
		Call       struct {
			Encode gsrpcTypes.Bytes
		}
	}
	IsHrmpNewChannelOpenRequest bool
	AsHrmpNewChannelOpenRequest struct {
		Sender   gsrpcTypes.UCompact
		MaxSize  gsrpcTypes.UCompact
		Capacity gsrpcTypes.UCompact
	}
	IsHrmpChannelAccepted bool
	AsHrmpChannelAccepted struct {
		Recipient gsrpcTypes.UCompact
	}
	IsHrmpChannelClosing bool
	AsHrmpChannelClosing struct {
		Initiator gsrpcTypes.UCompact
		Sender    gsrpcTypes.UCompact
		Recipient gsrpcTypes.UCompact
	}
	IsRelayedFrom bool
	AsRelayedFrom struct {
		Who     MultiLocation
		Message *Xcm
	}
}

func (x *Xcm) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		x.IsWithdrawAsset = true
		err = decoder.Decode(&x.AsWithdrawAsset)
		if err != nil {
			return err
		}
	case 1:
		x.IsReserveAssetDeposit = true
		err = decoder.Decode(&x.AsReserveAssetDeposit)
		if err != nil {
			return err
		}
	case 2:
		x.IsTeleportAsset = true
		err = decoder.Decode(&x.AsTeleportAsset)
		if err != nil {
			return err
		}
	case 3:
		x.IsQueryResponse = true
		err = decoder.Decode(&x.AsQueryResponse)
		if err != nil {
			return err
		}
	case 4:
		x.IsTransferAsset = true
		err = decoder.Decode(&x.AsTransferAsset)
		if err != nil {
			return err
		}
	case 5:
		x.IsTransferReserveAsset = true
		err = decoder.Decode(&x.AsTransferReserveAsset)
		if err != nil {
			return err
		}
	case 6:
		x.IsTransact = true
		err = decoder.Decode(&x.AsTransact)
		if err != nil {
			return err
		}
	case 7:
		x.IsHrmpNewChannelOpenRequest = true
		err = decoder.Decode(&x.AsHrmpNewChannelOpenRequest)
		if err != nil {
			return err
		}
	case 8:
		x.IsHrmpChannelAccepted = true
		err = decoder.Decode(&x.AsHrmpChannelAccepted)
		if err != nil {
			return err
		}
	case 9:
		x.IsHrmpChannelClosing = true
		err = decoder.Decode(&x.AsHrmpChannelClosing)
		if err != nil {
			return err
		}
	case 10:
		x.IsRelayedFrom = true
		var xcm Xcm
		err = decoder.Decode(&x.AsRelayedFrom.Who)
		if err != nil {
			return err
		}

		err = decoder.Decode(&xcm)
		if err != nil {
			return err
		}
		x.AsRelayedFrom.Message = &xcm
	}

	return nil
}
