package business

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/module/log"
	"gorm.io/gorm"
)

func (e *Entity) InsertNewBusiness() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SaveBusinessInfo error: %v", err)
	}
	return
}

func (e *Entity) DeleteBusinessItem(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Delete(map[string]interface{}{"id": pId}).Error
	if err != nil {
		log.Errorf("DelBusinessInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateBusinessState(state int) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ?", e.Id).Update("state", state).Error
	if err != nil {
		log.Errorf("UpdateBusinessInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateBusinessItemByMap(uMap map[string]interface{}) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? ", e.Id).Updates(uMap).Error
	if err != nil {
		log.Errorf("UpdateBusinessInfo error: %v", err)
	}
	return
}

func (e *Entity) FindBusinessItemById(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? ", pId).First(e).Error
	if err != nil {
		log.Errorf("FindBusinessOneInfo error: %v", err)
	}
	if e.Name == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindBPDetailItemById(pId int64) (item BPDetailInfo, err error) {
	selectSql := e.Db.Table(e.TableName()).Select("u.id as account_id,u.name as name," +
		"service.name as business_name ,service.coin as coin,service.sub_coin as sub_coin," +
		"service.checked_at as checked_at,service.remark as remark," +
		"service.checker_name as checker_name,service.state as business_status,service.create_time as create_time," +
		"service.id as id, service.phone as phone,service.email as email,service.account_status as account_status," +
		"a.type_name as type_name ,a.model_name as model_name,a.deduct_coin," +
		"a.chain_discount_unit,a.coin_discount_unit,a.combo_discount_unit,a.year_discount_unit," +
		"a.cover_fee,a.custody_fee,a.deploy_fee,a.deposit_fee,a.year_discount_nums," +
		"ss.is_account_check,ss.is_platform_check,ss.is_whitelist," +
		"a.top_up_type ,a.top_up_fee,a.addr_nums,a.chain_discount_nums,a.coin_discount_nums,a.combo_discount_nums," +
		"a.withdrawal_type ,a.withdrawal_fee," +
		"ss.client_id,ss.secret,ss.is_email,ss.is_sms,ss.is_ip,ss.is_get_addr,ss.is_withdrawal,ss.callback_url,ss.ip_addr")
	selectSql = selectSql.Joins("left join service_security ss on ss.business_id = service.id ")
	selectSql = selectSql.Joins("left join service_combo a on a.business_id = service.id ")
	selectSql = selectSql.Joins("left join user_info u on u.id = service.account_id ")

	err = selectSql.Where("service.id = ? ", pId).First(&item).Error
	if err != nil {
		log.Errorf("FindBusinessOneInfo error: %v", err)
	}
	return
}

func (e *Entity) FindBusinessItemByAccountId(aId int64) (item Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ? ", aId).Order("id desc").First(&item).Error
	if err != nil {
		log.Errorf("FindBusinessOneInfo error: %v", err)
	}
	return
}

func (e *Entity) FindBusinessListByReq(req domain.BusinessReqInfo) (list []BusinessListDB, total int64, err error) {
	//bp := businessPackage.Entity{} ,id ,account_id ,name ,email ,phone ,account_status ,
	//business_name ,business_id ,create_time ,coin ,sub_coin ,type_name ,model_name ,package_type ,
	//profit_number ,order_type ,top_up_type ,top_up_fee ,
	//withdrawal_type ,withdrawal_fee ,checker_name ,business_status  ,checker_time ,remark
	selectSql := e.Db.Table(e.TableName()).Select("u.id as account_id,u.name as name, service.name as business_name ,service.coin as coin," +
		"service.sub_coin as sub_coin,service.checked_at as checked_at,service.remark as remark," +
		"service.checker_name as checker_name,service.state as business_status,service.create_time as create_time," +
		"service.id as business_id, service.phone as phone,service.email as email,service.account_status as account_status," +
		"a.type_name as type_name ,a.model_name as model_name," +
		"a.top_up_type as top_up_type,a.top_up_fee as top_up_fee," +
		//"aso.order_type as order_type," +
		"a.withdrawal_type as withdrawal_type ,a.withdrawal_fee as withdrawal_fee")
	selectSql = selectSql.Joins("left join user_info u on u.id = service.account_id ")
	selectSql = selectSql.Joins("left join service_combo a on a.business_id = service.id ")
	//selectSql = selectSql.Joins("left join admin_service_order aso on aso.business_d = service.id ")
	if req.AccountId != 0 {
		selectSql = selectSql.Where("service.account_id = ?", req.BusinessId)
	}
	if req.BusinessId != 0 {
		selectSql = selectSql.Where("service.id = ?", req.BusinessId)
	}
	if req.ContactStr != "" {
		//通过手机/邮箱获取用户ID
		mInfo := merchant.NewEntity()
		mItem, _ := mInfo.FindMerchantItemByContactStr(req.ContactStr)
		accountId := mItem.Id
		selectSql = selectSql.Where("service.account_id = ?", accountId)
	}

	selectSql.Count(&total)
	err = selectSql.Limit(req.Limit).Offset(req.Offset).Order("service.id desc").Find(&list).Error
	if err != nil {
		log.Errorf("FindBusinessList error: %v", err)
	}
	return
}

//LockBusinessItemByAccountId 冻结用户所有业务线
func (e *Entity) LockBusinessItemByAccountId(userId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ?", userId).Update("state", 1).Error
	if err != nil {
		log.Errorf("FindBusinessOneInfo error: %v", err)
	}
	return
}

//UnlockBusinessItemByAccountId 解冻用户所有业务线
func (e *Entity) UnlockBusinessItemByAccountId(userId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ?", userId).Update("state", 0).Error
	if err != nil {
		log.Errorf("FindBusinessOneInfo error: %v", err)
	}
	return
}

func (e *Entity) FindBusinessByAccountId(mId int64, sid int) (item []Entity, err error) {
	db := e.Db.Table(e.TableName())
	if mId != 0 {
		db.Where("account_id = ? ", mId)
	}
	if sid != 0 {
		db.Where("id = ? ", sid)
	}
	db.Find(&item)
	err = model.ModelError(db, global.MsgWarnModelNil)
	if err != nil {
		log.Errorf("FindBusinessByAccountId error: %v", err)
	}
	return
}
func (e *Entity) FirstServiceCoinChain(mId int64, sid int) (item Entity, err error) {
	db := e.Db.Table(e.TableName())
	if mId != 0 {
		db.Where("account_id = ? ", mId)
	}
	if sid != 0 {
		db.Where("id = ? ", sid)
	}
	db.First(&item)
	err = model.ModelError(db, global.MsgWarnModelNil)
	if err != nil {
		log.Errorf("FindBusinessByAccountId error: %v", err)
	}
	return
}

func (e *Entity) FirstServiceBySId(sid int) (item Entity, err error) {
	db := e.Db.Table(e.TableName())
	if sid != 0 {
		db.Where("id = ? ", sid)
	}
	db.First(&item)
	err = model.ModelError(db, global.MsgWarnModelNil)
	if err != nil {
		log.Errorf("FirstServiceBySId error: %v", err)
	}
	return
}
