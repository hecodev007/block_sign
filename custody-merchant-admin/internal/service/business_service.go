package service

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/apply"
	"custody-merchant-admin/model/assets"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/business"
	"custody-merchant-admin/model/businessOrder"
	"custody-merchant-admin/model/businessPackage"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/orm"
	_package "custody-merchant-admin/model/package"
	"custody-merchant-admin/model/record"
	"custody-merchant-admin/model/serviceAuditConfig"
	"custody-merchant-admin/model/serviceAuditRole"
	"custody-merchant-admin/model/serviceChains"
	"custody-merchant-admin/model/serviceSecurity"
	"custody-merchant-admin/model/white"
	"custody-merchant-admin/module/blockChainsApi"
	"custody-merchant-admin/module/log"
	validator "github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"reflect"
	"strings"

	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// CreateBusinessItem 创建业务线
func CreateBusinessItem(user *domain.JwtCustomClaims, reqS *domain.CreateBusinessInfo, reqM *map[string]interface{}) (err error) {

	isValid, mInfo, bpInfo, err := determainBusinessParametersIsValid(reqS, reqM)
	if !isValid {
		return err
	}
	//
	b, errB := json.Marshal(reqS)
	if errB != nil {
		return global.OperationErrorText(errB.Error())
	}

	//钱包创建用户,每个业务线单独一个client_id,secret
	newMch := domain.BCMchReq{
		Name:  reqS.BusinessName,
		Phone: reqS.Phone,
		Email: reqS.Email,
		//CompanyImg: item.BusinessPicture,
	}
	var mchInfo domain.BCMchInfo
	mchInfo, err = blockChainsApi.BlockChainCreateClientIdSecret(newMch, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
	if err != nil {
		log.Errorf("钱包创建用户err:%v", err.Error())
		if strings.Contains(err.Error(), "注册商户数据重复") {
			return fmt.Errorf("商户数据重复")
		} else {
			return fmt.Errorf("注册钱包数据错误")
		}
	}

	//根据name获取id获取
	if reqS.DeductCoinName == "" && reqS.DeductCoinId != "" {
		reqS.DeductCoinName = GetDeductCoinName(reqS.DeductCoinId)
	} else if reqS.DeductCoinName != "" && reqS.DeductCoinId == "" {
		reqS.DeductCoinId = GetDeductCoinId(reqS.DeductCoinName)
	}
	//钱包绑定回调地址
	coinArr := strings.Split(reqS.DeductCoinName, ",")
	chainArr := make([]string, 0)
	for _, item := range coinArr {
		c := FindChainName(item)
		c = strings.ToLower(c)
		chainArr = append(chainArr, c)
	}
	err = blockChainsApi.BlockChainBindAddress(chainArr, mchInfo.ClientId, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
	if err != nil {
		log.Errorf("钱包创建用户,绑定回调地址err:%v", err.Error())
		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	}
	t := time.Now()
	db := orm.Cache(model.DB().Begin())
	//创建业务线
	bInfo := business.NewEntity()
	bInfo.Db = db
	bInfo.Name = reqS.BusinessName
	bInfo.Phone = reqS.Phone
	bInfo.Email = reqS.Email
	bInfo.AccountId = reqS.AccountId
	bInfo.AccountStatus = mInfo.IsTest
	bInfo.Coin = strings.ToUpper(reqS.Coin)
	bInfo.SubCoin = strings.ToUpper(reqS.SubCoin)
	bInfo.LimitTransfer = 0 // 是否限制转账：1是，0否
	bInfo.FounderId = user.Id
	bInfo.WithdrawalStatus = reqS.IsWithdrawal
	bInfo.LimitSameWithdrawal = 0
	bInfo.AuditType = 1
	bInfo.State = 2
	bInfo.CreateTime = &t
	bInfo.UpdateTime = &t
	err = bInfo.InsertNewBusiness()
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	}

	//获取新建业务线id
	newInfo, _ := bInfo.FindBusinessItemByAccountId(reqS.AccountId)
	newInfo.Db = db
	newId := newInfo.Id
	if newId == 0 {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnCreateComboErr)
	}

	bpInfo.Id = 0
	bpInfo.PackageId = int64(reqS.PackageId)
	bpInfo.DeductCoin = reqS.DeductCoinName
	bpInfo.AccountId = reqS.AccountId
	bpInfo.BusinessId = newId
	bpInfo.Db = db
	err = bpInfo.InsertNewBP()
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	}

	//业务线安全信息表
	ssInfo := serviceSecurity.NewEntity()
	err = json.Unmarshal(b, &ssInfo)
	ssInfo.Id = 0
	ssInfo.BusinessId = newId
	ssInfo.ClientId = mchInfo.ClientId
	ssInfo.IsAccountCheck = reqS.IsAccountCheck
	ssInfo.IsPlatformCheck = reqS.IsPlatformCheck
	ssInfo.IsWithdrawal = reqS.IsWithdrawal
	ssInfo.IsWhitelist = reqS.IsWhitelist
	ssInfo.ClientId = mchInfo.ClientId
	ssInfo.Secret = mchInfo.Secret
	ssInfo.Db = db
	err = ssInfo.InsertNewItem()
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	}
	// 关闭白名单
	white.CloseWhiteUseById(newId, map[string]interface{}{"use": 1})
	//serviceauditrole
	sar := serviceAuditRole.NewEntity()
	sar.Db = db
	sar.Uid = reqS.AccountId
	sar.Sid = int(newId)
	sar.Aid = 4
	sar.State = 1
	err = sar.InsertNewPackage()
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	}
	sac := serviceAuditConfig.NewEntity()
	sac.Db = db
	sac.ServiceId = int(newId)
	sac.Users = fmt.Sprintf("%d", reqS.AccountId)
	sac.NumWeek = decimal.Zero
	sac.NumEach = decimal.Zero
	sac.NumDay = decimal.Zero
	sac.NumMonth = decimal.Zero
	sac.LimitUse = 1 // 设置审核额度 limit_use 0是打开，1是关闭
	sac.Reason = ""
	sac.State = 0
	sac.AuditLevel = 4
	sac.AuditType = 1
	err = sac.CreateServiceConfigLevel()
	if err != nil {
		db.Rollback()
		return err
	}
	// 将业务线添加入超级管理员下
	//mtc := merchant.NewEntity()
	//err = mtc.FindMerchantItemByRole(1)
	//if err != nil {
	//	return err
	//}
	//if mtc != nil && mtc.Id > 0 {
	//	sr := serviceAuditRole.NewEntity()
	//	sr.Db = db
	//	sr.Uid = mtc.Id
	//	sr.Sid = int(newId)
	//	sr.Aid = 4
	//	sr.State = 1
	//	err = sr.InsertNewPackage()
	//	if err != nil {
	//		db.Rollback()
	//		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	//	}
	//}
	//创建业务线订单
	err = NewBusinessCreateNewOrder(user, &newInfo, bpInfo)
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	}
	//扣费多个币
	//创建新的业务线下币种地址
	log.Errorf("创建业务线 CreateNewChainAddress")
	CreateNewChainAddress(user.Id, mchInfo.ClientId, coinArr)
	log.Errorf("创建业务线 delayCreateAddress")
	go delayCreateAddress(newId, reqS.AccountId, mchInfo, coinArr)
	//保存操作日志
	log.Errorf("创建业务线 SaveRecord")
	err = SaveRecord(db, user, newId, record.BusinessRecord, "create")
	if err != nil {
		db.Rollback()
		return
	}
	err = db.Commit().Error
	log.Errorf("创建业务线 err = %v", err)
	return
}

