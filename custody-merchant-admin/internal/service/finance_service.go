package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/apply"
	"custody-merchant-admin/model/assets"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/business"
	"custody-merchant-admin/model/businessOrder"
	"custody-merchant-admin/model/finance"
	"custody-merchant-admin/model/financeAssets"
	"custody-merchant-admin/model/financeFlow"
	"custody-merchant-admin/model/fullYear"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/model/record"
	"custody-merchant-admin/module/log"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

func SearchPushApplyList(req *domain.MerchantReqInfo) (list []domain.FinanceListInfo, total int64, err error) {
	var l []finance.FinanceListDB
	fInfo := finance.NewEntity()
	l, total, err = fInfo.FindPushFinanceListByReq(*req)
	if err != nil {
		return list, total, global.DaoError(err)
	}
	list = make([]domain.FinanceListInfo, 0)
	for _, item := range l {
		b, errB := json.Marshal(item)
		if errB != nil {
			err = errB
			continue
		}
		info := domain.FinanceListInfo{}
		errC := json.Unmarshal(b, &info)
		if errC != nil {
			err = errC
			continue
		}
		if item.FvStatus == "" {
			info.FvStatus = "wait"
		} else {
			info.FvStatus = item.FvStatus
		}
		if item.RealNameStatus == "1" {
			info.RealNameStatus = "had_real"
		} else {
			info.RealNameStatus = "no_real"
		}
		if item.IsLockFinance == 1 {
			info.LockStatus = "lock"
		} else if item.IsLock == 0 {
			info.LockStatus = "unlock"
		} else {
			info.LockStatus = "lock"
		}
		info.CreatedAt = GetTimeString(item.CreatedAt)
		info.RealNameAt = GetTimeString(item.RealNameAt)
		info.ContractStartAt = GetTimeString(item.ContractStartAt)
		info.ContractEndAt = GetTimeString(item.ContractEndAt)
		list = append(list, info)
	}
	return
}

func GetFinanceVerifyImage(req *domain.ApplyImageReqInfo) (data domain.ApplyImageReqInfo, err error) {

	//查询id是否存在
	fInfo := finance.NewEntity()
	err = fInfo.FindFinanceItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		err = global.WarnMsgError(global.DataWarnNoDataErr)
		return
	}
	if err != nil {
		err = global.DaoError(err)
		return
	}
	aInfo := apply.NewEntity()
	aInfo, err = aInfo.FindApplyItemById(fInfo.ApplyId)
	var start string
	var end string
	if aInfo.ContractStartAt != nil {
		start = GetTimeString(*aInfo.ContractStartAt)
	}
	if aInfo.ContractEndAt != nil {
		end = GetTimeString(*aInfo.ContractEndAt)
	}
	data = domain.ApplyImageReqInfo{
		IdCardPicture:   aInfo.IdCardPicture,
		BusinessPicture: aInfo.BusinessPicture,
		ContractPicture: aInfo.ContractPicture,
		ContractStartAt: start,
		ContractEndAt:   end,
	}
	return
}

