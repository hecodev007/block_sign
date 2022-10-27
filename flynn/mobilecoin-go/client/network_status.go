package client

import (
	"context"
	"fmt"
	"github.com/group-coldwallet/flynn/mobilecoin-go/protos"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (g *GrpcClient) GetNetworkStatus() (*protos.GetNetworkStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.grpcTimeout)
	defer cancel()
	result, err := g.Client.GetNetworkStatus(ctx, new(emptypb.Empty))
	if g.isNeedReConnect(err) {
		return g.Client.GetNetworkStatus(ctx, new(emptypb.Empty))
	}
	if err != nil {
		return nil, fmt.Errorf("get network status error: %v", err)
	}
	return result, err
}
