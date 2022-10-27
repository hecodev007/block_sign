package serviceSecurity

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
)

func (e *Entity) InsertNewItem() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateItemByBusinessId(bId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ? ", bId).Updates(e).Error
	if err != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateItemByBusinessIdByMap(bId int64, umap map[string]interface{}) (err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ? ", bId).Updates(umap).Error
	if err != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdatesByBId(bId int64, mp map[string]interface{}) (err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ? ", bId).Updates(mp).Error
	if err != nil {
		log.Errorf("UpdatesByBId error: %v", err)
	}
	return
}

//FindItemByBusinessId 业务线ID搜索套餐详情
func (e *Entity) FindItemByBusinessId(bId int64) (err error) {
	selectSql := e.Db.Table(e.TableName())
	//selectSql = selectSql.Joins("left join service s on s.id = service_security.business_id")
	err = selectSql.Where("business_id = ? ", bId).First(&e).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	return
}

//FindItemByBusinessId 业务线ID搜索套餐详情
func (e *Entity) FindBindInfoByBusinessId(bId int64) (item BusinessSecurityDB, err error) {
	selectSql := e.Db.Table(e.TableName()).Select("service_security.*,s.phone,s.email")
	selectSql = selectSql.Joins("left join service s on s.id = service_security.business_id")
	err = selectSql.Where("business_id = ? ", bId).First(&item).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	return
}

func (e *Entity) GetBindInfoByClientId(clientId string) error {
	db := e.Db.Table(e.TableName()).
		Where("client_id =? ", clientId).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetBindInfoByBid(bid int) error {
	db := e.Db.Table(e.TableName()).
		Where("business_id =? ", bid).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}