//延迟获取地址
func delayCreateAddress(newId, accountId int64, mchInfo domain.BCMchInfo, coinArr []string) {
	var err error
	time.Sleep(5 * time.Second)
	log.Infof("delayCreateAddress Sleep over\n")
	log.Infof("delayCreateAddress coinArr = %v\n", coinArr)
	//扣费多个币
	//创建新的业务线下币种地址
	var address map[string]string
	address, err = CreateNewChainAddress(newId, mchInfo.ClientId, coinArr)
	if err != nil {
		log.Errorf("delayCreateAddress CreateNewChainAddress err:%v\n", err)
		return
	}
	fmt.Println(address)

	cArr := make([]string, 0)
	for _, item := range coinArr {
		chainName := FindChainName(item)
		upItem := strings.ToUpper(item)
		if upItem != chainName {
			cArr = append(cArr, chainName)
		}
		cArr = append(cArr, item)
	}
	log.Infof("delayCreateAddress cArr = %v\n", cArr)

	db := orm.Cache(model.DB().Begin())
	cInfoArr, _ := base.FindCoinsInName(cArr)
	log.Infof("delayCreateAddress cInfoArr = %+v\n", cInfoArr)

	for _, item := range cInfoArr {
		cName := FindChainName(item.Name)
		var cAddress string
		for k, v := range address {
			kname := strings.ToUpper(k)
			if cName == kname {
				cAddress = v
			}
		}

		//insert assets表
		scInfo := serviceChains.NewEntity()
		scInfo.Db = db
		scInfo.MerchantId = accountId
		scInfo.ServiceId = int(newId)
		scInfo.CoinName = strings.ToUpper(item.Name)
		scInfo.CoinId = int(item.Id)
		scInfo.ChainAddr = cAddress
		err = scInfo.InsertNewItem()
		if err != nil {
			db.Rollback()
			log.Errorf("delayCreateAddress InsertNewItem err:%v\n", err)
			return
		}

		aInfo := assets.NewEntity()
		aInfo.Db = db
		//aInfo.CoinId = reqS.DeductCoinId
		aInfo.ServiceId = int(newId)
		aInfo.CoinId = int(item.Id)
		aInfo.CoinName = strings.ToUpper(item.Name)
		//aInfo.ChainAddress = v
		aInfo.FinanceFreeze = decimal.Zero
		err = aInfo.InsertNewAssets()
		if err != nil {
			db.Rollback()
			log.Errorf("delayCreateAddress InsertNewAssets err:%v\n", err)
			return
		}
	}
	db.Commit()

}

