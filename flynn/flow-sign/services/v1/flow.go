package v1

import (
	"github.com/group-coldwallet/flow-sign/common"
	"github.com/group-coldwallet/flow-sign/conf"
	"github.com/group-coldwallet/flow-sign/model"
	flowClient "github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

type FlowService struct {
	*BaseService
	client *flowClient.Client
}

func (bs *BaseService) FLOWService() *FlowService {
	cs := new(FlowService)
	cs.BaseService = bs
	//初始化连接
	cs.client, _ = flowClient.New(conf.Config.FlowCfg.NodeUrl, grpc.WithInsecure())
	return cs
}

func (cs *FlowService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return cs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, cs.createAddressInfo)
	}
	return cs.BaseService.createAddress(req, cs.createAddressInfo)
}
func (cs *FlowService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {

	_, err := cs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, cs.createAddressInfo)
	return err
}

func (cs *FlowService) createAddressInfo() (common.AddrInfo, error) {

	var (
		addrInfo common.AddrInfo
	)

	return addrInfo, nil
}

func (cs *FlowService) ValidAddress(address string) error {

	return nil
}

func (cs *FlowService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {

	return nil, nil
}

func (cs *FlowService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, nil
}

/*
热钱包出账服务
*/
func (cs *FlowService) TransferService(req interface{}) (interface{}, error) {
	return nil, nil
}
