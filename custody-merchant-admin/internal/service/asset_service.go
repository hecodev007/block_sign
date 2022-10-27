package service

import (
	"custody-merchant-admin/model/assets"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/model/serviceChains"
	"fmt"
	"github.com/shopspring/decimal"
)

//钱包/财务 冻结/解冻/扣除
/*
bId 业务线id
amount 数额
coinName 币种
operateType 操作类型
remark 备注
*/

//LockAsset 冻结
func LockAsset(db *orm.CacheDB, bId int64, amount decimal.Decimal, coinName, operateType string, remark string) (err error) {
	return
}

//UnlockAsset 解冻
func UnlockAsset(db *orm.CacheDB, bId int64, amount decimal.Decimal, coinName, operateType string, remark string) (err error) {
	return
}

//DeductAsset 扣除（从冻结中扣除）
func DeductAsset(db *orm.CacheDB, bId int64, amount decimal.Decimal, coinName, operateType string, remark string) (err error) {
	return
}

func GetDateAssetsBySIdAndCId(serviceId int, coinId int) (*assets.Assets, error) {
	asDao := assets.NewEntity()
	coins, err := base.FindCoinsById(coinId)
	if err != nil {
		return asDao, err
	}
	if coins.Id == 0 {
		return asDao, fmt.Errorf("币种不存在")
	}
	bs := serviceChains.NewEntity()
	err = bs.FindServiceChainsInfo(serviceId, coins.Name)
	if err != nil {
		return asDao, err
	}
	return asDao.GetDateAssetsBySIdAndCId(serviceId, coinId)
}

func GetDateAssetsByPackageIdAndAccountId(serviceId int, coinId int) (*assets.Assets, error) {
	asDao := assets.NewEntity()
	coins, err := base.FindCoinsById(coinId)
	if err != nil {
		return asDao, err
	}
	if coins.Id == 0 {
		return asDao, fmt.Errorf("币种不存在")
	}
	bs := serviceChains.NewEntity()
	err = bs.FindServiceChainsInfo(serviceId, coins.Name)
	if err != nil {
		return asDao, err
	}
	return asDao.GetDateAssetsBySIdAndCId(serviceId, coinId)
}
