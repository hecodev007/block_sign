package v1

//
//import (
//	"github.com/group-coldwallet/wallet-sign/conf"
//	"github.com/group-coldwallet/wallet-sign/model"
//	"github.com/group-coldwallet/wallet-sign/util"
//)
//
///*
//service模板
//*/
//
///*
//币种服务结构体
//*/
//type TempService struct {
//	*BaseService
//}
///*
//初始化币种服务
//	注意：
//		方法接受者： BaseService
//		方法命名： 币种大写 + Service
//*/
//func (bs *BaseService)TEMPService()*TempService{
//	tp:=new(TempService)
//	tp.BaseService = bs
//	//初始化连接
//	return tp
//}
//
///*
//接口创建地址服务
//	无需改动
//*/
//func (tp *TempService)CreateAddressService(req *model.ReqCreateAddressParams)(*model.RespCreateAddressParams, error){
//	if conf.Config.IsStartThread {
//		return tp.BaseService.multiThreadCreateAddress(req.Num,req.CoinName,req.MchId,req.OrderId,tp.createAddressInfo)
//	}
//	return tp.BaseService.createAddress(req,tp.createAddressInfo)
//}
//
///*
//离线创建地址服务，通过多线程创建
//	无需改动
//*/
//func (tp *TempService)MultiThreadCreateAddrService(nums int,coinName,mchId,orderId string)error{
//	_,err:= tp.BaseService.multiThreadCreateAddress(nums,coinName,mchId,orderId,tp.createAddressInfo)
//	return err
//}
///*
//签名服务
//*/
//func (tp *TempService)SignService(req *model.ReqSignParams)(interface{},error){
//	return nil,nil
//}
//
///*
//热钱包出账服务
//*/
//func (tp *TempService)TransferService(req interface{})(interface{},error){
//	return nil,nil
//}
//
///*
//创建地址实体方法
//*/
//func (tp *TempService)createAddressInfo()(util.AddrInfo, error){
//	return util.AddrInfo{},nil
//}

//func (tp *TempService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
//	return nil, errors.New("unsupport it")
//}
//func (tp *TempService) ValidAddress(address string) error {
//	return errors.New("unsupport it")
//}
