package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/apply"
	"custody-merchant-admin/model/finance"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/model/record"
	"custody-merchant-admin/module/log"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"strings"
	"time"
)

// SearchApplyList 商户申请列表
func SearchApplyList(req *domain.ApplyReqInfo) (list []domain.ApplyInfo, total int64, err error) {
	var l []apply.ApplyDb
	pInfo := apply.NewEntity()
	l, total, err = pInfo.FindApplyListByReq(*req)
	if err != nil {
		return list, total, global.DaoError(err)
	}
	list = make([]domain.ApplyInfo, 0)
	for _, item := range l {
		b, errB := json.Marshal(item)
		if errB != nil {
			err = errB
			continue
		}
		info := domain.ApplyInfo{}
		errC := json.Unmarshal(b, &info)
		if errC != nil {
			err = errC
			continue
		}
		if item.VerifyStatus == "" {
			info.VerifyStatus = "no_verify"
			info.VerifyResult = ""
		} else {
			info.VerifyStatus = "had_verify"
			info.VerifyResult = item.VerifyStatus
		}
		if item.IsTest == 1 || item.AccountId == 0 {
			info.AccountStatus = "test"
		} else {
			info.AccountStatus = "formal"
		}
		//info.CardType = "company"
		info.CardType = "企业认证"
		info.CreatedAt = GetTimeString(item.CreatedAt)
		if item.VerifyAt != nil {
			info.VerifyAt = GetTimeString(*item.VerifyAt)
		}
		if item.ContractStartAt != nil {
			info.ContractStartAt = GetTimeString(*item.ContractStartAt)
		}
		if item.ContractEndAt != nil {
			info.ContractEndAt = GetTimeString(*item.ContractEndAt)
		}
		//info.ContractStartAt = GetTimeString(*item.ContractStartAt)
		//info.ContractEndAt = GetTimeString(*item.ContractEndAt)
		list = append(list, info)
	}
	return
}

func OperateApplyItem(user *domain.JwtCustomClaims, req *domain.MerchantOperateInfo) (err error) {

	pInfo := apply.NewEntity()
	pInfo, err = pInfo.FindApplyItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}
	if pInfo.VerifyStatus != "" {
		return global.WarnMsgError(global.DataWarnHadVerifyErr)
	}

	operatorName := user.Name
	pInfo.TestEnd = req.TestEnd
	//创建新用户
	var accountId int64
	db := orm.Cache(model.DB().Begin())
	if req.Operate == "agree" {
		newId, err := CreateNewMerchantByApply(db, *pInfo)
		accountId = newId
		if accountId == 0 || err != nil {
			db.Rollback()
			log.Error(err.Error())
			err = errors.New(global.DataWarnCreateUserErr)
			return err
		}
		//设置权限
		err = SetNewMerchantCasbin(db, accountId)
		if err != nil {
			db.Rollback()
			log.Error(err.Error())
			err = errors.New(global.DataWarnCreateUserErr)
			return err
		}
	}

	//修改审核状态
	pInfo.Db = db
	pInfo.Id = req.Id
	pInfo.TestEnd = req.TestEnd
	pInfo.AccountId = accountId
	t := time.Now()
	pInfo.VerifyAt = &t
	pInfo.VerifyUser = operatorName
	if req.Operate == "agree" || req.Operate == "refuse" {
		pInfo.VerifyStatus = req.Operate
	} else {
		db.Rollback()
		return global.WarnMsgError(global.DataWarnNoOperateErr)
	}

	err = pInfo.UpdateApplyItem()
	if err != nil {
		db.Rollback()
		return
	}
	//保存操作日志
	err = SaveRecord(db, user, req.Id, record.ApplyRecord, req.Operate)
	if err != nil {
		db.Rollback()
		return
	}

	db.Commit()
	return
}

