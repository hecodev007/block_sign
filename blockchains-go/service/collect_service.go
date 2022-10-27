package service

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/robfig/cron/v3"
)

type Collector interface {
	//获取服务币种名称
	Name() string
	Spec() string
	cron.Job
}

type CollectService struct {
	cfg conf.CollectConfig
}

//func NewCollectService(cfg conf.CollectConfig) *CollectService {
//	return &CollectService{
//		cfg,
//	}
//}
//
//func (s *CollectService) Name() string {
//	return s.cfg.Name
//}
//
//func (s *CollectService) Spec() string {
//	return s.cfg.Spec
//}
//
//func (s *CollectService) Run() {
//	//查找币种信息
//	coinSet, err := dao.FcCoinSetGetByName(s.cfg.Name, 1)
//	if err != nil {
//		log.Errorf("find coin err %v", err)
//		return
//	}
//	//查找需要归集的商户信息
//	mchSet, err := dao.FcMchFindByPlatformsAndStatus(2, s.cfg.Platforms)
//	if err != nil {
//		log.Errorf("find platforms err %v", err)
//		return
//	}
//
//	//轮训每个商户需要归集的订单
//	for _, mch := range mchSet {
//		//查找该商户下归集地址
//		if toAddresses, err := dao.FcGenerateAddressListFindAddresses(1, 2, mch.Id, s.cfg.Name); err == nil {
//			//查找有需要进行归集的地址
//			if fromAddresses, err := dao.FcAddressAmountFindAddresses(1, 2, mch.Id, s.cfg.Name, fmt.Sprintf("%f", s.cfg.MinAmount)); err == nil {
//				for i, fromAddress := range fromAddresses {
//					//生成归集订单
//					applyOrder := &entity.FcTransfersApply{
//						Username:   "Robot",
//						Department: "冷钱包组",
//						Applicant:  mch.Platform,
//						Operator:   "Robot",
//						Type:       "gj",
//						Purpose:    "自动归集",
//						CallBack:   "",
//						OutOrderid: fmt.Sprintf("clt_%s_%d_%d", s.cfg.Name, time.Now().Unix(), i),
//						AppId:      mch.Id,
//					}
//
//					if coinSet.Type != 2 {
//						applyOrder.Eoskey = coinSet.Name
//						applyOrder.Eostoken = coinSet.Token
//					}
//
//					fromTA := &entity.FcTransfersApplyCoinAddress{
//						Address:     fromAddress,
//						AddressFlag: "from",
//						Status:      0,
//					}
//
//					toTA := &entity.FcTransfersApplyCoinAddress{
//						Address:     toAddresses[0],
//						AddressFlag: "to",
//						Status:      0,
//					}
//
//					dao.FcTransfersApplyCreate(applyOrder, []*entity.FcTransfersApplyCoinAddress{fromTA, toTA})
//				}
//			}
//		}
//	}
//}
