package base

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

func FindAllChainCoins() ([]ChainInfo, error) {

	var chain []ChainInfo
	db := model.DB().Table("chain_info").Order("name asc").Find(&chain)
	return chain, model.ModelError(db, global.MsgWarnModelNil)
}

func InsertChainCoins(l []ChainInfo) error {
	err := model.DB().Table("chain_info").Create(l).Error
	return err
}

func UpdateChainCoins(l []ChainInfo) error {
	err := model.DB().Table("chain_info").Updates(l).Error
	return err
}

func DelChainCoins(l []ChainInfo) error {
	err := model.DB().Table("chain_info").Delete(l).Error
	return err
}

func FindChainsInName(name []string) ([]ChainInfo, error) {

	var chain []ChainInfo
	db := model.DB().Table("chain_info").Where("name in (?)  ", name).Find(&chain)
	if len(chain) > 0 {
		return chain, nil
	}
	return chain, model.ModelError(db, global.MsgWarnModelNil)
}
func FindChainsByName(name string) (ChainInfo, error) {

	var chain ChainInfo
	db := model.DB().Table("chain_info").Where("name =?  ", name).First(&chain)
	return chain, model.ModelError(db, global.MsgWarnModelNil)
}

func UpdateChainsById(id int, mp map[string]interface{}) error {
	db := model.DB().Table("chain_info").Where("id =?", id).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

func FindChainsById(id int) (ChainInfo, error) {

	var chain ChainInfo
	db := model.DB().Table("chain_info").Where("id =?  ", id).First(&chain)
	return chain, model.ModelError(db, global.MsgWarnModelNil)
}