// DeleteBusinessItem 删除业务线
func DeleteBusinessItem(user *domain.JwtCustomClaims, packageId int64) (err error) {
	//查询id是否存在
	pInfo := business.NewEntity()
	err = pInfo.FindBusinessItemById(packageId)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}
	db := orm.Cache(model.DB().Begin())
	pInfo.Db = db
	err = pInfo.DeleteBusinessItem(packageId)
	if err != nil {
		db.Rollback()
		return global.DaoError(err)
	}

	//保存操作日志
	err = SaveRecord(db, user, packageId, record.BusinessRecord, "delete")
	if err != nil {
		db.Rollback()
		return
	}
	err = db.Commit().Error
	return
}

// UpdateBusinessItem 更新业务线
func UpdateBusinessItem(user *domain.JwtCustomClaims, req map[string]interface{}) (err error) {

	id := GetIntFromInterface(req["id"])
	//查询id是否存在
	bInfo := business.NewEntity()
	err = bInfo.FindBusinessItemById(id)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}

	isValid, _, bpInfo, err := determainBusinessMapIsValid(bInfo, req)
	if !isValid {
		return err
	}

	db := orm.Cache(model.DB().Begin())
	//更新业务线表
	bInfo.Db = db
	bUpdate := GetUpdateMap(bInfo, req)

	if _, ok := req["business_name"]; ok {
		bUpdate["name"] = req["business_name"]
	}
	// 更新业务线提币状态
	if _, ok := req["is_withdrawal"]; ok {
		bUpdate["withdrawal_status"] = req["is_withdrawal"]
	} else {
		bUpdate["withdrawal_status"] = 0
	}
	err = bInfo.UpdateBusinessItemByMap(bUpdate)
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnUpdateComboErr, err.Error())
	}
	//更新业务线套餐表
	bpInfo.Db = db
	bpUpdate := GetUpdateMap(bpInfo, req)
	if _, ok := req["deduct_coin_name"]; ok {
		bpUpdate["deduct_coin"] = req["deduct_coin_name"]
	}
	var orderType string
	_, ok := req["trade_type"]
	if ok {
		orderType = req["trade_type"].(string)
	}
	if strings.Contains(orderType, "变更套餐") {
		//变更套餐类型 清除以前数据
		pInfo := _package.NewEntity()
		var tName string
		var mName string
		if _, ok := req["type_name"]; ok {
			tName = req["type_name"].(string)
		}
		if _, ok := req["model_name"]; ok {
			mName = req["model_name"].(string)
		}
		pInfo.FindPackageByTypeModel(tName, mName)
		if pInfo.Id == 0 {
			err = fmt.Errorf("不支持的套餐变更(套餐或收费模式不符)")
			log.Errorf("不支持的套餐变更(套餐或收费模式不符),%v,%v", req["type_name"].(string), req["model_name"].(string))
			return
		}
		bpUpdate["package_id"] = pInfo.Id
		//变更套餐类型 清除以前数据
		bpInfo.UpdateBPWhenChangePId(bpInfo.AccountId, bpInfo.PackageId, pInfo.TypeNums)
	} else if _, ok := bpUpdate["package_id"]; ok {
		delete(bpUpdate, "package_id")
	}

	log.Errorf("GetUpdateMap bpUpdate= %+v\n", bpUpdate)
	err = bpInfo.UpdateBPItemByBusinessIdByMap(bInfo.Id, bpUpdate)
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnUpdateComboErr, err.Error())
	}

	//更新业务线安全表
	ssInfo := serviceSecurity.NewEntity()
	ssInfo.Db = db
	ssUpdate := GetUpdateMap(ssInfo, req)
	err = ssInfo.UpdateItemByBusinessIdByMap(bInfo.Id, ssUpdate)
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnUpdateComboErr, err.Error())
	}
	//创建业务线订单

	if orderType == "open" || strings.Contains(orderType, "开通") {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnUpdateComboErr, "更新业务线不支持此交易类型")
	}
	err = UpdateBusinessCreateNewOrder(user, bInfo, bpInfo, orderType)
	if err != nil {
		db.Rollback()
		return global.OperationErrorText(global.DataWarnCreateComboErr, err.Error())
	}
	//保存操作日志
	err = SaveRecord(db, user, bInfo.Id, record.BusinessRecord, "update")
	if err != nil {
		db.Rollback()
		return
	}
	log.Infof("is_whitelist:%d", req["is_whitelist"])
	if v, ok := req["is_whitelist"]; ok {
		log.Infof("is_whitelist:%d", req["is_whitelist"])
		if v.(float64) > 0 {
			white.CloseWhiteUseById(bInfo.Id, map[string]interface{}{"use": 1})
		}
	}
	err = db.Commit().Error
	return
}

