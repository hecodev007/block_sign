## 操作：

编译执行：

```shell
 protoc -I proto/  hello.proto --go_out=plugins=grpc:proto/hello
```

该模块的是通过函数名反射函数方法来进行动态匹配动态调用函数处理的目的

使用案例：grpcs/example/grpcs_test.go

```go

func TestServer(t *testing.T) {
	grpcs.StartServer()
}

func TestClient(t *testing.T) {
	kd := KeyData{Name: "myPost", Age: 12}
	kds := KeyData{}
	grpcs.NweGrpcClient("SendMessage", kd, &kds)
	fmt.Println(kds)
}

```


./services 目录下的服务接口

1. 其中：util.Deserialize 函数是必须引入的反序例化方法，用于解析客户端序列化的数据
2. 返回结果的格式必须为: *post.BodyReply/*get.ResultReply

例如：
```go

func (receiver *PostService) SendMessage(body *post.BodyRequest) *post.BodyReply {
	name := KeyData{}
	reply := &post.BodyReply{
		Code: 200,
		Msg:  "已经收到消息",
	}
	err := util.Deserialize(body.RpcBody["key_data"], &name)
	if err != nil {
		reply.Code = 400
		reply.Msg = err.Error()
		return reply
	}
	send := map[string][]byte{}
	name.Name = "123"
	name.Age = 123
	data, err := util.Serialize(name)
	if err != nil {
		reply.Code = 500
		reply.Msg = err.Error()
		return reply
	}
	send["data"] = data
	reply.Code = 200
	reply.Msg = "已经收到消息"
	reply.RpcReply = send
	return reply
}

```


