package abi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

type Method struct {
	Name    string
	Const   bool
	Inputs  Arguments
	Outputs Arguments
}

func (method Method) Sig() string {
	types := make([]string, len(method.Inputs))
	for i, input := range method.Inputs {
		types[i] = input.Type.String()
	}
	return fmt.Sprintf("%v(%v)", method.Name, strings.Join(types, ","))
}

func (method Method) String() string {
	inputs := make([]string, len(method.Inputs))
	for i, input := range method.Inputs {
		inputs[i] = fmt.Sprintf("%v %v", input.Type, input.Name)
	}
	outputs := make([]string, len(method.Outputs))
	for i, output := range method.Outputs {
		outputs[i] = output.Type.String()
		if len(output.Name) > 0 {
			outputs[i] += fmt.Sprintf(" %v", output.Name)
		}
	}
	constant := ""
	if method.Const {
		constant = "constant "
	}
	return fmt.Sprintf("function %v(%v) %sreturns(%v)", method.Name, strings.Join(inputs, ", "), constant, strings.Join(outputs, ", "))
}

func (method Method) Id() []byte {
	return crypto.Keccak256([]byte(method.Sig()))[:4]
}

func (method Method) singleInputUnpack(v interface{}, input []byte) error {

	valueOf := reflect.ValueOf(v)
	if reflect.Ptr != valueOf.Kind() {

		s := fmt.Sprintf("abi: Unpack(non-pointer %T)", v)
		return errors.New(s)
	}

	value := valueOf.Elem()
	marshalledValue, err := toGoType(0, method.Inputs[0].Type, input)
	if err != nil {
		return err
	}

	if err := myset(value, reflect.ValueOf(marshalledValue), method.Inputs[0]); err != nil {
		return err
	}

	return nil
}

func (method Method) multInputUnpack(v []interface{}, input []byte) error {

	j := 0
	for i := 0; i < len(method.Inputs); i++ {

		// v[i]必须是指针类型
		valueOf := reflect.ValueOf(v[i])
		if reflect.Ptr != valueOf.Kind() {

			s := fmt.Sprintf("abi: Unpack(non-pointer %T)", v)
			return errors.New(s)
		}

		toUnpack := method.Inputs[i]
		if toUnpack.Type.T == ArrayTy {
			j += toUnpack.Type.Size
		}

		marshalledValue, err := toGoType((i+j)*32, toUnpack.Type, input)
		if err != nil {
			return err
		}

		if err := myset(valueOf.Elem(), reflect.ValueOf(marshalledValue), method.Inputs[i]); err != nil {
			return err
		}
	}
	return nil
}
