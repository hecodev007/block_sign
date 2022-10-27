package adminDeal

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	modelUser "custody-merchant-admin/model/adminPermission/user"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/order"
	"custody-merchant-admin/model/orderAudit"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/model/serviceAuditRole"
	"custody-merchant-admin/model/serviceChains"
	"custody-merchant-admin/model/services"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

func GetMainUserInfo(sel *domain.SelectUserInfo) (domain.SelectUserList, error) {
	var (
		dao      = merchant.NewEntity()
		mainUser = domain.SelectUserList{}
	)
	pInfo, err := dao.GetMerchantPersonal(sel.MerchantId, sel.Account)
	if err != nil {
		return mainUser, err
	}
	if pInfo == nil || pInfo.Id == 0 {
		return mainUser, errors.New(global.MsgWarnPidIsNil)
	}
	mainUser.Id = pInfo.Id
	mainUser.Name = pInfo.Name
	mainUser.Sex = pInfo.Sex
	mainUser.SexName = dict.SexText[pInfo.Sex]
	mainUser.Email = pInfo.Email
	mainUser.Phone = pInfo.Phone
	mainUser.PhoneCode = pInfo.PhoneCode
	mainUser.Role = pInfo.Role
	mainUser.RoleName = dict.SysRoleNameList[pInfo.Role-1]
	mainUser.Remark = pInfo.Remark
	mainUser.State = pInfo.State
	mainUser.StateName = dict.StateText[pInfo.State]
	mainUser.Passport = pInfo.Passport
	mainUser.IsTest = pInfo.IsTest
	mainUser.IsTestName = dict.IsTestNameList[pInfo.IsTest]
	mainUser.AccountErr = "否"
	if pInfo.PhoneCodeErr >= 5 || pInfo.PwdErr >= 5 || pInfo.EmailCodeErr >= 5 {
		mainUser.AccountErr = "是"
	}
	plist, err := dao.GetMerchantErr(sel.MerchantId)
	if err != nil {
		return mainUser, err
	}

	mainUser.PwdErr = 0
	mainUser.PhoneCodeErr = 0
	mainUser.EmailCodeErr = 0
	for _, entity := range plist {
		mainUser.PwdErr += entity.PwdErr
		mainUser.PhoneCodeErr += entity.PhoneCodeErr
		mainUser.EmailCodeErr += entity.EmailCodeErr
	}
	return mainUser, nil
}

