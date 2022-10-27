package financeAssets

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (e *Entity) InsertNewItem() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavefinanceAssetsInfo error: %v", err)
	}
	return
}

//
//func (e *Entity) FindSameAddress(coinId int64, address string) (err error) {
//	err = e.Db.Table(e.TableName()).Where("coin_id = ? and address = ?", coinId, address).Find(e).Error
//	if err != nil {
//		log.Errorf("SavefinanceAssetsInfo error: %v", err)
//	}
//	return
//}

func (e *Entity) FindSameBusinessId(coinId int64, bid int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("coin_id = ? and business_id = ?", coinId, bid).Find(e).Error
	if err != nil {
		log.Errorf("SavefinanceAssetsInfo error: %v", err)
	}
	return
}

func (a *Entity) UpdateFinanceAssetsAddNumsByBid(coinId int64, bid int64, amount decimal.Decimal) error {
	err := model.DB().Model(&Entity{}).Where("coin_id = ? and business_id = ?", coinId, bid).Update("nums", gorm.Expr("nums + ?", amount)).Error
	return err
}

func (a *Entity) UpdateFinanceAssetsSubNumsByBid(coinId int64, bid int64, amount decimal.Decimal) error {
	err := model.DB().Model(&Entity{}).Where("coin_id = ? and business_id = ?", coinId, bid).Update("nums", gorm.Expr("nums - ?", amount)).Error
	return err
}

//
//func (a *Entity) UpdateFinanceAssetsAddNums(coinId int64, address string, amount decimal.Decimal) error {
//	err := model.DB().Model(&Entity{}).Where("coin_id = ? and address = ?", coinId, address).Update("nums", gorm.Expr("nums + ?", amount)).Error
//	return err
//}
//
//func (a *Entity) UpdateFinanceAssetsSubNums(coinId int64, address string, amount decimal.Decimal) error {
//	err := model.DB().Model(&Entity{}).Where("coin_id = ? and address = ?", coinId, address).Update("nums", gorm.Expr("nums - ?", amount)).Error
//	return err
//}

func (e *Entity) FindFinanceAssetList(coinId int) (list []Entity, err error) {
	db := model.DB().Table(e.TableName())
	if coinId != 0 {
		db.Where("coin_id = ?", coinId)
	}
	db.Find(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}
