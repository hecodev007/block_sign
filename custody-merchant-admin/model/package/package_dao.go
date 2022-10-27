package _package

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/module/log"
	"gorm.io/gorm"
)

func (e *Entity) InsertNewPackage() (err error) {
	err = e.Db.Table(e.TableName()).Create(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) DeletePackageItem(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ?", pId).Delete(e).Error
	if err != nil {
		log.Errorf("DelPackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdatePackageItem(pid int64, req Entity) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ?", pid).Updates(req).Error
	if err != nil {
		log.Errorf("UpdatePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) FindPackageItemById(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? ", pId).Find(&e).Error
	if err != nil {
		log.Errorf("FindPackageOneInfo error: %v", err)
	}
	if e.TypeName == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindPackageListByReq(req domain.PackageReqInfo) (list []Entity, total int64, err error) {
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

func (e *Entity) FindPackageByTypeModel(typeName, modelName string) (err error) {
	selectSql := e.Db.Table(e.TableName())
	selectSql = selectSql.Where("type_name = ? and model_name = ?", typeName, modelName)
	err = selectSql.Find(e).Error
	if err != nil {
		log.Errorf("FindPackageList error: %v", err)
	}
	return
}

func (e *PackagePay) FindAllPackagePayList() (list []PackagePay, total int64, err error) {
	selectSql := e.Db.Table("admin_package_pay")
	selectSql.Count(&total)
	err = selectSql.Find(&list).Error
	if err != nil {
		log.Errorf("FindPackageList error: %v", err)
	}
	return
}
func (e *PackageTrade) FindAllPackageTradeList() (list []PackageTrade, total int64, err error) {
	selectSql := e.Db.Table("admin_package_trade")
	selectSql.Count(&total)
	err = selectSql.Find(&list).Error
	if err != nil {
		log.Errorf("FindPackageList error: %v", err)
	}
	return
}

func (e *Entity) FindPackageScreen(req domain.PackageReqInfo) (typeArr []PackagePay, tradeArr []PackageTrade, modelMap map[string][]string, err error) {
	modelMap = make(map[string][]string)
	if req.Screen == "type" {
		name := &PackagePay{}
		err = e.Db.Table(name.PackagePayTableName()).Find(&typeArr).Error
		if err != nil {
			log.Errorf("FindPackageScreen type_name error: %v", err)
		}
	} else if req.Screen == "trade" {
		name := &PackageTrade{}
		err = e.Db.Table(name.PackageTradeTableName()).Find(&tradeArr).Error
		if err != nil {
			log.Errorf("FindPackageScreen type_name error: %v", err)
		}
	} else if req.Screen == "model" {
		var models []Entity
		err = e.Db.Table(e.TableName()).Distinct("type_name,model_name").Find(&models).Error
		if err != nil {
			log.Errorf("FindPackageScreen model_name error: %v", err)
		}
		for _, item := range models {
			arr := modelMap[item.TypeName]
			if len(arr) == 0 {
				arr = make([]string, 0)
			}
			arr = append(arr, item.ModelName)
			modelMap[item.TypeName] = arr
		}
	} else {
		name1 := &PackagePay{}
		err = e.Db.Table(name1.PackagePayTableName()).Find(&typeArr).Error
		if err != nil {
			log.Errorf("FindPackageScreen type_name error: %v", err)
		}

		name2 := &PackageTrade{}
		err = e.Db.Table(name2.PackageTradeTableName()).Find(&tradeArr).Error
		if err != nil {
			log.Errorf("FindPackageScreen type_name error: %v", err)
		}

		var models []Entity
		err = e.Db.Table(e.TableName()).Distinct("type_name,model_name").Find(&models).Error
		if err != nil {
			log.Errorf("FindPackageScreen model_name error: %v", err)
		}
		for _, item := range models {
			arr := modelMap[item.TypeName]
			if len(arr) == 0 {
				arr = make([]string, 0)
			}
			arr = append(arr, item.ModelName)
			modelMap[item.TypeName] = arr
		}
	}

	return
}
