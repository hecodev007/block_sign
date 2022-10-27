package grpcs

import (
	"context"
	"reflect"
)

type GrpcService struct{}

func (g *GrpcService) SendRequest(ctx context.Context, params *ParamRequest) (*ParamReply, error) {
	values := CallMethodByName(params)
	if values != nil {
		v := values.(*ParamReply)
		return v, nil
	}
	return nil, nil
}

func CallMethodByName(param *ParamRequest) interface{} {
	// 要获取的结构体
	myType := &DealService{}
	// 获取结构体内部信息
	mtV := reflect.ValueOf(&myType).Elem()
	// 根据名称反射获取结构体内的方法
	call := mtV.MethodByName(param.Method)
	// 判断方法是否存在
	if call.Kind() != reflect.Func {
		return nil
	}
	// 传入一个参数
	params := make([]reflect.Value, 1)
	params[0] = reflect.ValueOf(param)
	// 调用该方法
	values := call.Call(params)
	// 返回方法的处理结果
	return values[0].Interface()
}