// UpdateFinanceVerifyImage 编辑商户
func UpdateFinanceVerifyImage(user *domain.JwtCustomClaims, req map[string]interface{}) (err error) {
	id := GetIntFromInterface(req["id"])
	//查询id是否存在
	fInfo := finance.NewEntity()
	err = fInfo.FindFinanceItemById(id)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}

	if fInfo.VerifyStatus != "" {
		return global.WarnMsgError(global.DataWarnHadVerifySusErr)
	}
	db := orm.Cache(model.DB().Begin())

	aInfo := apply.NewEntity()
	aInfo.Id = fInfo.ApplyId
	aInfo.Db = db
	aMap := make(map[string]interface{})
	if _, ok := req["business"]; ok {
		arr := req["business"].([]interface{})
		arrS := GetStringFromInterfaceArr(arr)
		aMap["business_picture"] = arrS
	}
	if _, ok := req["contract"]; ok {
		arr := req["contract"].([]interface{})
		arrS := GetStringFromInterfaceArr(arr)
		aMap["contract_picture"] = arrS

	}
	if _, ok := req["identity"]; ok {
		arr := req["identity"].([]interface{})
		arrS := GetStringFromInterfaceArr(arr)
		aMap["id_card_picture"] = arrS
	}
	cstart, _ := req["contract_start_at"].(string)
	cend, _ := req["contract_end_at"].(string)

	if cstart != "" {
		sTime := TimeFromString(cstart)
		aMap["contract_start_at"] = &sTime
	}
	if cend != "" {
		eTime := TimeFromString(cend)
		aMap["contract_end_at"] = &eTime
	}

	err = aInfo.UpdateApplyItemByMap(aMap)
	if err != nil {
		db.Rollback()
		log.Errorf("更新财务审核信息出错err :%v", err)
		return global.WarnMsgError(global.DataWarnUpdateDataErr)
	}
	//保存操作日志
	err = SaveRecord(db, user, id, record.FinanceRecord, "update")
	if err != nil {
		db.Rollback()
		return
	}
	err = db.Commit().Error
	return
}

func FinanceAgreeRefuse(user *domain.JwtCustomClaims, req *domain.FinanceOperateInfo) (err error) {

	//查询id是否存在
	fInfo := finance.NewEntity()
	err = fInfo.FindFinanceItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		log.Errorf("财务审核回调失败，数据不存在：req:%v", req)
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		log.Errorf("财务审核回调失败，数据不存在：req:%v \n,err:%v", req, err)
		return global.DaoError(err)
	}
	db := orm.Cache(model.DB().Begin())
	fInfo.Db = db
	fInfo.VerifyStatus = req.Operate
	t := time.Now().Local()
	fInfo.VerifyAt = &t
	fInfo.VerifyUser = req.OperateName
	err = fInfo.UpdateFinanceItem()
	if err != nil {
		db.Rollback()
		return global.WarnMsgError(global.DataWarnUpdateDataErr)
	}
	//记录满年开始时间
	aInfo := apply.NewEntity()
	aInfo, err = aInfo.FindApplyItemById(fInfo.ApplyId)
	if aInfo.Id == 0 {
		db.Rollback()
		log.Errorf("财务审核 记录满年开始时间err:%v", err)
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	fyInfo := fullYear.NewEntity()
	fyInfo.Db = db
	fyInfo.AccountId = fInfo.AccountId
	fyInfo.PackageId = 0
	fyInfo.BusinessId = 0
	fyInfo.LatestTime = *aInfo.ContractStartAt
	err = fyInfo.InsertNewItem()
	if err != nil {
		db.Rollback()
		log.Errorf("财务审核 记录满年开始时间err:%v", err)
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	//保存操作日志
	err = SaveRecord(db, user, req.Id, record.FinanceRecord, req.Operate)
	if err != nil {
		db.Rollback()
		return
	}
	err = db.Commit().Error
	// 商户账号修改为正式账号
	mDao := merchant.NewEntity()
	err = mDao.UpdatePersonalUser(fInfo.AccountId, map[string]interface{}{"is_test": 0})
	if err != nil {
		return err
	}
	// 修改子账号为正式账号
	err = mDao.UpdatePersonalSubUser(fInfo.AccountId, map[string]interface{}{"is_test": 0})
	if err != nil {
		return err
	}

	return
}

//UpdateFinanceLock 财务冻结
func UpdateFinanceLock(user *domain.JwtCustomClaims, req *domain.FinanceOperateInfo) (err error) {

	//查询id是否存在
	aInfo := finance.NewEntity()
	err = aInfo.FindFinanceItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}

	db := orm.Cache(model.DB().Begin())
	//冻结账户/资产
	bInfo := business.NewEntity()
	mInfo := merchant.NewEntity()
	aInfo.Db = db
	bInfo.Db = db
	mInfo.Db = db
	var err1, err2, err3 error
	operate := req.Operate
	//arr := strings.Split(req.Operate, ",")
	//for _, operate := range arr {
	if operate == "lock_user" { //冻结用户和资产
		err1 = bInfo.LockBusinessItemByAccountId(aInfo.AccountId)
		err2 = mInfo.LockMerchantItemById(aInfo.AccountId)
		err3 = aInfo.OperateFinanceItemByFid(req.Id, operate)
	} else if operate == "lock_asset" { //冻结资产
		err1 = bInfo.LockBusinessItemByAccountId(aInfo.AccountId)
		err3 = aInfo.OperateFinanceItemByFid(req.Id, operate)
	} else if operate == "unlock_user" { //解冻用户和资产
		err1 = bInfo.UnlockBusinessItemByAccountId(aInfo.AccountId)
		err2 = mInfo.UnlockMerchantItemById(aInfo.AccountId)
		err3 = aInfo.OperateFinanceItemByFid(req.Id, operate)
	} else if operate == "unlock_asset" { //解冻资产
		if aInfo.IsLock == 1 {
			db.Rollback()
			return global.WarnMsgError(global.DataWarnUnlockAccountLockErr)
		}
		err1 = bInfo.UnlockBusinessItemByAccountId(aInfo.AccountId)
		err3 = aInfo.OperateFinanceItemByFid(req.Id, req.Operate)
	} else {
		db.Rollback()
		return global.WarnMsgError(global.DataWarnNoOperateErr)
	}
	if err1 != nil || err2 != nil || err3 != nil {
		db.Rollback()
		return global.WarnMsgError(global.DataWarnUpdateLockErr)
	}
	//保存操作日志
	err = SaveRecord(db, user, req.Id, record.FinanceRecord, req.Operate)
	if err != nil {
		db.Rollback()
		return
	}
	//}

	err = db.Commit().Error

	return
}