func GetUserInfoList(userSelect *domain.SelectUserInfo) ([]domain.SelectUserList, int64, error) {

	var (
		mDao          = merchant.NewEntity()
		sar           = serviceAuditRole.NewEntity()
		sdao          = services.ServiceEntity{}
		selectList    []domain.SelectUserList
		sline         []int
		serviceLevels []domain.ServiceAuditRole
	)

	list, count, err := mDao.FindSubUserInfoList(userSelect.MerchantId, userSelect)
	if err != nil {
		return selectList, count, err
	}
	// 遍历用户信息
	for i, _ := range list {
		serviceName := []string{}
		serviceLevelName := []string{}
		// 查询用户的业务线
		uids, err := sar.GetUserServiceByUid(list[i].Id)
		if err != nil {
			return selectList, count, err
		}
		// 查询业务线的信息并且拼接
		for i2, _ := range uids {
			service, err := sdao.GetServiceById(uids[i2].Sid)
			if err != nil {
				return selectList, count, err
			}
			if service != nil && service.Id > 0 {
				serviceName = append(serviceName, service.Name)
				sline = append(sline, service.Id)
			}
		}
		// 查询用户所拥有的审核角色
		sRole, err := sar.FindUserAuditRoleName(list[i].Id)
		if err != nil {
			return selectList, count, err
		}
		roleAndService := []string{}
		roleAndAudit := []string{}
		for k, _ := range sRole {
			sn := sRole[k].ServiceName
			roleAudit := dict.SysRoleNameList[list[i].Role-1]
			if sRole[k].AuditName != "" {
				sn += "-" + sRole[k].AuditName
				roleAudit += "-" + sRole[k].AuditName + "(" + sRole[k].ServiceName + ")"
			}
			roleAndAudit = append(roleAndAudit, roleAudit)
			serviceLevelName = append(serviceLevelName, sn)
			roleAndService = append(roleAndService, dict.SysRoleNameList[list[i].Role-1]+"-"+sRole[k].ServiceName)
			serviceLevels = append(serviceLevels, domain.ServiceAuditRole{
				ServiceId:  sRole[k].Sid,
				AuditLevel: sRole[k].Aid,
			})
		}
		isErr := 0
		if list[i].EmailCodeErr >= 5 || list[i].PhoneCodeErr >= 5 || list[i].PwdErr >= 5 {
			isErr = 1
		}
		ct := ""
		if list[i].CreateTime != nil {
			ct = list[i].CreateTime.Format(global.YyyyMmDdHhMmSs)
		}
		lt := ""
		if list[i].LoginTime != nil {
			lt = list[i].LoginTime.Format(global.YyyyMmDdHhMmSs)
		}
		selectList = append(selectList, domain.SelectUserList{
			Serial:             i + 1,
			Id:                 list[i].Id,
			Pid:                userSelect.MerchantId,
			Name:               list[i].Name,
			Sex:                list[i].Sex,
			Email:              list[i].Email,
			Phone:              list[i].Phone,
			PhoneCode:          list[i].PhoneCode,
			RoleAndAudit:       strings.Join(roleAndAudit, ";"),
			RoleAndService:     strings.Join(roleAndService, ";"),
			ServiceLevelName:   strings.Join(serviceLevelName, ";"),
			ServiceAuditLevels: serviceLevels,
			Services:           sline,
			ServiceName:        strings.Join(serviceName, ";"),
			Role:               list[i].Role,
			IsTest:             list[i].IsTest,
			IsTestName:         dict.IsTestNameList[list[i].IsTest],
			RoleName:           dict.SysRoleNameList[list[i].Role-1],
			Remark:             list[i].Remark,
			Reason:             list[i].Reason,
			State:              list[i].State,
			Show:               xkutils.ThreeDo(list[i].Id == userSelect.MerchantId, 0, 1).(int),
			Passport:           list[i].Passport,
			Identity:           list[i].Identity,
			SexName:            dict.SexText[list[i].Sex],
			StateName:          dict.StateText[list[i].State],
			LoginTime:          lt,
			CreateTime:         ct,
			IsErr:              dict.BaseText[isErr],
		})
	}
	return selectList, count, nil
}

func GetUserInfoById(uid int64) (domain.SelectUserList, error) {
	var (
		dao           = merchant.NewEntity()
		sar           = serviceAuditRole.NewEntity()
		sdao          = services.ServiceEntity{}
		selectList    domain.SelectUserList
		sline         []int
		serviceLevels []domain.ServiceAuditRole
	)
	info, err := dao.GetUserMerchantPersonal(uid)
	if err != nil {
		return selectList, err
	}
	serviceName := []string{}
	serviceLevelName := []string{}
	sRole, err := sar.FindUserAuditRoleName(info.Id)
	if err != nil {
		return selectList, err
	}
	for k, _ := range sRole {
		sn := sRole[k].ServiceName
		if sRole[k].AuditName != "" {
			sn += "-" + sRole[k].AuditName
		}
		serviceLevelName = append(serviceLevelName, sn)
		serviceLevels = append(serviceLevels, domain.ServiceAuditRole{
			ServiceId:  sRole[k].Sid,
			AuditLevel: sRole[k].Aid,
		})
	}
	uids, err := sar.GetUserServiceByUid(info.Id)
	if err != nil {
		return selectList, err
	}
	for i2, _ := range uids {
		service, err := sdao.GetServiceById(uids[i2].Sid)
		if err != nil {
			return selectList, err
		}
		if service != nil && service.Id > 0 {
			serviceName = append(serviceName, service.Name)
			sline = append(sline, service.Id)
		}
	}

	if dict.SysMerchantRoleTagList[info.Role-1] != "audit" {
		serviceLevels = []domain.ServiceAuditRole{}
		for i, _ := range sline {
			sa := domain.ServiceAuditRole{}
			sa.ServiceId = sline[i]
			serviceLevels = append(serviceLevels, sa)
		}
	}

	lt := ""
	ct := ""
	if info.LoginTime != nil {
		lt = info.LoginTime.Format(global.YyyyMmDdHhMmSs)
	}
	if info.CreateTime != nil {
		ct = info.CreateTime.Format(global.YyyyMmDdHhMmSs)
	}

	selectList = domain.SelectUserList{
		Id:                 info.Id,
		Pid:                info.Pid,
		Name:               info.Name,
		Sex:                info.Sex,
		Email:              info.Email,
		Phone:              info.Phone,
		PhoneCode:          info.PhoneCode,
		ServiceLevelName:   strings.Join(serviceLevelName, ";"),
		ServiceName:        strings.Join(serviceName, ";"),
		ServiceAuditLevels: serviceLevels,
		Services:           sline,
		Role:               info.Role,
		RoleName:           dict.SysRoleNameList[info.Role-1],
		Remark:             info.Remark,
		Reason:             info.Reason,
		State:              info.State,
		SexName:            dict.SexText[info.Sex],
		StateName:          dict.StateText[info.State],
		Passport:           info.Passport,
		Identity:           info.Identity,
		IsTest:             info.IsTest,
		PwdErr:             info.PwdErr,
		PhoneCodeErr:       info.PhoneCodeErr,
		EmailCodeErr:       info.EmailCodeErr,
		LoginTime:          lt,
		CreateTime:         ct,
	}
	return selectList, nil
}

