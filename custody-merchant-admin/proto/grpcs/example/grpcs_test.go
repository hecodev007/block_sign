package example

import (
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/proto/grpcs"
	"fmt"
	"google.golang.org/grpc/credentials"
	"reflect"
	"testing"
)

type KeyData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestRpc1(t *testing.T) {
	myType := &grpcs.DealService{}
	mtV := reflect.ValueOf(&myType).Elem()
	call := mtV.MethodByName("StringService")
	if call.Kind() == reflect.Func {
		fmt.Println("true")
	}
	methods := call.Call(nil)
	str := methods[0].Interface().(string)
	fmt.Println(str)
	err := methods[1].Interface()
	if err != nil {
		fmt.Println(err.(error))
	}
}

func TestServer(t *testing.T) {
	grpcs.ConnServiceConfig().StartServer()
}

func TestClient(t *testing.T) {

	_, err := credentials.NewClientTLSFromFile("../tls/server.pem", "")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
	}

	// 连接服务器
	// conn, err := grpc.Dial(Address, grpc.WithTransportCredentials(reds), grpc.WithPerRPCCredentials(&token))
	//
	kd := KeyData{Name: "myPost", Age: 12}
	kds := KeyData{}
	grpcs.ConnClientConfig().NweGrpcClient("SendMessage", kd, &kds)
	kd.Name = "post"
	grpcs.CGCf.NweGrpcClient("SendMessage", kd, &kds)
	fmt.Println(kds)
}
