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

type BlockchainTxn struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Txn:
	//	*BlockchainTxn_Payment
	//	*BlockchainTxn_PaymentV2
	Txn IsBlockchainTxn_Txn `protobuf_oneof:"txn"`
}

func (x *BlockchainTxn) Reset() {
	*x = BlockchainTxn{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchain_txn_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockchainTxn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockchainTxn) ProtoMessage() {}

func (x *BlockchainTxn) ProtoReflect() protoreflect.Message {
	mi := &file_blockchain_txn_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}
func (x *BlockchainTxn) SetBlockTxn(txn IsBlockchainTxn_Txn) {
	x.Txn = txn
}

// Deprecated: Use BlockchainTxn.ProtoReflect.Descriptor instead.
func (*BlockchainTxn) Descriptor() ([]byte, []int) {
	return file_blockchain_txn_proto_rawDescGZIP(), []int{0}
}

func (m *BlockchainTxn) GetTxn() IsBlockchainTxn_Txn {
	if m != nil {
		return m.Txn
	}
	return nil
}

func (x *BlockchainTxn) GetPayment() *BlockchainTxnPaymentV1 {
	if x, ok := x.GetTxn().(*BlockchainTxn_Payment); ok {
		return x.Payment
	}
	return nil
}

func (x *BlockchainTxn) GetPaymentV2() *BlockchainTxnPaymentV2 {
	if x, ok := x.GetTxn().(*BlockchainTxn_PaymentV2); ok {
		return x.PaymentV2
	}
	return nil
}

type IsBlockchainTxn_Txn interface {
	isBlockchainTxn_Txn()
}

type BlockchainTxn_Payment struct {
	Payment *BlockchainTxnPaymentV1 `protobuf:"bytes,8,opt,name=payment,proto3,oneof"`
}

type BlockchainTxn_PaymentV2 struct {
	PaymentV2 *BlockchainTxnPaymentV2 `protobuf:"bytes,24,opt,name=payment_v2,json=paymentV2,proto3,oneof"`
}

func (*BlockchainTxn_Payment) isBlockchainTxn_Txn() {}

func (*BlockchainTxn_PaymentV2) isBlockchainTxn_Txn() {}

type BlockchainTxnBundleV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Transactions []*BlockchainTxn `protobuf:"bytes,1,rep,name=transactions,proto3" json:"transactions,omitempty"`
}

func (x *BlockchainTxnBundleV1) Reset() {
	*x = BlockchainTxnBundleV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchain_txn_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockchainTxnBundleV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockchainTxnBundleV1) ProtoMessage() {}

func (x *BlockchainTxnBundleV1) ProtoReflect() protoreflect.Message {
	mi := &file_blockchain_txn_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockchainTxnBundleV1.ProtoReflect.Descriptor instead.
func (*BlockchainTxnBundleV1) Descriptor() ([]byte, []int) {
	return file_blockchain_txn_proto_rawDescGZIP(), []int{1}
}

func (x *BlockchainTxnBundleV1) GetTransactions() []*BlockchainTxn {
	if x != nil {
		return x.Transactions
	}
	return nil
}

var File_blockchain_txn_proto protoreflect.FileDescriptor

var file_blockchain_txn_proto_rawDesc = []byte{
	0x0a, 0x14, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x1a, 0x1f,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e, 0x5f, 0x70,
	0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e, 0x5f,
	0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x9a, 0x01, 0x0a, 0x0e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f,
	0x74, 0x78, 0x6e, 0x12, 0x3d, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e, 0x5f, 0x70, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x31, 0x48, 0x00, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6d, 0x65,
	0x6e, 0x74, 0x12, 0x42, 0x0a, 0x0a, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x32,
	0x18, 0x18, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e, 0x5f, 0x70,
	0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x32, 0x48, 0x00, 0x52, 0x09, 0x70, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x56, 0x32, 0x42, 0x05, 0x0a, 0x03, 0x74, 0x78, 0x6e, 0x22, 0x56, 0x0a,
	0x18, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e, 0x5f,
	0x62, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x5f, 0x76, 0x31, 0x12, 0x3a, 0x0a, 0x0c, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x16, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x5f, 0x74, 0x78, 0x6e, 0x52, 0x0c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_blockchain_txn_proto_rawDescOnce sync.Once
	file_blockchain_txn_proto_rawDescData = file_blockchain_txn_proto_rawDesc
)

func file_blockchain_txn_proto_rawDescGZIP() []byte {
	file_blockchain_txn_proto_rawDescOnce.Do(func() {
		file_blockchain_txn_proto_rawDescData = protoimpl.X.CompressGZIP(file_blockchain_txn_proto_rawDescData)
	})
	return file_blockchain_txn_proto_rawDescData
}

var file_blockchain_txn_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_blockchain_txn_proto_goTypes = []interface{}{
	(*BlockchainTxn)(nil),          // 0: protos.blockchain_txn
	(*BlockchainTxnBundleV1)(nil),  // 1: protos.blockchain_txn_bundle_v1
	(*BlockchainTxnPaymentV1)(nil), // 2: protos.blockchain_txn_payment_v1
	(*BlockchainTxnPaymentV2)(nil), // 3: protos.blockchain_txn_payment_v2
}
var file_blockchain_txn_proto_depIdxs = []int32{
	2, // 0: protos.blockchain_txn.payment:type_name -> protos.blockchain_txn_payment_v1
	3, // 1: protos.blockchain_txn.payment_v2:type_name -> protos.blockchain_txn_payment_v2
	0, // 2: protos.blockchain_txn_bundle_v1.transactions:type_name -> protos.blockchain_txn
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_blockchain_txn_proto_init() }
func file_blockchain_txn_proto_init() {
	if File_blockchain_txn_proto != nil {
		return
	}
	file_blockchain_txn_payment_v1_proto_init()
	file_blockchain_txn_payment_v2_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_blockchain_txn_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockchainTxn); i {
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
		file_blockchain_txn_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockchainTxnBundleV1); i {
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
	file_blockchain_txn_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*BlockchainTxn_Payment)(nil),
		(*BlockchainTxn_PaymentV2)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_blockchain_txn_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_blockchain_txn_proto_goTypes,
		DependencyIndexes: file_blockchain_txn_proto_depIdxs,
		MessageInfos:      file_blockchain_txn_proto_msgTypes,
	}.Build()
	File_blockchain_txn_proto = out.File
	file_blockchain_txn_proto_rawDesc = nil
	file_blockchain_txn_proto_goTypes = nil
	file_blockchain_txn_proto_depIdxs = nil
}
