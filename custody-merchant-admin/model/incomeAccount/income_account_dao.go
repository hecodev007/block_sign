package incomeAccount

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"strings"
	"time"
)

func (e *Entity) CreateIncome() error {
	e.CreatedAt = time.Now().Local()
	db := e.Db.Begin()
	db.Table(e.TableName()).Omit("updated_at", "deleted_at").Create(e)
	if db.Error != nil {
		db.Rollback()
		return db.Error
	}
	db.Commit()
	return nil
}

func (e *Entity) FindInfo(search *domain.SearchIncome) error {
	db := e.Db.Table(e.TableName())
	db.Where("merchant_id=? and coin_id = ? and combo_id = ? and service_id=?", search.MerchantId, search.CoinId, search.ComboId, search.ServiceId).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) UpdateInfo(id int64, mp map[string]interface{}) error {
	db := e.Db.Table(e.TableName())
	db.Where("id=?", id).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

func (e *Entity) UpdateInfoByVersion(id, version int64, mp map[string]interface{}) (int64, error) {
	db := e.Db.Table(e.TableName())
	res := db.Where("id=? and version=?", id, version).Updates(mp)
	return res.RowsAffected, model.ModelError(db, global.MsgWarnModelUpdate)
}

func (e *Entity) FindPage(search *domain.SearchIncome) (list []domain.IncomeInfo, count int64, err error) {

	db := model.DB().Model(e).
		Select("income_account.merchant_id," +
			"income_account.service_id," +
			"income_account.combo_type_name," +
			"income_account.combo_model_name," +
			"user_info.name as user_name," +
			"user_info.phone as user_phone," +
			"user_info.email as user_email," +
			"service.id as service_id," +
			"service.name as service_name," +
			"coin_info.name as coin_name," +
			"chain_info.name as chain_name," +
			"sum(income_account.withdraw_income + income_account.top_up_income + income_account.combo_income * ifnull(coin_info.price_usd / chain_info.price_usd,0)) as total_income," +
			"sum(income_account.combo_income) as combo_income," +
			"sum(income_account.withdraw_income) as withdraw_income," +
			"sum(income_account.top_up_income) as top_up_income," +
			"sum(income_account.top_up_nums) as top_up_nums," +
			"sum(income_account.top_up_price) as top_up_price," +
			"sum(income_account.top_up_fee) as top_up_fee," +
			"sum(income_account.top_up_destroy) as top_up_destroy," +
			"sum(income_account.withdraw_nums) as withdraw_nums," +
			"sum(income_account.withdraw_price) as withdraw_price," +
			"sum(income_account.withdraw_fee) as withdraw_fee," +
			"sum(income_account.withdraw_destroy) as withdraw_destroy," +
			"sum(income_account.miner_fee) as miner_fee").
		Joins("left join user_info on user_info.id = income_account.merchant_id").
		Joins("left join service on service.id = income_account.service_id").
		Joins("left join coin_info on coin_info.id = income_account.coin_id").
		Joins("left join chain_info on chain_info.id = coin_info.chain_id")

	if search.MerchantId != 0 {
		db.Where("income_account.merchant_id=?", search.MerchantId)
	}
	if search.Account != "" {
		if strings.Contains(search.Account, "@") {
			db.Where("user_info.email =? ", search.Account)
		} else {
			db.Where("user_info.phone =? ", search.Account)
		}
	}
	if search.ServiceId != 0 {
		db.Where("service.id =? ", search.ServiceId)
	}
	if search.CoinId != 0 {
		db.Where("coin_info.id =? ", search.CoinId)
	}
	if search.ChainId != 0 {
		db.Where("chain_info.id =? ", search.ChainId)
	}
	if search.StartTime != "" {
		db.Where("create_at >= ? ", search.StartTime)
	}
	if search.EndTime != "" {
		db.Where("create_at <= ? ", search.EndTime)
	}
	db.Group("income_account.merchant_id, income_account.service_id, income_account.coin_id,income_account.combo_type_name,income_account.combo_model_name")

	db.Offset(search.Offset).Limit(search.Limit).Find(&list).Offset(-1).Limit(-1).Count(&count).Debug()
	err = model.ModelError(db, global.MsgWarnModelNil)
	return list, count, err
}

func (e *Entity) FindChart(search *domain.SearchIncome) (list []domain.IncomeInfo, err error) {

	db := model.DB().Model(e).Debug().
		Select("income_account.coin_id,coin_info.name as coin_name," +
			"sum(income_account.withdraw_income+income_account.top_up_income + income_account.combo_income) * chain_info.price_usd  as total_income," +
			"sum(income_account.combo_income * coin_info.price_usd ) as combo_income," +
			"sum(income_account.withdraw_income * chain_info.price_usd) as withdraw_income," +
			"sum(income_account.top_up_income * chain_info.price_usd ) as top_up_income").
		Joins("left join user_info on user_info.id = income_account.merchant_id").
		Joins("left join service on service.id = income_account.service_id").
		Joins("left join coin_info on coin_info.id = income_account.coin_id").
		Joins("left join chain_info on chain_info.id = chain_id")
	if search.MerchantId != 0 {
		db.Where("income_account.merchant_id=?", search.MerchantId)
	}
	if search.Account != "" {
		if strings.Contains(search.Account, "@") {
			db.Where("user_info.email =? ", search.Account)
		} else {
			db.Where("user_info.phone =? ", search.Account)
		}
	}
	if search.ServiceId != 0 {
		db.Where("service.id =? ", search.ServiceId)
	}
	if search.CoinId != 0 {
		db.Where("coin_info.id =? ", search.CoinId)
	}
	if search.ChainId != 0 {
		db.Where("chain_info.id =? ", search.ChainId)
	}
	if search.StartTime != "" {
		db.Where("create_at >= ? ", search.StartTime)
	}
	if search.EndTime != "" {
		db.Where("create_at <= ? ", search.EndTime)
	}
	db.Group("income_account.coin_id").Order("total_income desc").Find(&list)
	err = model.ModelError(db, global.MsgWarnModelNil)
	return list, err
}
