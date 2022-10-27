package service

import (
	"bytes"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/order"
	"custody-merchant-admin/model/orderAudit"
	"custody-merchant-admin/model/serviceAuditRole"
	"custody-merchant-admin/model/services"
	"custody-merchant-admin/module/dict"
	"errors"
	"github.com/tealeg/xlsx"
	"strconv"
	"strings"
	"time"
)

func FindOrderDetail(oId int64) (map[string]interface{}, error) {
	var (
		dao      = order.Orders{}
		s        = services.ServiceEntity{}
		auditDao = orderAudit.NewEntity()
		merchant = merchant.NewEntity()
	)

	var res = map[string]interface{}{
		"list":         [][]domain.AuditDetail{},
		"service_name": "",
		"audit_type":   0,
		"audit_name":   dict.AuditTypeList[0],
	}

	order, err2 := dao.FindOrderByOId(oId)
	if err2 != nil {
		return res, err2
	}
	if order == nil {
		return res, err2
	}
	service, err := s.GetServiceById(order.ServiceId)
	if err != nil {
		return res, err
	}
	if service == nil {
		return res, err2
	}

	info, err := auditDao.FindOrderDetailByASC(oId)
	if err != nil {
		return res, global.DaoError(err)
	}
	nullInfo, err := auditDao.FindOrderDetailByNull(oId)
	if err != nil {
		return res, global.DaoError(err)
	}
	if nullInfo != nil {
		info = append(info, nullInfo...)
	}
	superInfo, err := auditDao.FindOrderDetailBySuper(oId)
	if err != nil {
		return res, global.DaoError(err)
	}
	var (
		sList       = []domain.UserAudit{}
		orderDetail = []domain.OrderDetail{}
		nulltype    = 0
	)
	bstate := 0
	edits := true
	// 遍历数据库里的审核进度
	for i := 0; i < len(info); i++ {
		personal, err := merchant.GetUserMerchantPersonal(info[i].UserId)
		if err != nil {
			return res, err
		}
		if personal == nil {
			continue
		}
		audit := domain.UserAudit{
			AuditResult: info[i].AuditResult,
			ResultName:  dict.OrderResult[info[i].AuditResult],
			AuditLevel:  info[i].AuditLevel,
			LevelName:   dict.AuditLevelName[info[i].AuditLevel-1],
			UserId:      info[i].UserId,
			UserName:    personal.Name,
		}
		if i < service.AuditType {
			if !info[i].UpdateTime.IsZero() {
				// 有更新情况，则为已审核操作过
				icon := ""
				if v, ok := dict.IconMap[dict.OrderResult[info[i].AuditResult]]; ok {
					mps := v.(map[string]string)
					icon = mps["icon"]
				}

				od := domain.OrderDetail{
					Icon:       icon,
					UpDateTime: info[i].UpdateTime.Local().Format(global.YyyyMmDdHhMmSs),
					State:      info[i].AuditResult,
					TypeName:   dict.AuditUserList[i],
					StatusName: dict.OrderResult[info[i].AuditResult],
					UserList:   append([]domain.UserAudit{}, audit),
				}
				// 添加为审核进度
				orderDetail = append(orderDetail, od)
			} else {
				// 未审核过
				if nulltype == 0 && edits {
					nulltype = i
					edits = false
				}
				if audit.AuditResult != 0 {
					bstate = audit.AuditResult
				}
				// 未审核过列表
				sList = append(sList, audit)
			}
		}

		// 超出审核制度
		//if i > service.AuditType {
		//	if audit.AuditResult != 0 {
		//		bstate = audit.AuditResult
		//	}
		//	// 放入下一个审核进度里
		//	sList = append(sList, audit)
		//}

	}

	if len(sList) != 0 {
		od := domain.OrderDetail{
			State:      bstate,
			TypeName:   dict.AuditUserList[nulltype],
			UserList:   sList,
			StatusName: dict.OrderResult[bstate],
		}
		orderDetail = append(orderDetail, od)
		nulltype++
	}

	if nulltype != 0 {
		if service != nil {
			for i := nulltype; i < service.AuditType; i++ {
				od := domain.OrderDetail{
					UpDateTime: "",
					State:      0,
					TypeName:   dict.AuditUserList[i],
					UserList:   []domain.UserAudit{},
					StatusName: dict.OrderResult[0],
				}
				orderDetail = append(orderDetail, od)
			}
		} else {
			return res, err
		}
	}

	// 没有审核员
	if nulltype == 0 && len(info) == 0 {
		for i := 0; i < service.AuditType; i++ {
			od := domain.OrderDetail{
				UpDateTime: "",
				State:      0,
				TypeName:   dict.AuditUserList[i],
				UserList:   []domain.UserAudit{},
				StatusName: dict.OrderResult[0],
			}
			orderDetail = append(orderDetail, od)
		}
	}

	state := 0
	updateTime := ""
	sList = []domain.UserAudit{}
	passList := []domain.UserAudit{}

	for i := 0; i < len(superInfo); i++ {
		personal, err := merchant.GetUserMerchantPersonal(info[i].UserId)
		if err != nil {
			return res, err
		}
		if personal == nil {
			continue
		}
		audit := domain.UserAudit{
			AuditResult: superInfo[i].AuditResult,
			ResultName:  dict.OrderResult[superInfo[i].AuditResult],
			AuditLevel:  superInfo[i].AuditLevel,
			LevelName:   dict.AuditLevelName[superInfo[i].AuditLevel-1],
			UserId:      superInfo[i].UserId,
			UserName:    personal.Name,
		}

		if superInfo[i].AuditResult != 0 && state != 1 && state != 4 {
			updateTime = superInfo[i].UpdateTime.Local().Format(global.YyyyMmDdHhMmSs)
			state = superInfo[i].AuditResult
			passList = append(passList, audit)
		}
		sList = append(sList, audit)
	}

	icon := ""
	if v, ok := dict.IconMap[dict.OrderResult[state]]; ok {
		mps := v.(map[string]string)
		icon = mps["icon"]
	}
	if len(passList) != 0 {
		sList = passList
	}
	od := domain.OrderDetail{
		Icon:       icon,
		UpDateTime: updateTime,
		State:      state,
		StatusName: dict.OrderResult[state],
		TypeName:   dict.AuditUserList[5],
		UserList:   sList,
	}

	orderDetail = append(orderDetail, od)

	if len(orderDetail) != 0 {
		orders := orderDetail[len(orderDetail)-1]
		if orders.State == 1 || orders.State == 4 {
			for i, _ := range orderDetail {
				if orderDetail[i].State == 0 {
					orderDetail[i].UserList = []domain.UserAudit{}
				}
			}
		}
	}

	return map[string]interface{}{
		"list":         orderDetail,
		"service_name": service.Name,
		"audit_type":   service.AuditType,
		"audit_name":   dict.AuditTypeList[service.AuditType-1],
	}, nil
}

