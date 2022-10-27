package model

import "reflect"

func MakeEntity(v interface{}) reflect.Type {
	return reflect.TypeOf(v)
}