func SearchApplyImageInfo(req *domain.MerchantImageInfo) (i interface{}, err error) {

	pInfo := apply.NewEntity()
	pInfo, err = pInfo.FindApplyItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		return i, global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return i, global.DaoError(err)
	}
	backMap := make(map[string]interface{})

	backMap["identity"] = pInfo.IdCardPicture
	backMap["business"] = pInfo.BusinessPicture
	backMap["contract"] = pInfo.ContractPicture

	return backMap, err

}

func SearchMerchantImageInfo(req *domain.MerchantImageInfo) (i interface{}, err error) {

	mInfo := merchant.NewEntity()
	err = mInfo.FindMerchantItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		return i, global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return i, global.DaoError(err)
	}

	pInfo := apply.NewEntity()
	pInfo, _ = pInfo.FindApplyItemById(mInfo.ApplyId)
	//if err == gorm.ErrRecordNotFound {
	//	return i, global.WarnMsgError(global.DataWarnNoDataErr)
	//}
	//if err != nil {
	//	return i, global.DaoError(err)
	//}
	backMap := make(map[string]interface{})

	backMap["identity"] = pInfo.IdCardPicture
	backMap["business"] = pInfo.BusinessPicture
	backMap["contract"] = pInfo.ContractPicture
	if pInfo.ContractStartAt != nil {
		backMap["contract_start_at"] = GetTimeString(*pInfo.ContractStartAt)
	} else {
		backMap["contract_start_at"] = ""
	}
	if pInfo.ContractEndAt != nil {
		backMap["contract_end_at"] = GetTimeString(*pInfo.ContractEndAt)
	} else {
		backMap["contract_end_at"] = ""
	}

	return backMap, err

}

func SearchMerchantInfo(req *domain.MerchantOperateInfo) (i domain.MerchantUpdateInfo, err error) {
	mInfo := merchant.NewEntity()
	err = mInfo.FindMerchantItemById(req.Id)
	if err == gorm.ErrRecordNotFound {
		return i, global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return i, global.DaoError(err)
	}
	i.PhoneCode = mInfo.PhoneCode
	i.Id = mInfo.Id
	i.Name = mInfo.Name
	i.Phone = mInfo.Phone
	i.Email = mInfo.Email
	i.Sex = mInfo.Sex
	i.Remark = mInfo.Remark
	aInfo := apply.NewEntity()
	aInfo, _ = aInfo.FindApplyItemById(mInfo.ApplyId)

	if aInfo.ContractStartAt != nil {
		i.ContractStartAt = GetTimeString(*aInfo.ContractStartAt)
	}
	if aInfo.ContractEndAt != nil {
		i.ContractEndAt = GetTimeString(*aInfo.ContractEndAt)
	}
	i.TestEnd = aInfo.TestEnd
	i.IdCardNum = aInfo.IdCardNum
	i.PassportNum = aInfo.PassportNum

	return i, err

}

func SearchMerchantList(req *domain.MerchantReqInfo) (list []domain.MerchantListInfo, total int64, err error) {
	var l []merchant.MerchantApply
	mInfo := merchant.NewEntity()
	l, total, err = mInfo.FindMerchantListByReq(*req)
	if err != nil {
		return list, total, global.DaoError(err)
	}
	list = make([]domain.MerchantListInfo, 0)
	for i, item := range l {
		b, errB := json.Marshal(item)
		if errB != nil {
			err = errB
			continue
		}
		info := domain.MerchantListInfo{}
		errC := json.Unmarshal(b, &info)
		if errC != nil {
			err = errC
			continue
		}
		if item.VerifyStatus == "" {
			info.FvStatus = "wait"
		} else {
			info.FvStatus = item.VerifyStatus
		}

		info.CreatedAt = GetTimeString(item.CreatedAt)
		info.RealNameAt = GetTimeString(item.RealNameAt)
		info.ContractStartAt = GetTimeString(item.ContractStartAt)
		info.ContractEndAt = GetTimeString(item.ContractEndAt)
		info.SerialNo = i + 1
		info.AccountId = info.Id
		list = append(list, info)
	}
	return
}

