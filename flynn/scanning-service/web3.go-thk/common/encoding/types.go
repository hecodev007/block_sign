package encoding

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func Unhex(str string) []byte {
	b, err := hex.DecodeString(strings.Replace(str, " ", "", -1))
	if err != nil {
		panic(fmt.Sprintf("invalid hex string: %q", str))
	}
	return b
}

type THValueType byte

// type header value
type THValue struct {
	N string      // name of the type header
	C byte        // code
	M byte        // mask
	W byte        // wildcard mask
	T THValueType // type of the header
}

func (thvalue THValue) Match(b byte) bool {
	return (b & thvalue.M) == thvalue.C
}

type TypeHeader byte

func (th TypeHeader) Name() string {
	thv, ok := headerTypeMap[th]
	if ok {
		return thv.N
	}
	return "TypeHeader" + strconv.Itoa(int(th))
}

func (th TypeHeader) String() string {
	return th.Name()
}

const (
	THSingleByte   TypeHeader = iota // single byte
	THZeroValue                      // zero value (empty string / false of bool)
	THTrue                           // true of bool
	THEmpty                          // empty value
	THArraySingle                    // array with no more than 16 elements
	THArrayMulti                     // array with more than 16 elements
	THPosNumSingle                   // positive number with bytes less and equal to 8
	THNegNumSingle                   // negative number with bytes less and equal to 8
	THPosBigInt                      // positive big.Int
	THNegBigInt                      // negative big.Int
	THStringSingle                   // string with length less and equal to 32
	THStringMulti                    // string with length more than 32

	THVTByte         THValueType = iota // one byte value
	THVTSingleHeader                    // single byte header
	THVTMultiHeader                     // multi bytes header

	MaxNested = 10 // max nested times when encoding. pointer, slice, array, map, struct
)

// Encoder is the interface which encoding package while invoke the Serialization()
// when encoding the object.
// ATTENTION: the receiver of Encoder.Serialization() and Decoder.Deserialization() MUST
// BE SAME. otherwise, they will not be use in same struct.
type Encoder interface {
	Serialization(w io.Writer) error
}

type Decoder interface {
	Deserialization(r io.Reader) error
}

var (
	// static encoding
	zeroValues  = []byte{headerTypeMap[THZeroValue].C}
	trueBools   = []byte{headerTypeMap[THTrue].C}
	emptyValues = []byte{headerTypeMap[THEmpty].C}

	NilOrFalse   = headerTypeMap[THZeroValue].C
	NotNilOrTrue = headerTypeMap[THTrue].C

	// header maker of encoding
	HeadMaker headMaker

	// codec for numerics
	Numeric numeric

	// big.Int
	bigint128    = big.NewInt(128)
	typeOfBigInt = reflect.TypeOf(big.Int{})

	// []interface{} type
	typeOfInterfaceSlice = reflect.TypeOf([]interface{}{})
	typeOfInterface      = reflect.TypeOf((*interface{})(nil)).Elem()

	// uint64
	typeOfUint64 = reflect.TypeOf((*uint64)(nil)).Elem()
	typeOfInt64  = reflect.TypeOf((*int64)(nil)).Elem()
	typeOfString = reflect.TypeOf("")
	typeOfByte   = reflect.TypeOf((*byte)(nil)).Elem()

	// header constants
	headerTypeMap = map[TypeHeader]THValue{
		THSingleByte:   {"SingleByte", 0x00, 0x80, ^byte(0x80), THVTByte},
		THZeroValue:    {"ZeroValue", 0x80, 0xFF, 0x00, THVTByte},
		THTrue:         {"True", 0x81, 0xFF, 0x00, THVTByte},
		THEmpty:        {"Empty", 0x82, 0xFF, 0x00, THVTByte},
		THArraySingle:  {"SmallArray", 0x90, 0xF0, ^byte(0xF0), THVTSingleHeader},
		THArrayMulti:   {"Array", 0x88, 0xF8, ^byte(0xF8), THVTMultiHeader},
		THPosNumSingle: {"PositiveNumberSingleByte", 0xA0, 0xF8, ^byte(0xF8), THVTSingleHeader},
		THNegNumSingle: {"NegativeNumberSIngleByte", 0xA8, 0xF8, ^byte(0xF8), THVTSingleHeader},
		THPosBigInt:    {"PositiveNumberMultiBytes", 0xB0, 0xF8, ^byte(0xF8), THVTMultiHeader},
		THNegBigInt:    {"NegativeNumberMultiBytes", 0xB8, 0xF8, ^byte(0xF8), THVTMultiHeader},
		THStringSingle: {"StringSingleByte", 0xC0, 0xE0, ^byte(0xE0), THVTSingleHeader},
		THStringMulti:  {"StringMultiBytes", 0xE0, 0xF8, ^byte(0xF8), THVTMultiHeader},
	}

	// primitive kind to valid TypeHeaders
	primKindTypeHeaderMap = map[reflect.Kind]map[TypeHeader]typeReaderFunc{
		reflect.Int:     intReaders,
		reflect.Int8:    intReaders,
		reflect.Int16:   intReaders,
		reflect.Int32:   intReaders,
		reflect.Int64:   intReaders,
		reflect.Uint:    uintReaders,
		reflect.Uint8:   uintReaders,
		reflect.Uint16:  uintReaders,
		reflect.Uint32:  uintReaders,
		reflect.Uint64:  uintReaders,
		reflect.Float32: floatReaders,
		reflect.Float64: floatReaders,
		reflect.Bool:    boolReaders,
		reflect.String:  stringReaders,
	}

	// cache for structFields
	typeInfoMap = new(sync.Map)

	// serialize/deserialize self
	TypeOfEncoderPtr = reflect.TypeOf((*Encoder)(nil))
	TypeOfEncoder    = TypeOfEncoderPtr.Elem()
	TypeOfDecoderPtr = reflect.TypeOf((*Decoder)(nil))
	TypeOfDecoder    = TypeOfDecoderPtr.Elem()

	// errors
	ErrUnsupported        = errors.New("unsupported")
	ErrNestingOverflow    = fmt.Errorf("nesting overflow: %d times", MaxNested)
	ErrInsufficientLength = errors.New("insufficient length of the slice")
	ErrDecode             = errors.New("decode error")
	ErrLength             = errors.New("length error")
	ErrDecodeIntoNil      = errors.New("rtl: decode pointer MUST NOT be nil")
	ErrDecodeNoPtr        = errors.New("rtl: value being decode MUST be a pointer")
)