//SearchFinanceLockRecordList 冻结详情列表
func SearchFinanceLockRecordList(req *domain.MerchantReqInfo) (list []domain.FinanceRecordInfo, total int64, err error) {
	return SearchFinanceListByReq(req)
}

//SyncToFinancialSystem 更新同步到财务系统
func SyncToFinancialSystem(fInfo finance.Entity) (err error) {
	//TODO:更新同步到财务
	return
}

//PushToFinancialSystem 提交审核到财务中台
func PushToFinancialSystem(mInfo []merchant.Entity) (err error) {
	//TODO:提交到财务中台
	return
}

//SyncToFinanceAssetsByBusinessOrder 业务线订单付款-财务表记录
func SyncToFinanceAssetsByBusinessOrder(bInfo *businessOrder.Entity, aInfo *assets.Assets) (err error) {
	//币信息
	var cInfo *base.CoinInfo
	cInfo, err = base.FindCoinsById(int(bInfo.OrderCoinId))
	var coinName, subname string
	if cInfo.FullName == "" {
		coinName = cInfo.Name
	} else {
		coinName = cInfo.FullName
		subname = cInfo.Name
	}

	//记录到财务账号
	fInfo := financeAssets.NewEntity()
	fInfo.Db = bInfo.Db
	fInfo.CoinId = bInfo.OrderCoinId
	fInfo.BusinessId = int64(aInfo.ServiceId)
	err = fInfo.FindSameBusinessId(bInfo.OrderCoinId, int64(aInfo.ServiceId))
	if fInfo.Id == 0 {
		fInfo.Nums = bInfo.ProfitNumber
		fInfo.CoinId = bInfo.OrderCoinId
		fInfo.Coin = coinName
		fInfo.SubCoin = subname
		fInfo.Token = cInfo.Token
		fInfo.BusinessId = int64(aInfo.ServiceId)
		//fInfo.Address = aInfo.ChainAddress
		err = fInfo.InsertNewItem()
	} else {
		err = fInfo.UpdateFinanceAssetsAddNumsByBid(bInfo.OrderCoinId, int64(aInfo.ServiceId), bInfo.ProfitNumber)
	}
	if err != nil {
		return err
	}

	//记录流水
	ffInfo := financeFlow.NewEntity()
	ffInfo.Db = bInfo.Db
	ffInfo.OrderId = bInfo.OrderId
	ffInfo.FlowType = "in"
	ffInfo.BusinessId = int64(aInfo.ServiceId)
	//ffInfo.FromAddress = aInfo.ChainAddress
	ffInfo.Nums = bInfo.ProfitNumber
	ffInfo.CoinId = bInfo.OrderCoinId
	ffInfo.CoinName = bInfo.OrderCoinName
	ffInfo.Token = cInfo.Token

	err = ffInfo.InsertNewItem()
	return err
}