// SearchBusinessItem 搜索业务线详情
func SearchBusinessItem(pid int64) (info domain.BusinessDetailInfo, err error) {

	//查询id是否存在
	var item business.BPDetailInfo
	pInfo := business.NewEntity()
	item, err = pInfo.FindBPDetailItemById(pid)
	if err == gorm.ErrRecordNotFound {
		return info, global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return info, err
	}

	b, errB := json.Marshal(item)
	if errB != nil {
		return info, global.OperationErrorText(errB.Error())
	}

	errC := json.Unmarshal(b, &info)
	if errC != nil {
		return info, global.OperationErrorText(errC.Error())
	}
	boInfo := businessOrder.NewEntity()
	boInfo.FindLatestPackageType(int(pid))
	//info.ProfitNumber = boInfo.ProfitNumber
	info.TradeType = boInfo.OrderType
	info.ApiSecret = item.Secret
	info.DeductCoinName = info.DeductCoin
	return info, nil
}

// SearchBusinessPackageInfo 搜索业务线套餐详情
//func SearchBusinessPackageInfo(pid int64) (info domain.BusinessPackageInfo, err error) {
//	//查询id是否存在
//	pInfo := businessPackage.NewEntity()
//	err = pInfo.FindBPItemByBusinessId(pid)
//	if err == gorm.ErrRecordNotFound {
//		return info, global.WarnMsgError(global.DataWarnNoDataErr)
//	}
//	if err != nil {
//		return info, err
//	}
//
//	b, errB := json.Marshal(pInfo)
//	if errB != nil {
//		return info, global.OperationErrorText(errB.Error())
//	}
//
//	errC := json.Unmarshal(b, &info)
//	if errC != nil {
//		return info, global.OperationErrorText(errC.Error())
//	}
//	boInfo := businessOrder.NewEntity()
//	boInfo.FindLatestPackageType(int(pInfo.BusinessId))
//	//info.ProfitNumber = boInfo.ProfitNumber
//	info.TradeType = boInfo.OrderType
//	return info, nil
//}

