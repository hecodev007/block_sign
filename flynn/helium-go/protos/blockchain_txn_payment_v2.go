package protos

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"reflect"
	"sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type BlockchainTxnPaymentV2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payer     []byte     `protobuf:"bytes,1,opt,name=payer,proto3" json:"payer,omitempty"`
	Payments  []*Payment `protobuf:"bytes,2,rep,name=payments,proto3" json:"payments,omitempty"`
	Fee       uint64     `protobuf:"varint,3,opt,name=fee,proto3" json:"fee,omitempty"`
	Nonce     uint64     `protobuf:"varint,4,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Signature []byte     `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *BlockchainTxnPaymentV2) Reset() {
	*x = BlockchainTxnPaymentV2{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchain_txn_payment_v2_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockchainTxnPaymentV2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockchainTxnPaymentV2) ProtoMessage() {}
func (x *BlockchainTxnPaymentV2) ProtoReflect() protoreflect.Message {
	mi := &file_blockchain_txn_payment_v2_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockchainTxnPaymentV2.ProtoReflect.Descriptor instead.
func (*BlockchainTxnPaymentV2) Descriptor() ([]byte, []int) {
	return file_blockchain_txn_payment_v2_proto_rawDescGZIP(), []int{0}
}

func (x *BlockchainTxnPaymentV2) GetPayer() []byte {
	if x != nil {
		return x.Payer
	}
	return nil
}

func (x *BlockchainTxnPaymentV2) GetPayments() []*Payment {
	if x != nil {
		return x.Payments
	}
	return nil
}

func (x *BlockchainTxnPaymentV2) GetFee() uint64 {
	if x != nil {
		return x.Fee
	}
	return 0
}

func (x *BlockchainTxnPaymentV2) GetNonce() uint64 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *BlockchainTxnPaymentV2) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type Payment struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payee  []byte `protobuf:"bytes,1,opt,name=payee,proto3" json:"payee,omitempty"`
	Amount uint64 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *Payment) Reset() {
	*x = Payment{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchain_txn_payment_v2_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Payment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Payment) ProtoMessage() {}

func (x *Payment) ProtoReflect() protoreflect.Message {
	mi := &file_blockchain_txn_payment_v2_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Payment.ProtoReflect.Descriptor instead.
func (*Payment) Descriptor() ([]byte, []int) {
	return file_blockchain_txn_payment_v2_proto_rawDescGZIP(), []int{1}
}

func (x *Payment) GetPayee() []byte {
	if x != nil {
		return x.Payee
	}
	return nil
}

func (x *Payment) GetAmount() uint64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

var File_blockchain_txn_payment_v2_proto protoreflect.FileDescriptor

var file_blockchain_txn_payment_v2_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e,
	0x5f, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x22, 0xa4, 0x01, 0x0a, 0x19, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e, 0x5f, 0x70, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x32, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x61, 0x79, 0x65, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x70, 0x61, 0x79, 0x65, 0x72, 0x12, 0x2b, 0x0a,
	0x08, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74,
	0x52, 0x08, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x65,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x66, 0x65, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x6e, 0x6f, 0x6e,
	0x63, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x22, 0x37, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x70,
	0x61, 0x79, 0x65, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x70, 0x61, 0x79, 0x65,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_blockchain_txn_payment_v2_proto_rawDescOnce sync.Once
	file_blockchain_txn_payment_v2_proto_rawDescData = file_blockchain_txn_payment_v2_proto_rawDesc
)

func file_blockchain_txn_payment_v2_proto_rawDescGZIP() []byte {
	file_blockchain_txn_payment_v2_proto_rawDescOnce.Do(func() {
		file_blockchain_txn_payment_v2_proto_rawDescData = protoimpl.X.CompressGZIP(file_blockchain_txn_payment_v2_proto_rawDescData)
	})
	return file_blockchain_txn_payment_v2_proto_rawDescData
}

var file_blockchain_txn_payment_v2_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_blockchain_txn_payment_v2_proto_goTypes = []interface{}{
	(*BlockchainTxnPaymentV2)(nil), // 0: protos.blockchain_txn_payment_v2
	(*Payment)(nil),                // 1: protos.payment
}
var file_blockchain_txn_payment_v2_proto_depIdxs = []int32{
	1, // 0: protos.blockchain_txn_payment_v2.payments:type_name -> protos.payment
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_blockchain_txn_payment_v2_proto_init() }
func file_blockchain_txn_payment_v2_proto_init() {
	if File_blockchain_txn_payment_v2_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_blockchain_txn_payment_v2_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockchainTxnPaymentV2); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_blockchain_txn_payment_v2_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Payment); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_blockchain_txn_payment_v2_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_blockchain_txn_payment_v2_proto_goTypes,
		DependencyIndexes: file_blockchain_txn_payment_v2_proto_depIdxs,
		MessageInfos:      file_blockchain_txn_payment_v2_proto_msgTypes,
	}.Build()
	File_blockchain_txn_payment_v2_proto = out.File
	file_blockchain_txn_payment_v2_proto_rawDesc = nil
	file_blockchain_txn_payment_v2_proto_goTypes = nil
	file_blockchain_txn_payment_v2_proto_depIdxs = nil
}