//SyncToFinanceAssetsByWithdraw 商户提现手续费-财务表记录
/*
db =数据mysql事务
coinId 币id
nums 财务收入数量
orderId 订单id
address 财务收入来源地址

回参
flowId 流水id，用户回滚
err
*/
func SyncToFinanceAssetsByWithdraw(db *orm.CacheDB, coinId int64, nums decimal.Decimal, orderId string, bid int64) (flowId int, err error) {
	//币信息
	var cInfo *base.CoinInfo
	cInfo, err = base.FindCoinsById(int(coinId))
	var coinName, subname string
	if cInfo.FullName == "" {
		coinName = cInfo.Name
	} else {
		coinName = cInfo.FullName
		subname = cInfo.Name
	}

	//记录到财务账号
	fInfo := financeAssets.NewEntity()
	fInfo.Db = db
	fInfo.CoinId = coinId
	fInfo.BusinessId = bid
	err = fInfo.FindSameBusinessId(coinId, bid)
	if fInfo.Id == 0 {
		fInfo.Nums = nums
		fInfo.CoinId = coinId
		fInfo.Coin = coinName
		fInfo.SubCoin = subname
		fInfo.Token = cInfo.Token
		fInfo.BusinessId = bid
		//fInfo.Address = address
		err = fInfo.InsertNewItem()
	} else {
		err = fInfo.UpdateFinanceAssetsAddNumsByBid(coinId, bid, nums)
	}
	if err != nil {
		return
	}
	//记录流水
	ffInfo := financeFlow.NewEntity()
	ffInfo.Db = db
	ffInfo.OrderId = orderId
	ffInfo.FlowType = "in"
	ffInfo.BusinessId = bid
	//ffInfo.FromAddress = address
	ffInfo.Nums = nums
	ffInfo.CoinId = coinId
	ffInfo.CoinName = coinName
	ffInfo.Token = cInfo.Token
	err = ffInfo.InsertNewItem()
	if err != nil {
		return
	}
	err = ffInfo.FindItemByOrderId(orderId)
	if err != nil {
		return
	}
	flowId = int(ffInfo.Id)
	return
}

//SyncToFinanceAssetsByRecharge 商户充值手续费-财务表记录
/*
db =数据mysql事务
coinId 币id
nums 财务收入数量
orderId 订单id
address 财务收入来源地址
回参
flowId 流水id，用户回滚
err
*/
func SyncToFinanceAssetsByRecharge(db *orm.CacheDB, coinId int64, nums decimal.Decimal, orderId string, bid int64) (flowId int, err error) {
	//币信息
	var cInfo *base.CoinInfo
	cInfo, err = base.FindCoinsById(int(coinId))
	var coinName, subname string
	if cInfo.FullName == "" {
		coinName = cInfo.Name
	} else {
		coinName = cInfo.FullName
		subname = cInfo.Name
	}

	//记录到财务账号
	fInfo := financeAssets.NewEntity()
	fInfo.Db = db
	fInfo.CoinId = coinId
	fInfo.BusinessId = bid
	err = fInfo.FindSameBusinessId(coinId, bid)
	if fInfo.Id == 0 {
		fInfo.Nums = nums
		fInfo.CoinId = coinId
		fInfo.Coin = coinName
		fInfo.SubCoin = subname
		fInfo.Token = cInfo.Token
		fInfo.BusinessId = bid
		err = fInfo.InsertNewItem()
	} else {
		err = fInfo.UpdateFinanceAssetsAddNumsByBid(coinId, bid, nums)
	}
	if err != nil {
		return
	}
	//记录流水
	ffInfo := financeFlow.NewEntity()
	ffInfo.Db = db
	ffInfo.OrderId = orderId
	ffInfo.FlowType = "in"
	ffInfo.BusinessId = bid
	ffInfo.Nums = nums
	ffInfo.CoinId = coinId
	ffInfo.CoinName = coinName
	ffInfo.Token = cInfo.Token
	err = ffInfo.InsertNewItem()
	if err != nil {
		return
	}
	err = ffInfo.FindItemByOrderId(orderId)
	if err != nil {
		return
	}
	flowId = int(ffInfo.Id)
	return
}

