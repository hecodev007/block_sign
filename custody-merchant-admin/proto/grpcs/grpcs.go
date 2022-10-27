package grpcs

import (
	"context"
	. "custody-merchant-admin/config"
	"custody-merchant-admin/module/log"
	grpc_mw2 "custody-merchant-admin/proto/grpc_mw"
	"custody-merchant-admin/util"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

var SGCf *GrpcConfig
var CGCf *GrpcConfig

type GrpcConfig struct {
	Host string
	Port string
}

func ConnClientConfig() *GrpcConfig {
	client := Conf.Grpc["client"]
	if CGCf == nil {
		CGCf = &GrpcConfig{
			Host: client.Host,
			Port: client.Port,
		}
	}
	return CGCf
}

func ConnServiceConfig() *GrpcConfig {
	if SGCf == nil {
		server := Conf.Grpc["server"]
		SGCf = &GrpcConfig{
			Host: server.Host,
			Port: server.Port,
		}
	}
	return SGCf
}

func (g *GrpcConfig) StartServer() {
	// 监听本地的8972端口
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", g.Host, g.Port))
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	// 创建gRPC服务器
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(grpc_mw2.ZapInterceptor()),
			grpc_auth.StreamServerInterceptor(grpc_mw2.AuthInterceptor),
			grpc_recovery.StreamServerInterceptor(grpc_mw2.RecoveryInterceptor()),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(grpc_mw2.ZapInterceptor()),
			grpc_auth.UnaryServerInterceptor(grpc_mw2.AuthInterceptor),
			grpc_recovery.UnaryServerInterceptor(grpc_mw2.RecoveryInterceptor()),
		)),
	)
	// 在gRPC服务端注册服务
	RegisterGreeterServer(s, &GrpcService{})
	//在给定的gRPC服务器上注册服务器反射服务
	reflection.Register(s)
	// Serve方法在lis上接受传入连接，为每个连接创建一个ServerTransport和server的goroutine。
	// 该goroutine读取gRPC请求，然后调用已注册的处理程序来响应它们。
	err = s.Serve(lis)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}

func (g *GrpcConfig) NweGrpcClient(method string, v interface{}, reply interface{}) {
	// 连接服务器
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", g.Host, g.Port), grpc.WithInsecure())
	if err != nil {
		fmt.Printf("faild to connect: %v", err)
	}
	defer conn.Close()
	mb := map[string][]byte{}
	serialize, err := util.Serialize(v)
	if err != nil {
		log.Errorf("grpc 客户端序列化出错: %v ", err)
		return
	}
	// 序列化
	mb["key_data"] = serialize
	// 调用服务端的SayHello
	timeStamp := time.Now().Unix()
	c := NewGreeterClient(conn)
	r, err := c.SendRequest(context.Background(), &ParamRequest{
		Method:    method,
		Params:    mb,
		TimeStamp: timeStamp,
	})
	if err != nil {
		log.Errorf("grpc post客户端出错: could not greet: %v", err)
		return
	}
	// 反序列化
	err = util.Deserialize(r.RpcReply["data"], reply)
	if err != nil {
		log.Errorf("grpc 客户端反序列化出错: %v ", err)
		return
	}
}
