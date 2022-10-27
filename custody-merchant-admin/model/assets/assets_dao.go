package assets

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (e *Assets) InsertNewAssets() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SaveBusinessInfo error: %v", err)
	}
	return
}

func (a *Assets) CreateAssets(ass Assets) (int64, error) {
	// 通过数据的指针来创建
	result := model.DB().Begin()
	result.Create(&ass)
	err := model.ModelError(result, global.MsgWarnModelAdd)
	if err != nil {
		result.Rollback()
		return 0, err
	} else {
		result.Commit()
	}
	return result.RowsAffected, nil
}

func (a *Assets) UpDateAssetsBySCId(sid, cid int, version int64, mp map[string]interface{}) (int64, error) {
	db := model.DB().Model(&Assets{}).Where("service_id = ? and coin_id = ? and version =? ", sid, cid, version).Updates(mp)
	return db.RowsAffected, model.ModelError(db, global.MsgWarnModelUpdate)
}

func (a *Assets) UpDateAssetsSubNumsBySCId(sid, cid int, amount decimal.Decimal) error {
	uMap := map[string]interface{}{
		"nums":           gorm.Expr("nums - ?", amount),
		"finance_freeze": gorm.Expr("finance_freeze + ?", amount),
	}
	db := model.DB().Model(&Assets{}).Where("service_id = ? and coin_id = ? "+
		"and nums >= ?", sid, cid, amount).Updates(uMap)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

func (a *Assets) GetDateAssetsBySIdAndCId(sid, cid int) (*Assets, error) {
	ass := Assets{}
	db := model.DB().Table("assets").Debug().Where("service_id = ? and coin_id = ?", sid, cid).First(&ass)
	return &ass, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) GetAssetsNumsBySIdAndCId(sid, cid int) (*Assets, error) {
	ass := new(Assets)
	db := model.DB().Model(&Assets{}).Where("service_id = ? and coin_id = ?", sid, cid).First(ass)
	return ass, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) GetDateAssetsByCName(cname string) (*Assets, error) {
	ass := new(Assets)
	db := model.DB().Model(&Assets{}).Where("coin_name = ?", cname).First(ass)
	return ass, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) GetAssets(id int64) ([]Assets, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []Assets
	)

	build.AddString("select sum(nums) as nums, sum(freeze) as freeze, service_id, coin_id from assets").
		AddString(" where (select count(1) from coin_info where coin_info.id = assets.coin_id limit 1) > 0 and (select count(1) from user_service where user_service.sid = assets.service_id ")
	if id != 0 {
		build.StringBuild(" and user_service.uid = %d ", id)
	}
	build.AddString(" limit 1) > 0 ")
	build.AddString(" group by service_id, coin_id ")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) FindAssetsList(as *domain.AssetsSelect, id int64) ([]AssetsList, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AssetsList
		limit  int
	)
	if model.FilteredSQLInject(as.CoinName) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	limit = as.Limit
	if as.Limit == 0 {
		limit = 10
	}
	build.AddString(" select assets.coin_id, sum(assets.freeze) as freeze,sum(assets.nums) as nums, ").
		AddString("  coin_info.name as coin_name,chain_info.name as chain_name, sum(assets.nums * coin_info.price_usd) as valuation,").
		AddString(" chain_info.id as chain_id from assets ").
		AddString(" left join coin_info on coin_info.id = assets.coin_id ").
		AddString(" left join chain_info on coin_info.chain_id = chain_info.id ").
		AddString(" left join service_chains on service_chains.service_id = assets.service_id ").
		AddString(" left join user_info on user_info.id = service_chains.merchant_id ").
		AddString(" where  assets.coin_id = service_chains.coin_id ").
		AddString(" and (select count(1) from service where service.id = assets.service_id and service.state !=2 limit 1) > 0")
	if id != 0 {
		build.StringBuild(" and service_chains.merchant_id = %d ", id)
	}
	if as.CoinId != 0 {
		build.StringBuild(" and assets.coin_id = %d", as.CoinId)
	}
	// 正式还是非正式账号
	if as.IsTest != -1 {
		build.StringBuild(" and user_info.is_test = %d", as.IsTest)
	}
	if as.ServiceId > 0 {
		build.StringBuild(" and assets.service_id = %d ", as.ServiceId)
	}
	if as.Show > 0 && as.CoinState == -1 {
		build.AddString(" and (assets.nums * coin_info.price_usd > 1 or assets.freeze * coin_info.price_usd > 1)")
	}
	if as.Show > 0 && as.CoinState == 0 {
		build.AddString(" and assets.nums * coin_info.price_usd > 1 ")
	}
	if as.Show > 0 && as.CoinState == 1 {
		build.AddString(" and assets.freeze * coin_info.price_usd > 1  ")
	}
	build.StringBuild(" group by assets.coin_id order by valuation desc")
	build.StringBuild(" limit %d,%d", as.Offset, limit)
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) CountAssetsList(as *domain.AssetsSelect, id int64) (int64, error) {
	var (
		build = new(xkutils.StringBuilder)
		count int64
	)
	if model.FilteredSQLInject(as.CoinName) {
		return 0, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString(" select count(1) as count from assets ").
		AddString(" left join coin_info on coin_info.id = assets.coin_id ").
		AddString(" left join chain_info on coin_info.chain_id = chain_info.id ").
		AddString(" left join service_chains on service_chains.service_id = assets.service_id ").
		AddString(" left join user_info on user_info.id = service_chains.merchant_id ").
		AddString(" where assets.coin_id = service_chains.coin_id ").
		AddString(" and (select count(1) from service where service.id = assets.service_id and service.state !=2 limit 1) > 0")
	if id != 0 {
		build.StringBuild(" and service_chains.merchant_id = %d ", id)
	}
	if as.CoinId != 0 {
		build.StringBuild(" and assets.coin_id = %d", as.CoinId)
	}
	// 正式还是非正式账号
	if as.IsTest != -1 {
		build.StringBuild(" and user_info.is_test = %d", as.IsTest)
	}
	if as.ServiceId > 0 {
		build.StringBuild(" and assets.service_id = %d ", as.ServiceId)
	}
	if as.Show > 0 && as.CoinState == -1 {
		build.AddString(" and (assets.nums * coin_info.price_usd > 1 or assets.freeze * coin_info.price_usd > 1)")
	}
	if as.Show > 0 && as.CoinState == 0 {
		build.AddString(" and assets.nums * coin_info.price_usd > 1 ")
	}
	if as.Show > 0 && as.CoinState == 1 {
		build.AddString(" and assets.freeze * coin_info.price_usd > 1  ")
	}
	build.StringBuild(" group by assets.coin_id ")
	db := model.DB().Raw(build.ToString()).Count(&count)
	return count, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) FindAssetsListGroup(id int64) ([]Assets, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []Assets
	)
	build.AddString(" select sum(assets.nums) as nums,sum(assets.freeze) as freeze,sum((assets.nums + assets.freeze) * coin_info.price_usd) as price, coin_info.name as coin_name ").
		AddString(" from assets left join coin_info on coin_info.id = assets.coin_id ").
		AddString(" where  (select count(1) from service_audit_role where service_audit_role.sid = assets.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d ", id)
	}
	build.AddString(" limit 1 ) > 0  group by coin_info.id order by price desc ")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) FindServiceAssetsList(as *domain.AssetsSelect, id int64) ([]AssetsList, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AssetsList
		limit  int
	)
	if model.FilteredSQLInject(as.CoinName) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	limit = as.Limit
	if as.Limit == 0 {
		limit = 10
	}
	build.AddString(" select assets.coin_id, sum(assets.freeze) as freeze,sum(assets.nums) as nums, ").
		AddString("  coin_info.name as coin_name,chain_info.name as chain_name, sum(assets.nums * coin_info.price_usd) as valuation,").
		AddString(" chain_info.id as chain_id from assets ").
		AddString(" left join coin_info on coin_info.id = assets.coin_id ").
		AddString(" left join chain_info on coin_info.chain_id = chain_info.id ").
		AddString(" left join service_chains on service_chains.service_id = assets.service_id ").
		AddString(" left join user_info on user_info.id = service_chains.merchant_id ").
		AddString(" where  assets.coin_id = service_chains.coin_id ").
		AddString(" and (select count(1) from service where service.id = assets.service_id and service.state !=2 limit 1) > 0")
	if id != 0 {
		build.StringBuild(" and service_chains.merchant_id = %d ", id)
	}
	if as.CoinId != 0 {
		build.StringBuild(" and assets.coin_id = %d", as.CoinId)
	}
	if as.ServiceId > 0 {
		build.StringBuild(" and assets.service_id = %d ", as.ServiceId)
	}
	build.StringBuild(" group by assets.coin_id order by valuation desc")
	build.StringBuild(" limit %d,%d", as.Offset, limit)
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) CountServiceAssetsList(as *domain.AssetsSelect, id int64) (int64, error) {
	var (
		build = new(xkutils.StringBuilder)
		count int64
	)
	if model.FilteredSQLInject(as.CoinName) {
		return 0, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString(" select count(1) as count from assets ").
		AddString(" left join coin_info on coin_info.id = assets.coin_id ").
		AddString(" left join chain_info on coin_info.chain_id = chain_info.id ").
		AddString(" left join service_chains on service_chains.service_id = assets.service_id ").
		AddString(" left join user_info on user_info.id = service_chains.merchant_id ").
		AddString(" where assets.coin_id = service_chains.coin_id ").
		AddString(" and (select count(1) from service where service.id = assets.service_id and service.state !=2 limit 1) > 0")
	if id != 0 {
		build.StringBuild(" and service_chains.merchant_id = %d ", id)
	}
	if as.CoinId != 0 {
		build.StringBuild(" and assets.coin_id = %d", as.CoinId)
	}
	if as.ServiceId > 0 {
		build.StringBuild(" and assets.service_id = %d ", as.ServiceId)
	}
	build.StringBuild(" group by assets.coin_id")
	db := model.DB().Raw(build.ToString()).Count(&count)
	return count, model.ModelError(db, global.MsgWarnModelNil)
}

func (a *Assets) FindFinanceAssetsList(as *domain.AssetsSelect) ([]FinanceServiceAsset, int64, error) {
	financeModel := []FinanceServiceAsset{}
	count := int64(0)
	db := model.DB().Table("assets").
		Select("s.name as service_name,s.account_id as account_id,ui.name as user_name,assets.coin_name, assets.finance_freeze").
		Joins("left join service s on s.id = assets.service_id").
		Joins("left join user_info ui on s.account_id = ui.id").
		Joins("left join coin_info ci on ci.id = assets.coin_id")
	if as.CoinId != 0 {
		db.Where("coin_info.id = ?", as.CoinId)
	}
	if as.CoinName != "" {
		db.Where("assets.coin_name = ?", as.CoinName)
	}
	db.Count(&count).Limit(as.Limit).Offset(as.Offset).Find(&financeModel)
	return financeModel, count, model.ModelError(db, global.MsgWarnModelNil)
}