func CountOrderStatus(uId int64) ([]domain.CountStatus, error) {
	var (
		dao    = order.Orders{}
		status = []domain.CountStatus{}
		total  int
	)

	info, err := dao.CountOrderStatus(uId)

	if err != nil {
		return status, global.DaoError(err)
	}

	status = append(status, domain.CountStatus{
		Count:     total,
		State:     -1,
		StateName: "全部订单",
	})

	status = append(status, domain.CountStatus{
		Count:     0,
		State:     0,
		StateName: "待审核",
	})

	status = append(status, domain.CountStatus{
		Count:     0,
		State:     1,
		StateName: "已通过",
	})

	status = append(status, domain.CountStatus{
		Count:     0,
		State:     2,
		StateName: "冻结",
	})

	status = append(status, domain.CountStatus{
		Count:     0,
		State:     4,
		StateName: "拒绝",
	})

	l := len(info)
	for i := 0; i < l; i++ {
		total += info[i].Count
		index := info[i].OrderResult

		if info[i].OrderResult == 4 {
			index = 4
		} else {
			index = info[i].OrderResult + 1
		}
		buf := status[index]
		buf.Count = info[i].Count
		status[index] = buf
	}

	status[0].Count = total
	return status, nil
}

// UpdateThawOrder
// 解冻订单
func UpdateThawOrder(uInfo *domain.UpdateOrders, uId int64) error {
	sarDao := serviceAuditRole.NewEntity()
	auditDao := orderAudit.NewEntity()
	dao := order.Orders{}
	order, err := dao.FindOrderByOId(uInfo.Id)
	if err != nil {
		return global.DaoError(err)
	}
	sar, err := sarDao.FindSARByUIdAndSId(uId, order.ServiceId)
	if err != nil {
		return err
	}
	if sar.Aid != 4 {
		return global.OperationError(global.OperationUserNotSuperAudit)
	}
	// 订单处于冻结，操作用户为超级审
	if order.OrderResult == 2 {
		// 该用户的层级审核
		_, err = auditDao.UpdateAuditInfo(order.Id, uId, map[string]interface{}{
			"audit_result": 3,
			"update_time":  time.Now().Local(),
		})
		// 解冻
		_, err3 := dao.UpdateOrdersInfo(order.Id, map[string]interface{}{
			"order_result": 0,
			"reason":       uInfo.Reason,
		})
		if err3 != nil {
			return global.DaoError(err3)
		}
	} else {
		return global.OperationError(global.OperationUpdateThawOrderErr)
	}
	return nil
}

