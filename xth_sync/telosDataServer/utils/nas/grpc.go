package nas

//
//import (
//	"context"
//	"github.com/nebulasio/go-nebulas/rpc"
//	"github.com/nebulasio/go-nebulas/rpc/pb"
//)
//
//type NasGrpcClient struct {
//	Client  rpcpb.ApiServiceClient
//	NodeURI string
//}
//
//func NewNasGrpcClient(url string) (*NasGrpcClient, error) {
//	conn, err := rpc.Dial(url)
//	if err != nil {
//		return nil, err
//	}
//
//	client := rpcpb.NewApiServiceClient(conn)
//
//	return &NasGrpcClient{
//		Client:  client,
//		NodeURI: url,
//	}, nil
//}
//
//func (c *NasGrpcClient) GetNebState() (*NodeState, error) {
//	res, err := c.Client.GetNebState(context.Background(), &rpcpb.NonParamsRequest{})
//	if err != nil {
//		return nil, err
//	}
//
//	return toNodeState(res), nil
//}
//
//func (c *NasGrpcClient) LatestIrreversibleBlock() (*Block, error) {
//	res, err := c.Client.LatestIrreversibleBlock(context.Background(), &rpcpb.NonParamsRequest{})
//	if err != nil {
//		return nil, err
//	}
//	return toBlock(res), nil
//}
//
//func (c *NasGrpcClient) GetBlockByHeight(height uint64, full bool) (*Block, error) {
//	res, err := c.Client.GetBlockByHeight(context.Background(), &rpcpb.GetBlockByHeightRequest{Height: height, FullFillTransaction: full})
//	if err != nil {
//		return nil, err
//	}
//	return toBlock(res), nil
//}
//
//func (c *NasGrpcClient) GetBlockByHash(hash string, full bool) (*Block, error) {
//	res, err := c.Client.GetBlockByHash(context.Background(), &rpcpb.GetBlockByHashRequest{Hash: hash, FullFillTransaction: full})
//	if err != nil {
//		return nil, err
//	}
//	return toBlock(res), nil
//}
//
//func (c *NasGrpcClient) GetTransactionReceipt(txid string) (*Transaction, error) {
//	res, err := c.Client.GetTransactionReceipt(context.Background(), &rpcpb.GetTransactionByHashRequest{Hash: txid})
//	if err != nil {
//		return nil, err
//	}
//	return toTransaction(res), nil
//}
//
//func (c *NasGrpcClient) GetEventsByHash() (*rpcpb.EventsResponse, error) {
//	return c.Client.GetEventsByHash(context.Background(), &rpcpb.HashRequest{})
//}
//*/
