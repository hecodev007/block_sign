package businessPackage

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"github.com/shopspring/decimal"

	//"custody-merchant/model"
	"gorm.io/gorm"
)

func (e *Entity) InsertNewBP() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) DeleteBPItem(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Delete(map[string]interface{}{"id": pId}).Error
	if err != nil {
		log.Errorf("DelPackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateBPItem() (err error) {
	err = e.Db.Table(e.TableName()).Updates(e).Error
	if err != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateBPItemByBusinessId(bId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ? ", bId).Updates(e).Error
	if err != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateBPItemByBusinessIdByMap(bId int64, uMap map[string]interface{}) (err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ? ", bId).Updates(uMap).Error
	if err != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateBPWhenChangePId(accountId, packageId int64, typeNum decimal.Decimal) (err error) {
	uMap := map[string]interface{}{
		"type_nums": typeNum,
		"had_used":  decimal.Zero,
	}
	err = e.Db.Table(e.TableName()).Where("account_id = ? and package_id != ?", accountId, packageId).Updates(uMap).Error
	if err != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateBPItemUsdBySIdByMap(bId int64, uMap map[string]interface{}) (int64, error) {
	db := e.Db.Table(e.TableName())
	err := db.Where("business_id = ?", bId).Updates(uMap)
	if db.Error != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return e.Db.RowsAffected, db.Error
}

func (e *Entity) AddBPItemTypeNumsBySIdByMap(bId int64, num decimal.Decimal) error {
	db := e.Db.Table(e.TableName())
	err := db.Where("id = ?", bId).Update("type_nums", gorm.Expr("type_nums + ?", num))
	if db.Error != nil {
		log.Errorf("AddBPItemTypeNumsBySIdByMap error: %v", err)
	}
	return db.Error
}

func (e *Entity) FindBPItemById(pId int64) (item *Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? ", pId).Find(item).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	if item.BusinessId == 0 {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindBPItemByAccountId(accountId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ? ", accountId).Order("id desc").First(e).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	if e.BusinessId == 0 {
		err = gorm.ErrRecordNotFound
	}
	return
}

//FindBPItemByBusinessId 业务线ID搜索套餐详情
func (e *Entity) FindBPItemByBusinessId(bId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ? ", bId).First(e).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	return
}

func (e *Entity) FindBPListByReq(req domain.PackageReqInfo) (list []Entity, total int64, err error) {
	selectSql := e.Db.Table(e.TableName())
	if req.TypeName != "" {
		selectSql = selectSql.Where("type_name = ?", req.TypeName)
	}
	if req.ModelName != "" {
		selectSql = selectSql.Where("model_name = ?", req.ModelName)
	}
	selectSql.Count(&total)
	err = selectSql.Limit(req.Limit).Offset(req.Offset).Find(&list).Error
	if err != nil {
		log.Errorf("FindPackageList error: %v", err)
	}
	return
}

//查询制定业务线，制定套餐
func (e *Entity) FindBPItemByPIdAndUserId(pId, userId int64) (item []Entity, err error) {
	sql := e.Db.Table(e.TableName()).Select("service_combo.id")
	sql = sql.Joins("left join service s on s.id = service_combo.business_id")
	sql = sql.Where("s.state != 2")
	err = sql.Where("service_combo.account_id = ? and service_combo.package_id = ? ", userId, pId).Order("service_combo.id asc").Find(&item).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}

	return
}

//查询制定业务线，制定套餐
func (e *Entity) FindBPItemByUserId(userId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ?", userId).Order("id desc").First(e).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	return
}

//查询制定业务线，制定套餐
func (e *Entity) FindBPItemBySId(sid int64) (item Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ?", sid).First(&item).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	return
}

func (e *Entity) FindMchBusinessByPackageId(pId, userId int64) (list []MchPackageDB, err error) {
	//套餐获益户
	selectSql := e.Db.Table(e.TableName()).Debug().Select("s.name,s.coin,s.sub_coin," +
		"service_combo.type_name as type_name,service_combo.model_name as model_name," +
		"service_combo.chain_nums as chain_nums,service_combo.coin_nums as coin_nums,service_combo.service_nums as service_nums," +
		"service_combo.deploy_fee as deploy_fee,service_combo.cover_fee as cover_fee,service_combo.chain_discount_nums as chain_discount_nums," +
		"service_combo.coin_discount_nums as coin_discount_nums,service_combo.year_discount_nums as year_discount_nums," +
		"service_combo.deposit_fee as deposit_fee,service_combo.addr_nums as addr_nums," +
		"service_combo.deduct_coin as deduct_coin")
	selectSql = selectSql.Joins("left join service s on s.id = service_combo.business_id")
	err = selectSql.Where("service_combo.package_id = ? and service_combo.account_id = ?", pId, userId).Find(&list).Error
	if err != nil {
		log.Errorf("FindBusinessOneInfo error: %v", err)
	}
	return
}

//SumBPItemByAccountId 业务线ID搜索套餐详情
func (e *Entity) SumBPItemByAccountId(aId int64) (*Entity, error) {
	et := &Entity{}
	err := model.DB().Table(e.TableName()).
		Select("sum(had_used) as had_used").
		Where("account_id = ? ", aId).
		Group("account_id").First(et).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	return et, err
}