// UpdateFreezeOrder
// 冻结订单
func UpdateFreezeOrder(uInfo *domain.UpdateOrders, uId int64) error {
	dao := order.Orders{}
	auditDao := orderAudit.NewEntity()
	order, err := dao.FindOrderByOId(uInfo.Id)
	if err != nil {
		return global.DaoError(err)
	}
	err = optionServiceState(order.ServiceId, uId)
	if err != nil {
		return err
	}
	// 订单处于拒绝或者通过
	if order.OrderResult != 0 {
		return global.OperationError(global.OperationUpdateNormalOrderErr)
	}
	// 该用户的层级审核
	_, err = auditDao.UpdateAuditInfo(order.Id, uId, map[string]interface{}{
		"audit_result": 2,
		"update_time":  time.Now().Local(),
	})
	// 冻结
	_, err3 := dao.UpdateOrdersInfo(order.Id, map[string]interface{}{
		"order_result": 2,
		"reason":       uInfo.Reason,
	})
	if err3 != nil {
		return global.DaoError(err3)
	}
	return nil
}

// UpdateRefuseOrder
// 拒绝订单
func UpdateRefuseOrder(uInfo *domain.UpdateOrders, uId int64) error {
	dao := order.Orders{}
	auditDao := orderAudit.NewEntity()

	order, err := dao.FindOrderByOId(uInfo.Id)
	if err != nil {
		return global.DaoError(err)
	}
	err = optionServiceState(order.ServiceId, uId)
	if err != nil {
		return err
	}
	// 订单处于拒绝或者通过
	if order.OrderResult != 0 {
		return global.WarnMsgError(global.OperationUpdateNormalOrderErr)
	}
	// 该用户的层级审核
	_, err = auditDao.UpdateAuditInfo(order.Id, uId, map[string]interface{}{
		"audit_result": 4,
		"update_time":  time.Now().Local(),
	})
	_, err3 := dao.UpdateOrdersInfo(order.Id, map[string]interface{}{
		"order_result": 4,
		"reason":       uInfo.Reason,
	})
	if err3 != nil {
		return global.DaoError(err3)
	}
	// 拒绝订单，回滚
	return deals.RollbackBillAssets(order.SerialNo)
}

func UpdatePassOrder(oId, uId int64) (int, error) {
	dao := order.Orders{}
	s := services.ServiceEntity{}
	auditDao := orderAudit.NewEntity()
	sarDao := serviceAuditRole.NewEntity()
	order, err := dao.FindOrderByOId(oId)
	if err != nil {
		return 0, global.DaoError(err)
	}
	err = optionServiceState(order.ServiceId, uId)
	if err != nil {
		return 0, err
	}
	service, err := s.GetServiceById(order.ServiceId)
	if err != nil {
		return 0, err
	}
	if service == nil {
		return 0, errors.New(global.MsgWarnNoHaveService)
	}
	// 订单处于拒绝或者通过
	if order.OrderResult != 0 {
		return 0, nil
	}
	// 该用户的层级审核
	ad, err := auditDao.UpdateAuditInfo(oId, uId, map[string]interface{}{
		"audit_result": 1,
		"update_time":  time.Now().Local(),
	})
	if err != nil {
		return 0, global.DaoError(err)
	}
	if ad == 1 {
		// 用户审核通过
		// 查看用户是不是超级审核员
		sar, err := sarDao.FindSARByUIdAndSId(uId, order.ServiceId)
		if err != nil {
			return 0, err
		}
		if sar.Aid == 4 {
			// 直接通过
			_, err3 := dao.UpdateOrdersInfo(oId, map[string]interface{}{
				"order_result": 1,
				"update_time":  time.Now().Local(),
			})
			if err3 != nil {
				return 0, global.DaoError(err3)
			}
			// 推送给管理后台，告诉钱包可以提币
			err := SendBillOutMsg(order.SerialNo)
			if err != nil {
				return 0, err
			}
		} else {
			// 订单信息
			count, err := auditDao.CountOrderAuditPassByOIdUId(oId)
			if service.AuditType <= int(count) {
				_, err = dao.UpdateOrdersInfo(oId, map[string]interface{}{
					"order_result": 1,
					"update_time":  time.Now().Local(),
				})
				if err != nil {
					return 0, err
				}
				// 更改账单状态
				// 推送至管理后台上链请求,转到链上地址-冻结
				err := SendBillOutMsg(order.SerialNo)
				if err != nil {
					return 0, err
				}
			}
		}
	}
	return ad, nil
}

