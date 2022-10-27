package encoding

import (
	"bytes"
	"io"
	"reflect"

	"github.com/golang/protobuf/proto"
)

func MarshalObj(o interface{}) (body []byte, err error) {
	v, ok := o.(proto.Message)
	if ok {
		// protobuf
		body, err = proto.Marshal(v)
	} else {
		// encoding
		body, err = Marshal(o)
	}
	return
}

func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	var err error
	err = Encode(v, buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodeObj(o interface{}, w io.Writer) error {
	v, ok := o.(proto.Message)
	if ok {
		// protobuf
		body, err := proto.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(body)
		if err != nil {
			return err
		}
		return nil
	} else {
		// encoding
		return Encode(o, w)
	}
}

func Encode(v interface{}, w io.Writer) error {
	// typ := reflect.TypeOf(v)
	// if typ.Implements(TypeOfEncoder) {
	// 	// if the object can be serialized by itself
	// encoder, _ := v.(Encoder)
	// err := encoder.Serialization(w)
	// return err
	// }

	vv := reflect.ValueOf(v)
	_, err := valueWriter(w, vv)
	return err
}

func EncodeBigInt(v interface{}, w io.Writer) error {
	value := reflect.ValueOf(v)
	typ := value.Type()
	for typ.Kind() == reflect.Ptr {
		value = value.Elem()
		typ = value.Type()
	}
	if !typ.AssignableTo(typeOfBigInt) {
		return ErrUnsupported
	}
	_, err := bigIntWriter(w, value)
	return err
}
