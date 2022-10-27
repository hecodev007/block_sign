package service

import (
	"bytes"
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/deals"
	user "custody-merchant-admin/model/adminPermission/user"
	"custody-merchant-admin/model/audit"
	"custody-merchant-admin/model/financeFlow"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/order"
	"custody-merchant-admin/model/serviceAuditRole"
	"custody-merchant-admin/model/serviceSecurity"
	"custody-merchant-admin/model/services"
	"custody-merchant-admin/module/blockChainsApi"
	"fmt"
	"time"
)

// CreateOrderInfo
// 创建订单
func CreateOrderInfo(info *domain.OrderInfo, uId int64) error {
	// 查看地址是内部地址还是外部地址,true,false
	var (
		err         error
		orders      []audit.OrderAudit
		auditDao    = new(audit.OrderAudit)
		orderDao    = new(order.Orders)
		sar         = serviceAuditRole.NewEntity()
		security    = serviceSecurity.NewEntity()
		merchantDao = merchant.NewEntity()
		userDao     = user.NewEntity()
		s           = new(services.ServiceAuditConfig)
	)
	nowTime := time.Now().Local()
	startTime := nowTime.Format(global.YyyyMmDdHhMmSs)
	info.CreateUser = uId
	// 创建订单
	orderId, err := orderDao.CreateOrderInfo(info)
	if err != nil {
		return err
	}
	// 创建审核进度
	if orderId != 0 {
		userInfo, err := sar.FindServiceHaveUserBySid(info.ServiceId)
		if err != nil {
			return err
		}
		for i, _ := range userInfo {
			// 判断业务线是否开启了提币门槛
			lid, err := s.GetServiceConfigBySLid(info.ServiceId, userInfo[i].Aid)
			if err != nil {
				return err
			}
			// 为空表示未设置
			if lid != nil && lid.Id == 0 {
				// limit_use 0是打开，1是关闭
				if lid.LimitUse == 0 {
					// 判断开启审核门槛
					orderNums, err := deals.FindLimitNums(startTime, info.ReceiveAddr, info.ServiceId)
					if err != nil {
						return err
					}
					err = deals.ServiceConfigLimit(info.ServiceId, userInfo[i].Aid, info.Nums, orderNums.DNums, orderNums.WNums, orderNums.MNums)
					if err != nil {
						continue
					}
				}
				if lid.State == 1 {
					// 审核制被冻结
					continue
				}
			}
			// 获取业务线安全信息
			err = security.FindItemByBusinessId(int64(info.ServiceId))
			if err != nil {
				return err
			}
			// 平台审核人员
			userUid, err := userDao.GetAdminUserUId(userInfo[i].Uid)
			if err != nil {
				return err
			}
			// 不是平台审核
			if security.IsPlatformCheck == 0 {
				// 跳过平台审核员
				if userUid != nil && userUid.Id != 0 {
					continue
				}
			}
			// 不是商户自行审核
			if security.IsAccountCheck == 0 && (userUid == nil || userUid.Uid != userInfo[i].Uid) {
				merchantUid, err := merchantDao.GetUserMerchantPersonal(userInfo[i].Uid)
				if err != nil {
					return err
				}
				if merchantUid != nil || merchantUid.Id == 0 {
					continue
				}
			}

			// 没开启提币门槛
			orders = append(orders, audit.OrderAudit{
				AuditLevel:  userInfo[i].Aid,
				State:       0,
				AuditResult: 0,
				OrderId:     orderId,
				UserId:      userInfo[i].Uid,
				CreateTime:  time.Now().Local(),
			})
		}
		err = auditDao.BatchCreateAuditInfo(orders)
		if err != nil {
			return err
		}
	} else {
		return global.OperationErrorText(global.OperationAddOrderErr)
	}
	return nil
}

func CheckServiceConfig(sid int) error {
	var serviceDao = new(services.ServiceEntity)
	cInfo, err := serviceDao.GetServiceById(sid)
	if err != nil {
		return err
	}
	if cInfo == nil {
		return global.WarnMsgError(global.MsgWarnNoHaveService)
	}
	// 0是关闭提币
	if cInfo.WithdrawalStatus == 0 {
		return global.WarnMsgError(global.MsgWarnCloseServiceWithdrawal)
	}
	return nil
}

func UpdatePassOrderService(oId, id int64) (int, error) {
	tf, err := UpdatePassOrder(oId, id)
	if err != nil {
		return 0, err
	}
	return tf, nil
}

func UpdateThawOrderService(uInfo *domain.UpdateOrders, id int64) error {
	err := UpdateThawOrder(uInfo, id)
	if err != nil {
		return err
	}
	return nil
}

func UpdateFreezeOrderService(uInfo *domain.UpdateOrders, id int64) error {
	err := UpdateFreezeOrder(uInfo, id)
	if err != nil {
		return err
	}
	return nil
}

func UpdateRefuseOrderService(uInfo *domain.UpdateOrders, id int64) error {
	dao := order.Orders{}
	err := UpdateRefuseOrder(uInfo, id)
	if err != nil {
		return err
	}
	order, err := dao.FindOrderByOId(uInfo.Id)
	if err != nil {
		return global.DaoError(err)
	}
	// 财务流水回滚
	ffInfo := financeFlow.NewEntity()
	err = ffInfo.FindItemByOrderId(order.SerialNo)
	if err != nil {
		return err
	}
	err = RollbackFinanceAssetsByFlowId(ffInfo.Db, int(ffInfo.Id))
	return nil
}

func UpdateOrderAllService(info *domain.SelectOrderInfo, id int64) (int, error) {
	tf, err := UpdateOrderAll(info, id)
	if err != nil {
		return 0, err
	}
	return tf, nil
}

func FindOrderListService(info *domain.SelectOrderInfo, id int64) ([]domain.SelectOrderList, int, error) {

	tf, total, err := FindOrderList(info, id)
	if err != nil {
		return tf, 0, err
	}
	return tf, total, nil
}

func FindOrderExportService(info *domain.SelectOrderInfo, id int64) (bytes.Buffer, error) {
	return FindOrderExport(info, id)
}

func CountOrderStatusService(id int64) ([]domain.CountStatus, error) {

	tf, err := CountOrderStatus(id)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func FindOrderDetailService(id int64) (map[string]interface{}, error) {

	tf, err := FindOrderDetail(id)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

//OrderRollBack 订单回退状态
/*
status 0-不可回滚，1-可回滚
*/
func OrderRollBack(req *domain.OrderOperateReq) (status int, err error) {
	//查询订单
	ssInfo := serviceSecurity.NewEntity()
	ssInfo.FindItemByBusinessId(req.BusinessId)
	if ssInfo.Id == 0 {
		err = fmt.Errorf("业务线不存在")
		return
	}
	if ssInfo.ClientId == "" {
		err = fmt.Errorf("业务线 用户clientId不存在")
		return
	}
	status, err = blockChainsApi.BlockChainOrderUpChainStatus(ssInfo.ClientId, req.OutOrderId, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
	return
}