func UpdateOrderAll(info *domain.SelectOrderInfo, id int64) (int, error) {
	dao := order.Orders{}
	list, err := dao.FindOrderListByState(info, id)
	if err != nil {
		return 0, global.DaoError(err)
	}
	// 遍历查出来的全部订单
	count := 0
	for i := 0; i < len(list); i++ {
		status, err := UpdatePassOrder(list[i].Id, id)
		if err != nil {
			return 0, global.DaoError(err)
		}
		count += status
	}
	return count, nil
}

// FindOrderList
// 查询订单列表
func FindOrderList(info *domain.SelectOrderInfo, id int64) ([]domain.SelectOrderList, int, error) {
	dao := order.Orders{}
	od := orderAudit.NewEntity()
	mr := merchant.NewEntity()
	var sl = []domain.SelectOrderList{}
	list, err := dao.FindOrderListByServices(info, id)
	if err != nil {
		return sl, 0, global.DaoError(err)
	}
	count, err := dao.CountOrderListByServices(info, id)
	if err != nil {
		return sl, 0, global.DaoError(err)
	}
	for i := 0; i < len(list); i++ {

		auditTypeName := ""
		auditResult := list[i].AuditResult
		if list[i].AuditType != 0 {
			index := list[i].AuditType - 1
			if index < 0 {
				return sl, 0, global.WarnMsgError("错误：-1")
			}
			auditTypeName = dict.AuditTypeList[index]
		}
		if list[i].OrderResult != 0 || auditResult == 2 {
			auditResult = list[i].OrderResult
		}
		// 统计审核人员
		asc, err := od.FindOrderAuditByOId(list[i].Id)
		if err != nil {
			return sl, 0, err
		}
		ms := ""
		mlist := []string{}
		for _, as := range asc {
			personal, err := mr.GetUserMerchantPersonal(as.UserId)
			if err != nil {
				return sl, 0, err
			}
			if personal == nil || personal.Id == 0 {
				continue
			}
			mlist = append(mlist, personal.Name)
		}
		at := ""
		ct := ""
		if !asc[0].UpdateTime.IsZero() {
			ct = list[i].CreateTime.Local().Format(global.YyyyMmDdHhMmSs)
		}
		if !asc[0].UpdateTime.IsZero() {
			at = list[i].UpdateTime.Local().Format(global.YyyyMmDdHhMmSs)
		}
		if len(mlist) != 0 {
			ms = strings.Join(mlist, ",")
			if !asc[0].UpdateTime.IsZero() {
				at = asc[0].UpdateTime.Local().Format(global.YyyyMmDdHhMmSs)
			}
		}
		statusName := dict.OrderResult[list[i].OrderResult]
		// 统计审核时间
		sl = append(sl, domain.SelectOrderList{
			SerialNo:      list[i].SerialNo,
			Id:            list[i].Id,
			MerchantId:    list[i].MerchantId,
			Phone:         list[i].Phone,
			CoinId:        list[i].CoinId,
			ChainId:       list[i].ChainId,
			ServiceId:     list[i].ServiceId,
			ServiceName:   list[i].ServiceName,
			CoinName:      list[i].CoinName,
			ChainName:     list[i].ChainName,
			Type:          list[i].Type,
			TypeName:      dict.OrderType[list[i].Type],
			AuditType:     list[i].AuditType,
			AuditTypeName: auditTypeName,
			Nums:          list[i].Nums,
			Fee:           list[i].Fee,
			UpChainFee:    list[i].UpChainFee,
			BurnFee:       list[i].BurnFee,
			DestroyFee:    list[i].DestroyFee,
			RealNums:      list[i].RealNums,
			ReceiveAddr:   list[i].ReceiveAddr,
			Memo:          list[i].Memo,
			ResultName:    statusName,
			OrderResult:   list[i].OrderResult,
			Reason:        list[i].Reason,
			ColorResult:   dict.OrderResultColor[statusName],
			CreateTime:    ct,
			AuditNames:    ms,
			AuditStatus:   auditResult,
			AuditTime:     at,
		})
	}
	return sl, count, nil
}