func PushMerchantToFinanceVerify(mid int64) (err error) {
	pInfo := merchant.NewEntity()
	err = pInfo.FindMerchantItemById(mid)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}
	if pInfo.IsPush != 0 {
		return global.WarnMsgError(global.DataWarnHadToFinanceErr)
	}

	aInfo := apply.NewEntity()
	aInfo, err = aInfo.FindApplyItemById(pInfo.ApplyId)

	if aInfo.BusinessPicture == "" || aInfo.IdCardPicture == "" {
		return global.WarnMsgError(global.DataWarnNoImageErr)
	}
	if aInfo.ContractPicture == "" {
		return global.WarnMsgError(global.DataWarnNoContractErr)
	}

	db := orm.Cache(model.DB().Begin())
	pInfo.Db = db
	pInfo.IsPush = 1
	err = pInfo.UpdateMerchantItem()
	if err != nil {
		db.Rollback()
		return global.DaoError(err)
	}
	//插入到财务审核表
	fInfo := finance.NewEntity()
	fInfo.Db = db
	fInfo.AccountId = pInfo.Id
	fInfo.ApplyId = pInfo.ApplyId
	err = fInfo.InsertNewFinance()
	if err != nil {
		db.Rollback()
		return global.DaoError(err)
	}
	//提交到财务中台

	arr := []merchant.Entity{*pInfo}
	err = PushToFinancialSystem(arr)
	if err != nil {
		db.Rollback()
		return global.DaoError(err)
	}
	err = db.Commit().Error
	return

}

//PushBatchApplysToFinanceVerify 批量推送到财务审核
func PushBatchApplysToFinanceVerify() (err error) {

	pInfo := merchant.NewEntity()
	var items []merchant.MerchantApply
	items, err = pInfo.FindPushAbleApplies()
	if len(items) <= 0 {
		return global.WarnMsgError(global.DataWarnNoPushUserErr)
	}
	ids := make([]int64, 0)
	for _, item := range items {
		ids = append(ids, item.AccountId)
	}
	// 更新推送状态
	db := orm.Cache(model.DB().Begin())
	pInfo.Db = db
	err = pInfo.UpdatePushAbleApplies(ids)
	if err != nil {
		db.Rollback()
		return global.DaoError(err)
	}

	//批量插入到财务审核表
	fInfo := finance.NewEntity()
	fInfo.Db = db
	err = fInfo.InsertPatchFinance(items)
	if err != nil {
		db.Rollback()
		return global.WarnMsgError(global.DataWarnBatchPushFinanceErr)
	}

	//提交到财务中台
	arr := make([]merchant.Entity, 0)
	var b []byte
	b, err = json.Marshal(items)
	if err != nil {
		db.Rollback()
		log.Errorf("json.Marshal err :%v", err)
		return global.WarnMsgError(global.DataWarnBatchPushFinanceErr)
	}
	err = json.Unmarshal(b, &arr)
	if err != nil {
		db.Rollback()
		log.Errorf("json.Unmarshal err :%v", err)
		return global.WarnMsgError(global.DataWarnBatchPushFinanceErr)
	}
	err = PushToFinancialSystem(arr)
	if err != nil {
		db.Rollback()
		log.Errorf("PushToFinancialSystem err :%v", err)
		return global.WarnMsgError(global.DataWarnBatchPushFinanceErr)
	}
	err = pInfo.Db.Commit().Error
	return

}

func UpdateMerchantImage(req *domain.MerchantImgInfo) error {

	pInfo := apply.NewEntity()
	up := 0
	mp := map[string]interface{}{}
	if len(req.ContractPicture) > 0 {
		up += 1
		mp["contract_picture"] = strings.Join(req.ContractPicture, ",")
	}
	if len(req.BusinessPicture) > 0 {
		up += 1
		mp["business_picture"] = strings.Join(req.BusinessPicture, ",")
	}
	if len(req.IdCardPicture) > 0 {
		up += 1
		mp["id_card_picture"] = strings.Join(req.IdCardPicture, ",")
	}
	if up != 0 {
		err := pInfo.UpdateApplyMap(req.Id, mp)
		if err != nil {
			return err
		}
	}
	return nil
}