func HaveUserByPIdAndUId(id, pid int64) error {
	dao := merchant.NewEntity()
	user, err := dao.HaveUserByPIdAndUId(id, pid)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(global.MsgWarnAccountIsSubErr)
	}
	return nil
}

func UpdateMerchantSubUser(user *domain.SaveUserInfo) error {
	dao := merchant.NewEntity()
	orderAuditDao := orderAudit.NewEntity()
	sarDao := serviceAuditRole.NewEntity()
	idu, err := dao.GetUserMerchantPersonal(user.Id)
	if err != nil {
		return err
	}
	if idu == nil {
		return errors.New(global.MsgWarnNotUser)
	}
	thisMerchant, err := dao.GetUserMerchantPersonal(idu.Pid)
	if err != nil {
		return err
	}
	if idu.Pid != 0 && thisMerchant == nil {
		return errors.New("无法查询主账户")
	}
	err = HavePhoneEmail(user.Phone, user.Email, idu)
	if err != nil {
		return err
	}

	u := map[string]interface{}{
		"pid":         user.Pid,
		"name":        user.Name,
		"sex":         user.Sex,
		"email":       user.Email,
		"phone":       user.Phone,
		"phone_code":  user.PhoneCode,
		"passport":    user.Passport,
		"identity":    user.Identity,
		"roles":       user.Role,
		"is_test":     thisMerchant.IsTest,
		"remark":      user.Remark,
		"state":       0,
		"update_time": time.Now().Local(),
	}
	if thisMerchant.TestTime == nil || !thisMerchant.TestTime.IsZero() {
		u["test_time"] = thisMerchant.TestTime
	}
	// 先更新用户信息
	err = dao.UpdatePersonalUser(user.Id, u)
	if err != nil {
		return err
	}
	// 更新业务线
	// 判断是不是审核员
	// 原来是审核员，现在不是
	if dict.SysMerchantRoleTagList[idu.Role-1] == "audit" && idu.Role != user.Role {
		// audit = true
		err = orderAuditDao.DelOrderAudit(user.Id, 0)
		if err != nil {
			return err
		}
		err = sarDao.DelUserAuditRole(user.Id)
		if err != nil {
			return err
		}
	}

	// 遍历更新 用户-业务线-审核角色
	err = saveUserServiceAuditRole(user)
	if err != nil {
		return err
	}
	return nil
}

