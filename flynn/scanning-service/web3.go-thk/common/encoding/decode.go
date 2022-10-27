package encoding

import (
	"bytes"
	"io"
	"reflect"

	"github.com/golang/protobuf/proto"
)

func UnmarshalObj(buf []byte, o interface{}) (err error) {
	v, ok := o.(proto.Message)
	if ok {
		// protobuf
		err = proto.Unmarshal(buf, v)
	} else {
		// decoding
		err = Unmarshal(buf, o)
	}
	return
}

func Unmarshal(buf []byte, v interface{}) error {
	return Decode(NewValueReader(bytes.NewBuffer(buf), 0), v)
}

// Decode reads bytes from r unmarshal to v, if you want use same Reader to Decode multi
// values, you should use encoding.ValueReader as io.Reader.
func Decode(r io.Reader, v interface{}) error {
	if v == nil {
		return ErrDecodeIntoNil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return ErrDecodeNoPtr
	}
	if rv.IsNil() {
		return ErrDecodeIntoNil
	}

	isDecoder, err := checkTypeOfDecoder(r, rv)
	if isDecoder || err != nil {
		return err
	}

	rev := rv.Elem()
	vr, ok := r.(ValueReader)
	if !ok {
		vr = NewValueReader(r, 0)
	}
	if err := valueReader(vr, rev); err != nil {
		return err
	}
	return nil
}

func checkTypeOfDecoder(r io.Reader, value reflect.Value) (isDecoder bool, err error) {
	typ := value.Type()

	if typ.Implements(TypeOfDecoder) {
		isDecoder = true
		if typ.Kind() == reflect.Ptr && value.IsNil() {
			nvalue := reflect.New(typ.Elem())
			value.Set(nvalue)
		}
		decoder, _ := value.Interface().(Decoder)
		err = decoder.Deserialization(r)
		return
	}

	if typ.Kind() == reflect.Ptr {
		etyp := typ.Elem()
		if etyp.Implements(TypeOfDecoder) {
			isDecoder = true
			elem := value.Elem()
			if elem.Kind() == reflect.Ptr && elem.IsNil() {
				evalue := reflect.New(etyp.Elem())
				elem.Set(evalue)
			}
			decoder, _ := elem.Interface().(Decoder)
			err = decoder.Deserialization(r)
			return
		}
	}

	return
}

func DecodeBigInt(r io.Reader, v interface{}) error {
	typ := reflect.TypeOf(v)
	if !typ.AssignableTo(typeOfBigInt) && !typ.AssignableTo(reflect.PtrTo(typeOfBigInt)) {
		return ErrUnsupported
	}
	vr, ok := r.(ValueReader)
	if !ok {
		vr = NewValueReader(r, 0)
	}
	value := reflect.ValueOf(v)
	th, length, err := vr.ReadHeader()
	if err != nil {
		return nil
	}
	return bigIntReader0(th, int(length), vr, value, 0)
}
