package finance

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/module/log"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

func (e *Entity) InsertNewFinance() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("InsertNewFinance error: %v", err)
	}
	return
}

//批量插入
func (e *Entity) InsertPatchFinance(arr []merchant.MerchantApply) (err error) {

	selectStr := "insert into service_finance (account_id,apply_id,created_at,updated_at) values "
	values := make([]string, 0)
	t := time.Now()
	tStr := t.Format(global.YyyyMmDdHhMmSs)
	for _, item := range arr {
		value := fmt.Sprintf("(%v,%v,'%v','%v')", item.AccountId, item.ApplyId, tStr, tStr)
		values = append(values, value)
	}
	selectStr = selectStr + strings.Join(values, ",")
	err = e.Db.Exec(selectStr).Error
	return
}

func (e *Entity) UpdateFinanceItem() (err error) {
	err = e.Db.Table(e.TableName()).Updates(e).Error
	if err != nil {
		log.Errorf("UpdateFinanceItem error: %v", err)
	}
	return
}

func (e *Entity) FindFinanceItemById(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? ", pId).First(e).Error
	if err != nil {
		log.Errorf("FindFinanceItemById error: %v", err)
	}
	if e.AccountId == 0 {
		err = gorm.ErrRecordNotFound
	}
	return
}

//OperateFinanceItemByFid 冻结解冻财务管理
func (e *Entity) OperateFinanceItemByFid(id int64, operate string) (err error) {
	updateMap := make(map[string]interface{})
	if operate == "lock_user" { //冻结用户和资产
		updateMap["is_lock"] = 1
		updateMap["is_lock_finance"] = 1
	} else if operate == "lock_asset" { //冻结资产
		updateMap["is_lock"] = 0
		updateMap["is_lock_finance"] = 1
	} else if operate == "unlock_user" { //解冻用户和资产
		updateMap["is_lock"] = 0
		updateMap["is_lock_finance"] = 0
	} else if operate == "unlock_asset" { //解冻资产
		updateMap["is_lock_finance"] = 0
	}
	err = e.Db.Table(e.TableName()).Where("id = ?", id).Updates(updateMap).Error
	if err != nil {
		log.Errorf("FindBusinessOneInfo error: %v", err)
	}
	return
}

//FindPushFinanceListByReq 获取所有已经推送财务审核
func (e *Entity) FindPushFinanceListByReq(req domain.MerchantReqInfo) (list []FinanceListDB, total int64, err error) {
	selectSql := e.Db.Table(e.TableName()).Select("u.name as account_name ,u.phone,u.email,u.is_test as account_status," +
		"service_finance.id as id,service_finance.account_id,service_finance.verify_status as fv_status," +
		"service_finance.remark as fv_remark,service_finance.is_lock,service_finance.is_lock_finance," +
		"service_finance.created_at," +
		"ap.id_card_num ,ap.passport_num,ap.real_name_status,ap.real_name_at,ap.test_end," +
		"ap.contract_start_at,ap.contract_end_at")
	//trade_type  交易类型
	selectSql = selectSql.Joins("left join user_info u on u.id = service_finance.account_id ")
	selectSql = selectSql.Joins("left join apply_pending ap on service_finance.account_id = ap.account_id ")
	if req.AccountName != "" {
		selectSql = selectSql.Where("u.name = ?", req.AccountName)
	}
	if req.CardNum != "" {
		selectSql = selectSql.Where("ap.id_card_num = ? or ap.passport_num = ?", req.CardNum, req.CardNum)
	}
	if req.AccountId != "" {
		selectSql = selectSql.Where("service_finance.account_id = ?", req.AccountId)
	}
	if req.RealNameStatus == "had_real" {
		selectSql = selectSql.Where("user_info.real_name_status = 1")
	} else if req.RealNameStatus == "no_real" {
		selectSql = selectSql.Where("user_info.real_name_status = 0")
	}
	//财务审核冻结，冻结正常异常筛选（normal/lock）
	if req.FvStatus != "" { //agree-通过，refuse-拒绝，wait-未处理
		if req.FvStatus == "agree" || req.FvStatus == "refuse" || req.FvStatus == "wait" {
			selectSql = selectSql.Where("service_finance.verify_status = ?", req.FvStatus)
		}
	}
	if req.LockStatus != "" { //agree-通过，refuse-拒绝，wait-未处理
		if req.LockStatus == "lock" {
			selectSql = selectSql.Where("service_finance.is_lock_finance = 1 ", req.FvStatus)
		} else if req.LockStatus == "unlock" {
			selectSql = selectSql.Where("service_finance.is_lock = 0", req.FvStatus)
		}
	}

	if req.ContactStr != "" {
		//
		if strings.Contains(req.ContactStr, "@") {
			selectSql = selectSql.Where("ap.email = ?", req.ContactStr)
		} else {
			selectSql = selectSql.Where("ap.phone = ?", req.ContactStr)
		}
	}
	selectSql.Count(&total)
	err = selectSql.Limit(req.Limit).Offset(req.Offset).Find(&list).Error
	if err != nil {
		log.Errorf("FindApplyList error: %v", err)
	}
	return
}