func saveUserServiceAuditRole(u *domain.SaveUserInfo) error {
	var (
		sarDao   = serviceAuditRole.NewEntity()
		oaDao    = orderAudit.NewEntity()
		orderDao = new(order.Orders)
		state    = 0
	)

	// 判断是不是审核员
	if dict.SysMerchantRoleTagList[u.Role-1] != "audit" && dict.SysMerchantRoleTagList[u.Role-1] != "admin" {
		var sclst []int
		for i, _ := range u.ServiceAuditLevels {
			sclst = append(sclst, u.ServiceAuditLevels[i].ServiceId)
		}
		err := sarDao.UpdateUserServices(u.Id, sclst)
		if err != nil {
			return err
		}
		return nil
	}

	if dict.SysMerchantRoleTagList[u.Role-1] == "admin" {
		state = 1
	}

	user := u.ServiceAuditLevels
	if len(user) != 0 {
		err := sarDao.DelUserAuditRole(u.Id)
		if err != nil {
			return err
		}
		err = oaDao.DelOrderAudit(u.Id, 0)
		if err != nil {
			return err
		}
	}
	for i, _ := range user {
		sar := &serviceAuditRole.Entity{
			Uid:   u.Id,
			Sid:   user[i].ServiceId,
			Aid:   user[i].AuditLevel,
			State: state,
		}
		err := sarDao.SaveSARInfo(sar)
		if err != nil {
			return err
		}

		// 更新业务线人员配置
		err = updateAddServiceConfigLevel(user[i].ServiceId, user[i].AuditLevel)
		if err != nil {
			return err
		}
		// 新增审核进度
		auditList, err := orderDao.FindNoResult(user[i].ServiceId)
		if err != nil {
			return err
		}
		// 等级是否存在审核进度
		if len(auditList) > 0 {
			// 要新增的审核进度的订单
			var orders = []orderAudit.Entity{}
			for x := 0; x < len(auditList); x++ {
				level, err := oaDao.FindAuditInfoByLevel(auditList[x].Id, sar.Uid)
				if err != nil {
					return err
				}
				if level.Id > 0 {
					continue
				}
				orders = append(orders, orderAudit.Entity{
					AuditLevel:  sar.Aid,
					State:       0,
					AuditResult: 0,
					OrderId:     auditList[x].Id,
					UserId:      sar.Uid,
					CreateTime:  time.Now().Local(),
				})
			}
			if len(orders) > 0 {
				// 批量新增
				err = oaDao.BatchCreateAuditInfo(orders)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func SaveSuperAudit(s *domain.MerchantService) error {
	// admin_user 修改用户
	mu := modelUser.NewEntity()
	oDao := orderAudit.NewEntity()
	orderDao := new(order.Orders)
	sar := serviceAuditRole.NewEntity()
	var (
		delList []int
		uid     int64
	)
	personal, err := mu.GetAdminPersonal(s.Id)
	if err != nil {
		return err
	}
	if personal == nil {
		return errors.New(global.MsgWarnAccountErr)
	}
	// user_info 用户新增
	if personal.Uid == 0 {
		db := orm.Cache(model.DB().Begin())
		user := merchant.Entity{
			Db:     db,
			Name:   "平台" + personal.Name,
			Phone:  personal.Phone,
			Email:  personal.Email,
			IsTest: 0, // 测试账户
		}
		err = user.InsertNewMerchant()
		if err != nil {
			return err
		}
		uid = user.Id
		_, err = personal.UpdateUserById(s.Id, map[string]interface{}{"uid": uid})
		if err != nil {
			return err
		}
	} else {
		// 已经有
		uid = personal.Uid
	}
	// 现在拥有的
	list, err := sar.GetUserServiceByUid(uid)
	if err != nil {
		return err
	}
	// 删除的
	for i, _ := range list {
		pass := true
		for j, _ := range s.HaveService {
			if s.HaveService[j] == list[i].Sid {
				pass = false
				break
			}
		}
		if pass {
			pass = false
			delList = append(delList, list[i].Sid)
		}
	}
	err = sar.DelUserHaveService(uid, delList)
	if err != nil {
		return err
	}
	for _, del := range delList {
		err = oDao.DelOrderAuditByUSId(uid, del)
		if err != nil {
			return err
		}
	}
	for _, sid := range s.AddService {
		err = sar.SaveSARInfo(&serviceAuditRole.Entity{
			Uid:   uid,
			Sid:   sid,
			Aid:   4,
			State: 1,
		})
		if err != nil {
			return err
		}
		// 新增审核进度
		auditList, err := orderDao.FindNoResult(sid)
		if err != nil {
			return err
		}
		// 等级是否存在审核进度
		if len(auditList) > 0 {
			// 要新增的审核进度的订单
			var orders = []orderAudit.Entity{}
			for x := 0; x < len(auditList); x++ {
				level, err := oDao.FindAuditInfoByLevel(auditList[x].Id, sar.Uid)
				if err != nil {
					return err
				}
				if level.Id > 0 {
					continue
				}
				orders = append(orders, orderAudit.Entity{
					AuditLevel:  sar.Aid,
					State:       0,
					AuditResult: 0,
					OrderId:     auditList[x].Id,
					UserId:      sar.Uid,
					CreateTime:  time.Now().Local(),
				})
			}
			if len(orders) > 0 {
				// 批量新增
				err = oDao.BatchCreateAuditInfo(orders)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func HavePhoneEmail(phone, email string, user *merchant.Entity) error {
	dao := merchant.NewEntity()
	if phone == "" && email == "" {
		return global.WarnMsgError(global.MsgWarnNoEmailPhone)
	} else {
		if email != "" && user.Email != email {
			userEmail, err := dao.GetSubUserByEmail(email)
			if err != nil {
				return err
			}
			if userEmail != nil && userEmail.Id > 0 {
				return global.WarnMsgError(global.MsgWarnHaveEmail)
			}
		}
		if phone != "" && user.Phone != phone {
			userPhone, err := dao.GetSubUserByPhone(phone)
			if err != nil {
				return err
			}
			if userPhone != nil && userPhone.Id > 0 {
				return global.WarnMsgError(global.MsgWarnHavePhone)
			}
		}
	}
	return nil
}

func FindServiceAudit(id int64) (domain.MerchantServiceList, error) {

	mu := modelUser.NewEntity()
	m := merchant.NewEntity()
	sc := serviceChains.NewEntity()
	ms := domain.MerchantServiceList{}
	ms.ServiceList = []domain.UServiceList{}
	ms.HaveService = []domain.UServiceList{}
	sids := []int{}
	personal, err := mu.GetAdminPersonal(id)
	if err != nil {
		return ms, err
	}
	// 查询 service_audit_role---service_chains---service 并且等于personal.Uid
	HaveList, err := sc.GetHaveServiceChainsList(personal.Uid)
	if err != nil {
		return ms, err
	}
	for _, info := range HaveList {
		merchantPersonal, err := m.GetUserMerchantPersonal(info.MerchantId)
		if err != nil {
			return ms, err
		}
		if merchantPersonal != nil {
			sids = append(sids, info.ServiceId)
			ms.HaveService = append(ms.HaveService, domain.UServiceList{
				ServiceId:       info.ServiceId,
				ServiceName:     info.Name,
				UserId:          personal.Uid,
				UserName:        personal.Name,
				MerchantId:      merchantPersonal.Id,
				MerchantName:    merchantPersonal.Name,
				ServiceMerchant: fmt.Sprintf("%s(%s-%d)", info.Name, merchantPersonal.Name, personal.Uid),
				ServiceRole:     fmt.Sprintf("%s-%s", info.Name, dict.AuditLevelName[3]),
			})
		}
	}
	// 查询  service_audit_role---service_chains---service 并且不等于personal.Uid
	chainsList, err := sc.GetNoServiceChainsList(sids)
	if err != nil {
		return ms, err
	}
	for _, info := range chainsList {
		merchantPersonal, err := m.GetUserMerchantPersonal(info.MerchantId)
		if err != nil {
			return ms, err
		}
		if merchantPersonal == nil {
			continue
		}
		ms.ServiceList = append(ms.ServiceList, domain.UServiceList{
			ServiceId:       info.ServiceId,
			ServiceName:     info.Name,
			UserId:          personal.Uid,
			UserName:        personal.Name,
			MerchantName:    merchantPersonal.Name,
			ServiceMerchant: fmt.Sprintf("%s(%s-%d)", info.Name, merchantPersonal.Name, personal.Uid),
		})
	}
	return ms, nil
}

// updateAddServiceConfigLevel
// 更新业务线人员配置
func updateAddServiceConfigLevel(sid, level int) error {
	s := new(services.ServiceAuditConfig)
	sarDao := serviceAuditRole.NewEntity()
	var uids = ""
	li, err := sarDao.FindSARBySIdAndAId(sid, level)
	if err != nil {
		return err
	}
	for i2, _ := range li {
		if i2 != 0 {
			uids += ","
		}
		uids += fmt.Sprintf("%d", li[i2].Uid)
	}
	if len(li) > 0 {
		aInfo, err := s.GetServiceConfigBySLid(sid, level)
		if err != nil {
			return err
		}
		if aInfo != nil {
			err = s.UpdateServiceConfigLevel(aInfo.Id, map[string]interface{}{"users": uids})
			if err != nil {
				return err
			}
		} else {
			aInfo = &services.ServiceAuditConfig{
				ServiceId:  sid,
				AuditLevel: level,
				AuditType:  0,
				Users:      uids,
				LimitUse:   0,
				NumEach:    decimal.Decimal{},
				NumDay:     decimal.Decimal{},
				NumWeek:    decimal.Decimal{},
				NumMonth:   decimal.Decimal{},
				State:      0,
			}
			err = s.CreateServiceConfigLevel(aInfo)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func UpdateSubUserInfoById(uInfo *domain.SelectUserList) error {
	var (
		dao      = merchant.NewEntity()
		upMap    map[string]interface{}
		sarDao   = serviceAuditRole.NewEntity()
		oaDao    = orderAudit.NewEntity()
		orderDao = new(order.Orders)
		sDao     = new(services.ServiceEntity)
	)
	if uInfo.State == 0 {
		upMap = map[string]interface{}{
			"state":  uInfo.State,
			"reason": "",
		}
	}
	if uInfo.State == 1 || uInfo.State == 2 {
		upMap = map[string]interface{}{
			"state":  uInfo.State,
			"reason": uInfo.Reason,
		}
	}
	info, err := dao.GetUserMerchantPersonal(uInfo.Id)
	if err != nil {
		return err
	}
	if info == nil {
		return global.NewError(global.MsgWarnAccountErr)
	}
	// 角色是审核员
	// 删除或者冻结
	if dict.SysMerchantRoleTagList[info.Role-1] == "audit" && uInfo.State != 0 {
		uh, err := sarDao.GetUserServiceByUid(uInfo.Id)
		if err != nil {
			return err
		}
		for i := 0; i < len(uh); i++ {
			if v, err := sDao.GetServiceById(uh[i].Sid); err == nil {
				sid, err := sarDao.GetUserServiceBySid(uh[i].Sid)
				if err != nil {
					return err
				}
				if v.AuditType > len(sid)-1 {
					return errors.New(global.OperationDelUserHaveServiceErr)
				}
			} else {
				return err
			}
		}
		// 更新审核业务
		err = sarDao.UpdateUserAuditRole(uInfo.Id, map[string]interface{}{"state": 2})
		if err != nil {
			return err
		}
		// 删除订单进度
		err = oaDao.DelOrderAudit(uInfo.Id, 0)
		if err != nil {
			return err
		}
		for i, _ := range uh {
			err = updateAddServiceConfig(uh[i].Sid)
			if err != nil {
				return err
			}
		}
	}

	// 更新其他信息或者解冻
	if uInfo.State == 0 {
		// 更新审核业务
		err = sarDao.UpdateUserAuditRole(uInfo.Id, map[string]interface{}{"state": 0})
		if err != nil {
			return err
		}
		ors, err := orderDao.FindOrderByUId(uInfo.Id)
		if err != nil {
			return err
		}
		for i, _ := range ors {

			uLst, err := sarDao.FindLevelUid(ors[i].ServiceId, uInfo.Id)
			if err != nil {
				return err
			}
			// 判断是否有审核等级
			if uLst == nil {
				continue
			}

			// 新增订单进度
			err = oaDao.CreateAuditInfo(&orderAudit.Entity{
				AuditLevel:  uLst.Aid,
				State:       0,
				AuditResult: 0,
				OrderId:     ors[i].Id,
				UserId:      uInfo.Id,
				CreateTime:  time.Now().Local(),
			})
		}
	}
	_, err = dao.UpdateSubUserById(uInfo.Id, upMap)
	if err != nil {
		return err
	}
	return nil
}

func updateAddServiceConfig(sid int) error {
	sarDao := serviceAuditRole.NewEntity()
	cDao := new(services.ServiceAuditConfig)
	for i := 0; i < 4; i++ {
		var (
			uids = ""
		)
		li, err := sarDao.FindSARBySIdAndAId(sid, i+1)
		if err != nil {
			return err
		}
		for i2, _ := range li {
			if i2 != 0 {
				uids += ","
			}
			uids += fmt.Sprintf("%d", li[i2].Uid)
		}
		if len(li) > 0 {
			aInfo, err := cDao.GetServiceConfigBySLid(sid, i+1)
			if err != nil {
				return err
			}
			if aInfo != nil {
				err = cDao.UpdateServiceConfigLevel(aInfo.Id, map[string]interface{}{"users": uids})
				if err != nil {
					return err
				}
			} else {
				aInfo = &services.ServiceAuditConfig{
					ServiceId:  sid,
					AuditLevel: i + 1,
					AuditType:  0,
					Users:      uids,
					LimitUse:   0,
					NumEach:    decimal.Decimal{},
					NumDay:     decimal.Decimal{},
					NumWeek:    decimal.Decimal{},
					NumMonth:   decimal.Decimal{},
					State:      0,
				}
				err = cDao.CreateServiceConfigLevel(aInfo)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func GetUserServiceAudit(id int64) (*domain.UserHaveServiceAuditLevel, error) {
	var (
		usal = new(domain.UserHaveServiceAuditLevel)
		dao  = serviceAuditRole.NewEntity()
		m    = merchant.NewEntity()
	)
	usal.UServices = []domain.UService{}
	usal.ServiceAndLevels = []domain.ServiceAndLevel{}
	info, err := dao.FindUserServiceByUid(id)
	if err != nil {
		return nil, err
	}
	personal, err := m.GetUserMerchantPersonal(id)
	if err != nil {
		return nil, err
	}
	if personal == nil || personal.Id <= 0 {
		return nil, errors.New(global.MsgWarnAccountIsNil)
	}
	usal.RoleId = personal.Role
	usal.RoleName = dict.SysRoleNameList[personal.Role-1]
	if err != nil {
		return usal, err
	}
	for i, _ := range info {
		usal.UServices = append(usal.UServices, domain.UService{
			ServiceId:   info[i].Id,
			ServiceName: info[i].Name,
		})
		findSAR, err := dao.FindSARByUIdAndSId(id, info[i].Id)
		if err != nil {
			return usal, err
		}
		if findSAR == nil {
			continue
		}
		// 放入列表
		usal.ServiceAndLevels = append(usal.ServiceAndLevels, domain.ServiceAndLevel{
			Level:     findSAR.Aid,
			LevelName: dict.AuditLevelName[findSAR.Aid-1],
			UService: domain.UService{
				ServiceId:   info[i].Id,
				ServiceName: info[i].Name,
			},
		})
	}
	return usal, nil
}

func GetAllMerchantService(id int64) ([]domain.UService, error) {
	var (
		usal = []domain.UService{}
		dao  = serviceAuditRole.NewEntity()
	)
	info, err := dao.FindUserServiceByUid(id)
	if err != nil {
		return usal, err
	}
	for i, _ := range info {
		usal = append(usal, domain.UService{
			ServiceId:   info[i].Id,
			ServiceName: info[i].Name,
		})
	}
	return usal, nil
}

func ClearSubInfoByPId(pid int64, mp map[string]interface{}) error {
	m := merchant.NewEntity()
	err := m.ClearSubUserByPId(pid, mp)
	if err != nil {
		return err
	}
	return nil
}

func ClearSubInfoById(id int64, mp map[string]interface{}) error {

	m := merchant.NewEntity()
	err := m.ClearSubUserByPId(id, mp)
	if err != nil {
		return err
	}
	return nil
}