//SearchBusinessSecurityInfo 安全信息
func SearchBusinessSecurityInfo(pid int64) (info domain.BusinessSecurityBoolInfo, err error) {
	//查询id是否存在
	var item serviceSecurity.BusinessSecurityDB
	ssInfo := serviceSecurity.NewEntity()
	item, err = ssInfo.FindBindInfoByBusinessId(pid)
	if err == gorm.ErrRecordNotFound {
		return info, global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return info, err
	}
	var isW bool
	var isIp bool
	if item.IsWithdrawal == 1 {
		isW = true
	}
	if item.IsIp == 1 {
		isIp = true
	}
	info = domain.BusinessSecurityBoolInfo{
		ClientId:     item.ClientId,
		Secret:       item.Secret,
		IpAddr:       item.IpAddr,
		CallbackUrl:  item.CallbackUrl,
		IsWithdrawal: isW,
		IsIp:         isIp,
		Phone:        item.Phone,
		Email:        item.Email,
	}

	return info, nil
}

func ResetBusinessClientIdAndSecret(pid int64) (info domain.BusinessSecurityInfo, err error) {
	//查询id是否存在
	var item serviceSecurity.BusinessSecurityDB
	ssInfo := serviceSecurity.NewEntity()
	item, err = ssInfo.FindBindInfoByBusinessId(pid)
	if err == gorm.ErrRecordNotFound {
		return info, global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return info, err
	}
	//更新钱包 client_id的secret
	newMch := domain.BCMchReq{
		ClientId: item.ClientId,
	}
	var mchInfo domain.BCMchInfo
	mchInfo, err = blockChainsApi.BlockChainReSecretClientIdSecret(newMch, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
	if err != nil {
		log.Errorf("重置钱包用户secret err:%v", err.Error())
		return info, global.WarnMsgError(global.DataWarnUpdateDataErr)
	}
	ssInfo.Id = item.Id
	ssInfo.ClientId = mchInfo.ClientId
	ssInfo.Secret = mchInfo.Secret
	err = ssInfo.UpdateItemByBusinessId(pid)
	if err != nil {
		log.Errorf("更新数据库update err:%v", err.Error())
		return info, global.WarnMsgError(global.DataWarnUpdateDataErr)
	}
	item.ClientId = mchInfo.ClientId
	item.Secret = mchInfo.Secret
	b, errB := json.Marshal(item)
	if errB != nil {
		return info, global.OperationErrorText(errB.Error())
	}
	errC := json.Unmarshal(b, &info)
	if errC != nil {
		return info, global.OperationErrorText(errC.Error())
	}
	return info, nil
}

// SearchBusiness 搜索业务线列表
func SearchBusiness(req *domain.BusinessReqInfo) (list []domain.BusinessListInfo, total int64, err error) {

	var l []business.BusinessListDB
	bInfo := business.NewEntity()
	l, total, err = bInfo.FindBusinessListByReq(*req)
	if err != nil {
		return list, total, global.DaoError(err)
	}
	list = make([]domain.BusinessListInfo, 0)
	ppInfo := _package.PackagePay{}
	ppInfo.Db = bInfo.Db
	typeList, _, _ := ppInfo.FindAllPackagePayList()
	//typeList, _, _ := ppInfo.FindAllPackagePayList()
	t := time.Now()
	for _, item := range l {
		b, errB := json.Marshal(item)
		pInfo := domain.BusinessListInfo{}
		if errB == nil {
			json.Unmarshal(b, &pInfo)
		}
		pInfo.AccountId = item.AccountId
		pInfo.Name = item.Name
		pInfo.Email = item.Email
		pInfo.Phone = item.Phone
		pInfo.IsTest = item.AccountStatus
		pInfo.BusinessName = item.BusinessName
		pInfo.BusinessId = item.BusinessId
		pInfo.CreateTime = GetTimeString(item.CreateTime)
		pInfo.Coin = item.Coin
		pInfo.SubCoin = item.SubCoin
		pInfo.TypeName = getTypeEnName(typeList, item.TypeName)
		pInfo.ModelName = item.ModelName
		pInfo.TopUpType = item.TopUpType
		pInfo.TopUpFee = item.TopUpFee
		pInfo.WithdrawalType = item.WithdrawalType
		pInfo.WithdrawalFee = item.WithdrawalFee
		pInfo.CheckerName = item.CheckerName
		pInfo.BusinessStatus = item.BusinessStatus

		uInfo := apply.NewEntity()
		uInfo.FindApplyItemByAccountId(item.AccountId)
		if uInfo.ContractEndAt != nil {
			if !TimeIsNull(*uInfo.ContractEndAt) {
				if t.After(*uInfo.ContractEndAt) && pInfo.BusinessStatus != 2 {
					pInfo.BusinessStatus = 3
				}
			}
		}

		pInfo.CheckedAt = GetTimeString(item.CheckedAt)
		pInfo.Remark = item.Remark
		//查询最新order
		//boInfo := businessOrder.NewEntity()
		//boInfo.FindLatestBusinessOrderItemByBid(int64(item.BusinessId))
		boInfo := businessOrder.NewEntity()
		boInfo.FindLatestPackageType(item.BusinessId)
		pInfo.ProfitNumber = boInfo.ProfitNumber
		pInfo.OrderType = boInfo.OrderType
		list = append(list, pInfo)
	}
	return
}

func OperateBusinessItem(user *domain.JwtCustomClaims, req *domain.BusinessOperateInfo) (err error) {
	pInfo := business.NewEntity()
	err = pInfo.FindBusinessItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if pInfo.State == 2 {
		return global.WarnMsgError(global.DataWarnDataUnableErr)

	}
	if err != nil {
		return global.DaoError(err)
	}
	if req.Operate == "lock" && pInfo.State == 1 {
		return global.WarnMsgError(global.DataWarnHadLockErr)
	}
	if req.Operate == "unlock" && pInfo.State != 1 {
		return global.WarnMsgError(global.DataWarnNoLockErr)
	}

	var operatorName string
	operatorName = user.Name

	pInfo.Id = req.Id
	t := time.Now().Local()
	pInfo.CheckedAt = &t
	pInfo.CheckerId = user.Id
	pInfo.CheckerName = operatorName
	if req.Operate == "lock" {
		pInfo.State = 1
	} else if req.Operate == "unlock" {
		pInfo.State = 0
	} else {
		return global.WarnMsgError(global.DataWarnNoOperateErr)
	}

	//更新数据表
	db := orm.Cache(model.DB().Begin())
	pInfo.Db = db
	err = pInfo.UpdateBusinessState(pInfo.State)
	if err != nil {
		pInfo.Db.Rollback()
		return
	}
	//保存操作日志
	err = SaveRecord(db, user, req.Id, record.BusinessRecord, req.Operate)
	if err != nil {
		db.Rollback()
		return
	}
	err = db.Commit().Error
	return
}

//业务线条件判断
//判断是否满足业务线各种限制
func determainBusinessParametersIsValid(req *domain.CreateBusinessInfo, reqM *map[string]interface{}) (isValid bool, mInfo *merchant.Entity, pbInfo *businessPackage.Entity, err error) {
	if req != nil {
		v := validator.New()
		err = v.Struct(req)
		if err != nil {
			log.Errorf("validate err %v", err)
			for _, err1 := range err.(validator.ValidationErrors) {
				err = global.NewError(global.DataWarnParamErr, err1)
				return
			}
		}
		//if !strings.Contains(req.Coin, req.DeductCoinName) {
		//	err = global.NewError(global.DataWarnParamErr, "业务线未包含扣费币种")
		//	return
		//}
	}

	mInfo = merchant.NewEntity()
	mInfo.FindMerchantItemById(req.AccountId)
	if mInfo.Id == 0 {
		err = global.WarnMsgError(global.DataWarnNoMerchantErr)
		return
	}
	pInfo := _package.NewEntity()
	pInfo.FindPackageItemById(int64(req.PackageId))
	if pInfo.Id == 0 {
		err = global.WarnMsgError(global.DataWarnNoPackageErr)
		return
	}
	isValid = true

	pbInfo = &businessPackage.Entity{
		AccountId:         mInfo.Id,
		PackageId:         int64(pInfo.Id),
		TypeName:          pInfo.TypeName,
		ModelName:         pInfo.ModelName,
		EnterUnit:         pInfo.EnterUnit,
		LimitType:         pInfo.LimitType,
		TypeNums:          pInfo.TypeNums,
		TopUpType:         pInfo.TopUpType,
		TopUpFee:          pInfo.TopUpFee,
		WithdrawalType:    pInfo.WithdrawalType,
		WithdrawalFee:     pInfo.WithdrawalFee,
		ServiceDiscount:   pInfo.ServiceDiscountNums,
		ChainNums:         pInfo.ChainNums,
		ChainDiscountUnit: pInfo.ChainDiscountUnit,
		ChainDiscountNums: pInfo.ChainDiscountNums,
		ChainTimeUnit:     pInfo.ChainTimeUnit,
		CoinNums:          pInfo.CoinNums,
		CoinDiscountUnit:  pInfo.CoinDiscountUnit,
		CoinDiscountNums:  pInfo.CoinDiscountNums,
		CoinTimeUnit:      pInfo.CoinTimeUnit,
		DeployFee:         pInfo.DeployFee,
		CustodyFee:        pInfo.CustodyFee,
		DepositFee:        pInfo.DepositFee,
		AddrNums:          pInfo.AddrNums,
		CoverFee:          pInfo.CoverFee,
		ComboDiscountUnit: pInfo.ComboDiscountUnit,
		ComboDiscountNums: pInfo.ComboDiscountNums,
		YearDiscountUnit:  pInfo.YearDiscountUnit,
		YearDiscountNums:  pInfo.YearDiscountNums,
	}
	rMap := *reqM
	t := reflect.TypeOf(pbInfo)
	fmt.Println(t)
	SetStructFieldByJsonName(pbInfo, rMap)
	return
}

func SetStructFieldByJsonName(ptr interface{}, fields map[string]interface{}) {
	log.Infof("fields:", fields)

	v := reflect.ValueOf(ptr).Elem() // the struct variable

	for i := 0; i < v.NumField(); i++ {

		fieldInfo := v.Type().Field(i) // a reflect.StructField
		tag := fieldInfo.Tag           // a reflect.StructTag
		name := tag.Get("json")

		if name == "" {
			name = strings.ToLower(fieldInfo.Name)
		}
		//去掉逗号后面内容 如 `json:"voucher_usage,omitempty"`
		name = strings.Split(name, ",")[0]
		log.Infof("JSONnAME:", name)

		if value, ok := fields[name]; ok {

			log.Infof("fieldInfo.Name:", fieldInfo.Name)
			//给结构体赋值
			//保证赋值时数据类型一致
			//vTypeName := reflect.ValueOf(value).Type().Name()
			fInfoTypeName := v.FieldByName(fieldInfo.Name).Type()
			log.Errorf("类型1：", reflect.ValueOf(value).Type(), "类型2：", v.FieldByName(fieldInfo.Name).Type())
			if reflect.ValueOf(value).Type() == v.FieldByName(fieldInfo.Name).Type() {
				v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(value))
			} else if fInfoTypeName == reflect.TypeOf(decimal.Decimal{}) {
				valueStr := GetFloat64FromInterface(value)
				vDecimal := decimal.NewFromFloat(valueStr)
				v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(vDecimal))
			}
		}
	}

	return
}

