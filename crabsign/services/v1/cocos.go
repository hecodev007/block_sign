package v1

//
//import (
//	"errors"
//	"fmt"
//	"github.com/coldwallet-group/substrate-go/rpc"
//	"wallet-sign/conf"
//	"wallet-sign/model"
//	"wallet-sign/util"
//	"github.com/shopspring/decimal"
//)
//
///*
//service模板
//*/
//
///*
//币种服务结构体
//*/
//type CocosService struct {
//	*BaseService
//	client   *rpc.Client
//	url      string
//}
//
///*
//初始化币种服务
//	注意：
//		方法接受者： BaseService
//		方法命名： 币种大写 + Service
//*/
//func (bs *BaseService) COCOSService() *CocosService {
//	ks := new(CocosService)
//	ks.BaseService = bs
//	ks.url = conf.Config.CocosCfg.NodeUrl
//	return ks
//}
//
///*
//接口创建地址服务
//	无需改动
//*/
//func (ks *CocosService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
//	if conf.Config.IsStartThread {
//		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
//	}
//	return ks.BaseService.createAddress(req, ks.createAddressInfo)
//}
//
///*
//离线创建地址服务，通过多线程创建
//	无需改动
//*/
//func (ks *CocosService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
//	fmt.Println("start create Cocos address")
//	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
//	return err
//}
//
///*
//签名服务
//*/
//func (ks *CocosService) SignService(req *model.ReqSignParams) (interface{}, error) {
//	return nil,errors.New("unsopport")
//}
//
///*
//热钱包出账服务
//*/
//func (ks *CocosService) TransferService(req interface{}) (interface{}, error) {
//
//	var tp model.CocosTransferParams
//	if err := ks.BaseService.parseData(req, &tp); err != nil {
//
//		return nil, err
//	}
//	if &tp == nil {
//		return nil, errors.New("transfer params is null")
//	}
//	if tp.FromAddress == "" || tp.ToAddress == "" || tp.ToAmount == decimal.Zero {
//		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.ToAmount.String())
//	}
//	//判断金额是否足够出账
//	// 获取链上余额
//
//}
//
///*
//创建地址实体方法
//*/
//func (ks *CocosService) createAddressInfo() (util.AddrInfo, error) {
//	var addrInfo util.AddrInfo
//	return addrInfo, nil
//}
