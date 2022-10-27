package record

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/module/log"
)

func (e *Entity) InsertNewPackage() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) FindPackageListByReq(req domain.RecordReqInfo) (list []Entity, total int64, err error) {
	selectSql := e.Db.Table(e.TableName())

	if req.BusinessId != 0 {
		selectSql = selectSql.Where("business_id = ?", req.BusinessId)
	}
	if req.Id != 0 {
		selectSql = selectSql.Where("business_id = ?", req.Id)
	}
	selectSql.Count(&total)
	err = selectSql.Order("id desc").Limit(req.Limit).Offset(req.Offset).Find(&list).Error
	if err != nil {
		log.Errorf("FindPackageList error: %v", err)
	}
	return
}

func (e *Entity) FindFinanceListByReq(req domain.MerchantReqInfo) (list []Entity, total int64, err error) {
	selectSql := e.Db.Table(e.TableName())
	selectSql = selectSql.Where("finance_id = ? and (operate = 'lock_asset' or operate = 'lock_user')", req.Id)
	selectSql.Count(&total)
	err = selectSql.Order("id desc").Limit(req.Limit).Offset(req.Offset).Find(&list).Error
	if err != nil {
		log.Errorf("FindPackageList error: %v", err)
	}
	return
}