//RollbackFinanceAssetsByOrderId 根据流水id 回滚财务，
func RollbackFinanceAssetsByOrderId(db *orm.CacheDB, flowId int) (err error) {
	//币信息
	fInfo := financeFlow.NewEntity()
	err = fInfo.FindItemById(flowId)
	if fInfo.Id == 0 {
		log.Errorf("流水不存在 err:%v", err)
		err = errors.New("流水不存在")
		return
	}
	var iscreate bool
	if db == nil {
		iscreate = true
		db = orm.Cache(model.DB().Begin())
	}
	faInfo := financeAssets.NewEntity()
	faInfo.Db = db
	err = faInfo.FindSameBusinessId(fInfo.CoinId, fInfo.BusinessId)
	if err != nil || faInfo.Id == 0 {
		log.Errorf("该流水财务地址不存在 err:%v", err)
		err = errors.New("该流水财务地址不存在")
		return
	}
	if fInfo.FlowType == "in" {
		err = faInfo.UpdateFinanceAssetsSubNumsByBid(fInfo.CoinId, fInfo.BusinessId, fInfo.Nums)
	} else if fInfo.FlowType == "out" {
		err = faInfo.UpdateFinanceAssetsAddNumsByBid(fInfo.CoinId, fInfo.BusinessId, fInfo.Nums)
	}
	if err != nil {
		if !iscreate {
			db.Rollback()
		}
		return
	}
	fInfo.Db = db
	err = fInfo.DeleteItemById(flowId)
	if err != nil {
		if !iscreate {
			db.Rollback()
		}
	}
	return
}

//RollbackFinanceAssetsByFlowId 根据流水id 回滚财务，
func RollbackFinanceAssetsByFlowId(db *orm.CacheDB, flowId int) (err error) {
	//币信息
	fInfo := financeFlow.NewEntity()
	err = fInfo.FindItemById(flowId)
	if fInfo.Id == 0 {
		log.Errorf("流水不存在 err:%v", err)
		err = errors.New("流水不存在")
		return
	}
	var iscreate bool
	if db == nil {
		iscreate = true
		db = orm.Cache(model.DB().Begin())
	}
	faInfo := financeAssets.NewEntity()
	faInfo.Db = db
	err = faInfo.FindSameBusinessId(fInfo.CoinId, fInfo.BusinessId)
	if err != nil || faInfo.Id == 0 {
		log.Errorf("该流水财务地址不存在 err:%v", err)
		err = errors.New("该流水财务地址不存在")
		return
	}
	if fInfo.FlowType == "in" {
		err = faInfo.UpdateFinanceAssetsSubNumsByBid(fInfo.CoinId, fInfo.BusinessId, fInfo.Nums)
	} else if fInfo.FlowType == "out" {
		err = faInfo.UpdateFinanceAssetsAddNumsByBid(fInfo.CoinId, fInfo.BusinessId, fInfo.Nums)
	}
	if err != nil {
		if !iscreate {
			db.Rollback()
		}
		return
	}
	fInfo.Db = db
	err = fInfo.DeleteItemById(flowId)
	if err != nil {
		if !iscreate {
			db.Rollback()
		}
	}
	return
}
