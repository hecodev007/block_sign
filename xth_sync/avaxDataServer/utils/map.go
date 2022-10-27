package utils

import (
	"github.com/goinggo/mapstructure"
	"reflect"
)

func StructToMap2(o interface{}) map[string]interface{} {
	mapInstance := make(map[string]interface{})
	err := mapstructure.Decode(mapInstance, o)
	if err != nil {
		return nil
	}
	return mapInstance
}

func StructToMap(o interface{}) map[string]interface{} {
	obj1 := reflect.TypeOf(o)
	obj2 := reflect.ValueOf(o)
	if obj1.Kind() != reflect.Struct {
		return nil
	}
	var data = make(map[string]interface{})
	for i := 0; i < obj1.NumField(); i++ {
		data[obj1.Field(i).Name] = obj2.Field(i).Interface()
	}
	return data
}
