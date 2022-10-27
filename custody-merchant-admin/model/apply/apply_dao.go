package apply

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/module/log"
	"gorm.io/gorm"
	"strings"
)

func (e *Entity) UpdateApplyItem() (err error) {
	err = e.Db.Table(e.TableName()).Updates(e).Error
	if err != nil {
		log.Errorf("UpdateApplyInfo error: %v", err)
	}
	return
}

func GetMapFromStruct(a interface{}) (m map[string]interface{}) {
	//aMap := make(map[string]interface{})
	//aByte, _ := json.Marshal(e)
	//json.Unmarshal(aByte, &aMap)
	return
}

func (e *Entity) UpdateApplyItemByMap(aMap map[string]interface{}) (err error) {
	//aMap := GetMapFromStruct(e)
	err = e.Db.Table(e.TableName()).Where("id = ?", e.Id).Updates(aMap).Error
	if err != nil {
		log.Errorf("UpdateApplyInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateApplyMap(id int64, mp map[string]interface{}) (err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ?", id).Updates(mp).Error
	if err != nil {
		log.Errorf("UpdateApplyInfo error: %v", err)
	}
	return
}

func (e *Entity) FindApplyItemById(pId int64) (item *Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("id = ?", pId).Find(&item).Error
	if err != nil {
		log.Errorf("FindApplyOneInfo error: %v", err)
	}
	if item.Phone == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindApplyItemByAccountId(pId int64) (item *Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ? ", pId).Find(&item).Error
	if err != nil {
		log.Errorf("FindApplyOneInfo error: %v", err)
	}
	if item.Phone == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindApplyListByReq(req domain.ApplyReqInfo) (list []ApplyDb, total int64, err error) {
	selectSql := e.Db.Table(e.TableName())
	selectSql = selectSql.Select("apply_pending.* ,u.is_test")
	selectSql = selectSql.Joins("left join user_info u on u.id = apply_pending.account_id")
	if req.AccountName != "" {
		selectSql = selectSql.Where("apply_pending.name = ?", req.AccountName)
	}
	if req.CardNum != "" {
		selectSql = selectSql.Where("apply_pending.id_card_num = ? or apply_pending.passport_num = ?", req.CardNum, req.CardNum)
	}

	if req.VerifyStatus != "" {
		if req.VerifyStatus == "had_verify" {
			selectSql = selectSql.Where("apply_pending.verify_status is not null")
		} else if req.VerifyStatus == "no_verify" {
			selectSql = selectSql.Where("apply_pending.verify_status is null")
		}
	}
	if req.VerifyResult != "" {
		if req.VerifyResult == "agree" {
			selectSql = selectSql.Where("apply_pending.verify_status = 'agree'")
		} else if req.VerifyResult == "refuse" {
			selectSql = selectSql.Where("apply_pending.verify_status = 'refuse'")
		}
	}
	if req.ContactStr != "" {
		if strings.Contains(req.ContactStr, "@") {
			selectSql = selectSql.Where("apply_pending.email = ?", req.ContactStr)
		} else {
			selectSql = selectSql.Where("apply_pending.phone = ?", req.ContactStr)
		}
	}
	selectSql.Count(&total)
	err = selectSql.Limit(req.Limit).Offset(req.Offset).Find(&list).Error
	if err != nil {
		log.Errorf("FindApplyList error: %v", err)
	}
	return
}