// FindOrderExport
// 查询订单列表
func FindOrderExport(info *domain.SelectOrderInfo, id int64) (bytes.Buffer, error) {
	dao := order.Orders{}
	info.Offset = 0
	info.Limit = 5000
	od := orderAudit.NewEntity()
	mr := merchant.NewEntity()
	xFile := xlsx.NewFile()
	sheet, err := xFile.AddSheet("Sheet1")
	if err != nil {
		return bytes.Buffer{}, err
	}
	page, err := dao.FindOrderListByServices(info, id)
	title := []string{"序号", "账单ID", "业务线ID", "业务线名称", "商户ID", "手机号", "主链币名", "代币名", "币种数量", "矿工费",
		"销毁数量", "实际到账数量", "订单时间", "审核人员", "审核状态", "审核时间", "备注"}
	r := sheet.AddRow()
	var ce *xlsx.Cell
	for _, v := range title {
		ce = r.AddCell()
		ce.Value = v
	}
	for i := 0; i < len(page); i++ {

		// 统计审核人员
		asc, err := od.FindOrderAuditByOId(page[i].Id)
		if err != nil {
			return bytes.Buffer{}, err
		}
		ms := ""
		mlist := []string{}
		for _, as := range asc {
			personal, err := mr.GetUserMerchantPersonal(as.UserId)
			if err != nil {
				return bytes.Buffer{}, err
			}
			if personal == nil || personal.Id == 0 {
				continue
			}
			mlist = append(mlist, personal.Name)
		}

		at := ""
		ct := ""
		if !asc[0].UpdateTime.IsZero() {
			ct = page[i].CreateTime.Local().Format(global.YyyyMmDdHhMmSs)
		}
		if !asc[0].UpdateTime.IsZero() {
			at = page[i].UpdateTime.Local().Format(global.YyyyMmDdHhMmSs)
		}

		if len(mlist) != 0 {
			ms = strings.Join(mlist, ",")
			if !asc[0].UpdateTime.IsZero() {
				at = asc[0].UpdateTime.Local().Format(global.YyyyMmDdHhMmSs)
			}
		}
		r = sheet.AddRow()
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(int64(i+1), 10) // 序号
		ce = r.AddCell()
		ce.Value = page[i].SerialNo // 账单ID
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(int64(page[i].ServiceId), 10) // 业务线ID
		ce = r.AddCell()
		ce.Value = page[i].ServiceName // 业务线名称
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(page[i].MerchantId, 10) // 商户ID
		ce = r.AddCell()
		ce.Value = page[i].Phone // 手机号
		ce = r.AddCell()
		ce.Value = page[i].ChainName // 主链币名
		ce = r.AddCell()
		ce.Value = page[i].CoinName // 代币名
		ce = r.AddCell()
		ce.Value = page[i].ServiceName // 业务线名称
		ce = r.AddCell()
		ce.Value = page[i].ChainName // 主链币
		ce = r.AddCell()
		ce.Value = page[i].CoinName // 代币
		ce = r.AddCell()
		ce.Value = page[i].Nums.String() // 币种数量
		ce = r.AddCell()
		ce.Value = page[i].UpChainFee.String() // 矿工费
		ce = r.AddCell()
		ce.Value = page[i].DestroyFee.String() // 销毁数量
		ce = r.AddCell()
		ce.Value = page[i].RealNums.String() // 实际到账数量
		ce = r.AddCell()
		ce.Value = ct // 订单时间
		ce = r.AddCell()
		ce.Value = ms // 审核人员
		ce = r.AddCell()
		ce.Value = dict.OrderResult[page[i].OrderResult] // 审核状态
		ce = r.AddCell()
		ce.Value = at // 审核时间
		ce = r.AddCell()
		ce.Value = page[i].Reason // 备注
	}
	//将数据存入buff中
	var buff bytes.Buffer
	if err = xFile.Write(&buff); err != nil {
		return bytes.Buffer{}, err
	}

	return buff, nil
}

func optionServiceState(sid int, uid int64) error {
	sconfig := services.ServiceAuditConfig{}
	sarDao := serviceAuditRole.NewEntity()
	user, err := sarDao.FindLevelUidAllService(sid, uid)
	if err != nil {
		return err
	}
	if user == nil {
		return global.WarnMsgError(global.OperationIsNotServiceAuthErr)
	}
	lid, err := sconfig.GetServiceConfigBySLid(sid, user.Aid)
	if err != nil {
		return err
	}
	if lid.State == 1 {
		return global.WarnMsgError(global.OperationIsServiceLevelFreeze)
	}
	return nil
}
