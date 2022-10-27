package client

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/flynn/mobilecoin-go/protos"
	"google.golang.org/grpc"
	"strings"
	"time"
)

type GrpcClient struct {
	Address     string
	Conn        *grpc.ClientConn
	Client      protos.MobilecoindAPIClient
	grpcTimeout time.Duration
}

// NewGrpcClient create grpc controller
func NewGrpcClient(address string) *GrpcClient {
	client := &GrpcClient{
		Address:     address,
		grpcTimeout: 10 * time.Second,
	}
	return client
}

// NewGrpcClientWithTimeout create grpc controller
func NewGrpcClientWithTimeout(address string, timeout time.Duration) *GrpcClient {
	client := &GrpcClient{
		Address:     address,
		grpcTimeout: timeout,
	}
	return client
}

// SetTimeout for Client connections
func (g *GrpcClient) SetTimeout(timeout time.Duration) {
	g.grpcTimeout = timeout
}

// Start initiate grpc  connection
func (g *GrpcClient) Start() error {
	var err error
	if len(g.Address) == 0 {
		return errors.New("grpc address is null")
	}
	g.Conn, err = grpc.Dial(g.Address, grpc.WithInsecure())

	if err != nil {
		return fmt.Errorf("Connecting GRPC Client: %v", err)
	}
	g.Client = protos.NewMobilecoindAPIClient(g.Conn)
	return nil
}

// Stop GRPC Connection
func (g *GrpcClient) Stop() {
	if g.Conn != nil {
		g.Conn.Close()
	}
}

// Reconnect GRPC
func (g *GrpcClient) Reconnect(url string) error {
	g.Stop()
	if len(url) > 0 {
		g.Address = url
	}
	g.Start()
	return nil
}
func (g *GrpcClient) isNeedReConnect(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "code = DeadlineExceeded desc = context deadline exceeded") {
		return true
	}
	return false
}