func GetUpdateMap(ptr interface{}, fields map[string]interface{}) (updateMap map[string]interface{}) {
	log.Infof("fields:", fields)
	updateMap = make(map[string]interface{})
	v := reflect.ValueOf(ptr).Elem() // the struct variable

	for i := 0; i < v.NumField(); i++ {

		fieldInfo := v.Type().Field(i) // a reflect.StructField
		tag := fieldInfo.Tag           // a reflect.StructTag
		name := tag.Get("json")

		if name == "" {
			name = strings.ToLower(fieldInfo.Name)
		}
		//去掉逗号后面内容 如 `json:"voucher_usage,omitempty"`
		name = strings.Split(name, ",")[0]
		log.Infof("JSONnAME:", name)
		for kM, vM := range fields {
			if kM == name && kM != "id" {
				updateMap[kM] = vM
			}

		}
	}
	return updateMap
}

func GetUpdateMapByStruct(ptr interface{}) (updateMap map[string]interface{}) {
	updateMap = make(map[string]interface{})
	v := reflect.ValueOf(ptr).Elem() // the struct variable

	for i := 0; i < v.NumField(); i++ {

		fieldInfo := v.Type().Field(i) // a reflect.StructField
		tag := fieldInfo.Tag           // a reflect.StructTag
		name := tag.Get("json")
		value := v.Field(i).Interface()
		if name == "" {
			name = strings.ToLower(fieldInfo.Name)
		}
		//去掉逗号后面内容 如 `json:"voucher_usage,omitempty"`
		name = strings.Split(name, ",")[0]
		log.Infof("JSONnAME:", name)
		if name == "" || name == "-" {
			continue
		}
		updateMap[name] = value
	}
	return updateMap
}