type headMaker struct{}

// stringBuf put string header into buf, len(buf) must bigger or equal to the number of header bytes
func (headMaker) stringBuf(length int, buf []byte) int {
	if length <= 0 {
		return 0
	}

	if length <= 32 {
		buf[0] = headerTypeMap[THStringSingle].C | (byte(length) & headerTypeMap[THStringSingle].W)
		return 1
	}

	l, _ := Numeric.writeUint(buf[1:], uint64(length))
	buf[0] = headerTypeMap[THStringMulti].C | (byte(l) & headerTypeMap[THStringMulti].W)
	return l + 1
}

func (h headMaker) string(length int) []byte {
	if length <= 0 {
		return nil
	}
	r := make([]byte, 9)
	l := h.stringBuf(length, r)
	return r[:l]
}

// numericBuf put numeric header into buf, len(buf) must bigger or equal to the number of header bytes
func (headMaker) numericBuf(isNegative bool, length int, buf []byte) int {
	if length <= 0 {
		return 0
	}

	if length <= 8 {
		// single byte header
		if isNegative {
			buf[0] = headerTypeMap[THNegNumSingle].C | (byte(length) & headerTypeMap[THNegNumSingle].W)
		} else {
			buf[0] = headerTypeMap[THPosNumSingle].C | (byte(length) & headerTypeMap[THPosNumSingle].W)
		}
		return 1
	}

	// multi bytes header
	l, _ := Numeric.writeUint(buf[1:], uint64(length))
	if isNegative {
		buf[0] = headerTypeMap[THNegBigInt].C | (byte(l) & headerTypeMap[THNegBigInt].W)
	} else {
		buf[0] = headerTypeMap[THPosBigInt].C | (byte(l) & headerTypeMap[THPosBigInt].W)
	}
	return l + 1
}

func (h headMaker) numeric(isNegative bool, length int) []byte {
	if length <= 0 {
		return nil
	}

	if length <= 8 {
		r := make([]byte, 1)
		h.numericBuf(isNegative, length, r)
		return r
	}

	r := make([]byte, 9)
	l := h.numericBuf(isNegative, length, r)
	return r[:l]
}

// arrayBuf put header of an array into buf, len(buf) must bigger or equal to the number of header bytes
func (headMaker) arrayBuf(length int, buf []byte) int {
	if length <= 0 {
		return 0
	}

	if length <= 16 {
		buf[0] = headerTypeMap[THArraySingle].C | (byte(length) & headerTypeMap[THArraySingle].W)
		return 1
	}

	l, _ := Numeric.writeUint(buf[1:], uint64(length))
	buf[0] = headerTypeMap[THArrayMulti].C | (byte(l) & headerTypeMap[THArrayMulti].W)
	return l + 1
}

func (h headMaker) array(length int) []byte {
	if length <= 0 {
		return nil
	}
	r := make([]byte, 9)
	l := h.arrayBuf(length, r)
	return r[:l]
}

type fieldNames struct {
	index int
	name  string
}

func structFields(typ reflect.Type) (fields []fieldNames, err error) {
	rv, ok := typeInfoMap.Load(typ)
	if ok {
		fields, _ = rv.([]fieldNames)
		return
	}
	for i := 0; i < typ.NumField(); i++ {
		// exported field
		if f := typ.Field(i); f.PkgPath == "" {
			tagStr := f.Tag.Get("rtl")
			// if tagStr == "" {
			// 	// compatible with json tag
			// 	tagStr = f.Tag.Get("json")
			// }
			ignored := false
			for _, tag := range strings.Split(tagStr, ",") {
				switch tag = strings.TrimSpace(tag); tag {
				case "-":
					ignored = true
					break
				}
			}
			if ignored {
				continue
			}
			fields = append(fields, fieldNames{i, f.Name})
		}
	}
	typeInfoMap.Store(typ, fields)
	return fields, nil
}

type StructCodec struct {
	structType reflect.Type
	isPtr      bool
}

func NewStructCodec(typ reflect.Type) (*StructCodec, error) {
	if typ == nil {
		return nil, errors.New("NewStructCodec: struct type should not be nil")
	}
	kind := typ.Kind()
	if kind != reflect.Struct {
		if kind == reflect.Ptr {
			if typ = typ.Elem(); typ.Kind() != reflect.Struct {
				panic("type of value must be a struct of ptr to a struct")
			}
		} else {
			panic("type of value must be a struct of ptr to a struct")
		}
	}
	ret := &StructCodec{structType: typ, isPtr: kind == reflect.Ptr}
	return ret, nil
}

func (c *StructCodec) Encode(o interface{}, w io.Writer) error {
	return Encode(o, w)
}

func (c *StructCodec) Decode(r io.Reader) (interface{}, error) {
	val := reflect.New(c.structType)
	if err := Decode(r, val.Interface()); err != nil {
		return nil, err
	}
	if c.isPtr {
		return val.Interface(), nil
	}
	return val.Elem().Interface(), nil
}
