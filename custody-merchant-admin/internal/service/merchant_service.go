package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/apply"
	"custody-merchant-admin/model/finance"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/model/roleMenu"
	"custody-merchant-admin/model/serviceChains"
	"custody-merchant-admin/model/userPermission"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/library"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

//CreateNewMerchantByApply 创建新商户
func CreateNewMerchantByApply(db *orm.CacheDB, item apply.Entity) (userId int64, err error) {
	nT := time.Now().Local()
	salt := "noSalt"
	var password = "HOO@2022"

	//password = library.EncryptSha256Password("Hoo@2022", salt)
	//if err != nil {
	//	log.Error(err)
	//	err = errors.New(global.DataWarnCreateUserErr)
	//	return 0, err
	//}

	pwd := library.EncryptSha256Password(password, salt)
	parse := TimeFromString(item.TestEnd)
	if TimeIsNull(parse) {
		err = fmt.Errorf("%v,时间格式不支持", item.TestEnd)
		return 0, err
	}
	user := merchant.Entity{
		Db:         db,
		Name:       item.Name,
		Phone:      item.Phone,
		Email:      item.Email,
		ApplyId:    item.Id,
		Role:       2,
		Salt:       salt,
		Password:   pwd,
		CreateTime: &nT,
		TestTime:   &parse,
		IsTest:     1, // 测试账户是1，正式是0
	}
	userInfo, err := user.FindMerchantItemByPhoneAndEmail(item.Phone, item.Email)
	if err != nil {
		return 0, err
	}
	if userInfo.Id > 0 {
		return 0, errors.New("手机号/邮箱，已有商户，无法重复注册")
	}
	user.InsertNewMerchant()
	var newUser merchant.Entity
	newUser, err = user.FindMerchantItemByPhone(item.Phone)
	userId = newUser.Id
	if userId == 0 || err != nil {
		log.Error("创建错误")
		err = errors.New(global.DataWarnCreateUserErr)
		return userId, err
	}

	//添加到权限表
	mids, err := roleMenu.SearchAllMid()
	midStr := strings.Join(mids, ",")
	userPermission := userPermission.Entity{
		Db:  db,
		Uid: userId,
		Mid: midStr,
	}
	err = userPermission.InsertNewUserPermission()
	if err != nil {
		err = errors.New(global.DataWarnCreateUserErr)
		return userId, err
	}
	return userId, err
}

func UpdateMerchantItem(user *domain.JwtCustomClaims, req map[string]interface{}) (err error) {
	id := GetIntFromInterface(req["id"])
	pInfo := merchant.NewEntity()
	err = pInfo.FindMerchantItemById(id)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}

	b, err := json.Marshal(req)
	err = json.Unmarshal(b, pInfo)

	if pInfo.IsPush == 1 {
		fInfo := finance.NewEntity()
		fInfo.FindFinanceItemById(id)
		if fInfo.VerifyStatus == "agree" {
			return global.WarnMsgError(global.DataWarnHadVerifySusErr)
		}
	}
	aMap := make(map[string]interface{})
	aInfo := apply.NewEntity()
	aInfo, err = aInfo.FindApplyItemById(pInfo.ApplyId)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}

	Name, nameOk := req["name"].(string)
	Phone, phoneOk := req["phone"].(string)
	Email, emailOk := req["email"].(string)
	passport, passportOk := req["passport_num"].(string)
	identity, identityOk := req["id_card_num"].(string)
	Remark, _ := req["remark"].(string)
	Sex, _ := req["sex"].(int)
	pInfo.Name = Name
	pInfo.Phone = Phone
	pInfo.Email = Email
	pInfo.Remark = Remark
	pInfo.Sex = Sex
	if passportOk {
		pInfo.Passport = passport
	}
	if identityOk {
		pInfo.Identity = identity
	}
	if nameOk {
		aMap["name"] = Name
	}
	if phoneOk {
		aMap["phone"] = Phone
	}
	if emailOk {
		aMap["email"] = Email
	}

	db := orm.Cache(model.DB().Begin())
	pInfo.Db = db
	aInfo.Db = db
	cstart, _ := req["contract_start_at"].(string)
	cend, _ := req["contract_end_at"].(string)
	te, _ := req["test_end"].(string)
	idNum, iOk := req["id_card_num"].(string)
	pNum, pOk := req["passport_num"].(string)

	if cstart != "" {
		sTime := TimeFromString(cstart)
		aMap["contract_start_at"] = &sTime
	}
	if cend != "" {
		eTime := TimeFromString(cend)
		aMap["contract_end_at"] = &eTime
	}
	if te != "" {
		aMap["test_end"] = te
		t := TimeFromString(te)
		pInfo.TestTime = &t
	}
	if iOk {
		aMap["id_card_num"] = idNum
	}
	if pOk {
		aMap["passport_num"] = pNum
	}
	pInfo.IsPush = 0
	err = pInfo.UpdateMerchantItem()
	if err != nil {
		db.Rollback()
		err = errors.New(global.DataWarnUpdateDataErr)
		return err
	}

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
	//aMap := GetUpdateMapByStruct(aInfo)
	err = aInfo.UpdateApplyItemByMap(aMap)
	if err != nil {
		db.Rollback()
		err = errors.New(global.DataWarnUpdateDataErr)
		return err
	}
	//记录编辑日志
	err = db.Commit().Error
	return
}

//func OperateMerchantItem(userId int64,req *domain.MerchantOperateInfo) ( err error) {
//	var item merchant.Entity
//	pInfo := merchant.NewEntity()
//	item,err = pInfo.FindMerchantItemById(req.Id)
//	if err == gorm.ErrRecordNotFound {
//		return global.WarnMsgError(global.DataWarnNoDataErr)
//	}
//	if err != nil {
//		return  global.DaoError(err)
//	}
//	if item.VerifyStatus != "" {
//		return global.WarnMsgError(global.DataWarnHadOperateErr)
//	}
//
//	var operatorName string
//	operatorName = GetOperatorNameById(userId)
//	item.Id = req.Id
//	item.VerifyAt = time.Now()
//	item.VerifyUser =  operatorName
//	if req.Operate == "agree" ||  req.Operate == "refuse" {
//		item.VerifyStatus = req.Operate
//	}else {
//		return global.WarnMsgError(global.DataWarnNoOperateErr)
//	}
//
//
//	item.Db = orm.Cache(model.DB().Begin())
//	err = item.UpdateMerchantItem()
//	if err != nil {
//		item.Db.Rollback()
//		return
//	}
//	//保存操作日志
//	rInfo := record.NewEntity()
//	rInfo.Db = item.Db
//	rInfo.OperatorId = userId
//	rInfo.OperatorName = operatorName
//	rInfo.Operate= req.Operate
//	rInfo.MerchantId = req.Id
//	err = rInfo.InsertNewPackage()
//	if err != nil {
//		rInfo.Db.Rollback()
//		return
//	}
//	rInfo.Db.Commit()
//	return
//}

func FindMerchantBySidCoin(sid int, coin string) (*serviceChains.Entity, error) {
	dao := serviceChains.NewEntity()
	err := dao.FindServiceChainsInfo(sid, coin)
	if err != nil {
		return nil, err
	}
	return dao, err
}
