package base

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/module/log"
	"github.com/shopspring/decimal"
)

func FindCoins() ([]CoinInfo, error) {

	var auth []CoinInfo
	db := model.DB().Table("coin_info").Where("state = 0").Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func FindCoinsByIds(ids []string) ([]CoinInfo, error) {
	var auth []CoinInfo
	selectSql := model.DB().Table("coin_info").Where("state = 0")
	var chainArr []ChainInfo
	chainIds := make([]int, 0)
	if len(ids) > 0 {
		model.DB().Table("chain_info").Where("state = 0").Where("name in (?) ", ids).Find(&chainArr)
		for _, item := range chainArr {
			chainIds = append(chainIds, item.Id)
		}
		selectSql = selectSql.Where("chain_id in (?) ", chainIds)
	}

	db := selectSql.Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func FindAllCoins() ([]CoinInfo, error) {
	var auth []CoinInfo
	db := model.DB().Table("coin_info").Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func FindCoinsByCid(cid int) ([]CoinInfo, error) {

	var auth []CoinInfo
	db := model.DB().Table("coin_info").Where("chain_id=? and state = 0", cid).Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func FindCoinsById(id int) (*CoinInfo, error) {

	var coin = new(CoinInfo)
	db := model.DB().Table("coin_info").Where("id =?", id).First(coin)
	if coin != nil && coin.Id > 0 {
		return coin, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func GetChainByCId(id int, chainName string) (*ChainInfo, error) {

	var coin = new(ChainInfo)
	db := model.DB().Table("chain_info").
		Joins("left join coin_info on coin_info.chain_id = chain_info.id").
		Where("coin_info.id =? and chain_info.name = ?", id, chainName).
		First(coin)
	if coin != nil && coin.Id > 0 {
		return coin, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func FindCoinsByName(name string) (CoinInfo, error) {

	var coin CoinInfo
	db := model.DB().Table("coin_info").Where("name =? or full_name = ? ", name, name).First(&coin)
	if coin.Id > 0 {
		return coin, nil
	}
	return coin, model.ModelError(db, global.MsgWarnModelNil)
}

func FindCoinsByChainName(coinName, chainName string) (CoinInfo, error) {
	var coin CoinInfo
	db := model.DB().Table("coin_info").Joins("left join chain_info on coin_info.chain_id = chain_info.id").
		Where("(coin_info.name =? or coin_info.full_name = ?) and chain_info.name =?  ",
			coinName, coinName, chainName).
		First(&coin)
	if coin.Id > 0 {
		return coin, nil
	}
	return coin, model.ModelError(db, global.MsgWarnModelNil)
}

func FindCoinsInName(name []string) ([]CoinInfo, error) {

	var coin []CoinInfo
	db := model.DB().Table("coin_info").Where("name in (?)  ", name).Find(&coin)
	if len(coin) > 0 {
		return coin, nil
	}
	return coin, model.ModelError(db, global.MsgWarnModelNil)
}

func FindCoinsInIds(name []string) ([]CoinInfo, error) {

	var coin []CoinInfo
	db := model.DB().Table("coin_info").Where("id in (?)  ", name).Find(&coin)
	if len(coin) > 0 {
		return coin, nil
	}
	return coin, model.ModelError(db, global.MsgWarnModelNil)
}

func UpdateCoinsById(id int64, mp map[string]interface{}) error {
	db := model.DB().Table("coin_info").Where("id =?", id).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

func InsertSubCoins(l []CoinInfo) error {
	err := model.DB().Table("coin_info").Create(l).Error
	return err
}

func UpdateSubCoins(l []CoinInfo) error {
	db := orm.Cache(model.DB().Begin())
	for _, item := range l {
		m := make(map[string]interface{})
		m["state"] = item.State
		m["chain_id"] = item.ChainId
		m["name"] = item.Name
		if item.Confirm != 0 {
			m["confirm"] = item.Confirm
		}
		m["token"] = item.Token
		if item.PriceUsd.Cmp(decimal.NewFromInt(0)) == 1 {
			m["price_usd"] = item.PriceUsd
		}
		m["full_name"] = item.FullName
		err := db.Table("coin_info").Where("id = ? ", item.Id).Updates(m).Error
		if err != nil {
			db.Rollback()
			log.Errorf("更新coininfo err : %v", err)
			return err
		}
	}
	err := db.Commit().Error
	return err
}

func DelSubCoins(l []CoinInfo) error {
	err := model.DB().Table("coin_info").Delete(l).Error
	return err
}