func determainBusinessMapIsValid(bInfo *business.Entity, reqM map[string]interface{}) (isValid bool, mInfo *merchant.Entity, bpInfo *businessPackage.Entity, err error) {

	mInfo = merchant.NewEntity()
	mInfo.FindMerchantItemById(bInfo.AccountId)
	if mInfo.Id == 0 {
		err = global.WarnMsgError(global.DataWarnNoMerchantErr)
		return
	}
	bpInfo = businessPackage.NewEntity()
	bpInfo.FindBPItemByBusinessId(bInfo.Id)
	if bpInfo.Id == 0 {
		err = global.WarnMsgError(global.DataWarnNoPackageErr)
		return
	}
	isValid = true
	SetStructFieldByJsonName(bpInfo, reqM)
	return
}

func SaveRecord(db *orm.CacheDB, user *domain.JwtCustomClaims, id int64, rType, operate string) error {
	//保存操作日志
	rInfo := record.NewEntity()
	rInfo.Db = db
	rInfo.OperatorId = user.Id
	rInfo.OperatorName = user.Name
	rInfo.RecordType = rType
	if rType == record.FinanceRecord {
		rInfo.FinanceId = id
	} else if rType == record.BusinessRecord {
		rInfo.BusinessId = id
	} else if rType == record.ApplyRecord {
		rInfo.MerchantId = id
	}
	rInfo.Operate = operate
	err := rInfo.InsertNewPackage()
	return err
}

func FindServiceCoinChain(mid int64, sid int) ([]business.Entity, error) {
	pInfo := business.NewEntity()
	return pInfo.FindBusinessByAccountId(mid, sid)
}

func FirstServiceCoinChain(mid int64, sid int) (business.Entity, error) {
	pInfo := business.NewEntity()
	return pInfo.FirstServiceCoinChain(mid, sid)
}

func getTypeEnName(list []_package.PackagePay, name string) (enName string) {
	for _, item := range list {
		if item.PayName == name {
			return item.PayType
		}
	}
	return name
}

func GetBindInfoByClientId(clientId string) (*serviceSecurity.Entity, error) {
	pInfo := serviceSecurity.NewEntity()
	err := pInfo.GetBindInfoByClientId(clientId)
	if err != nil {
		return &serviceSecurity.Entity{}, err
	}
	return pInfo, err
}

func FirstServiceBySId(sid int) (business.Entity, error) {
	pInfo := business.NewEntity()
	return pInfo.FirstServiceBySId(sid)
}
