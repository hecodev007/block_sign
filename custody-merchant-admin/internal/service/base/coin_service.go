package base

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/base"
	"encoding/json"
)

//SearchChainsList 主链币列表
func SearchChainsList() (i interface{}, err error) {
	var chains []base.ChainInfo
	chains, err = base.FindAllChainCoins()
	if err != nil {
		return i, global.DaoError(err)
	}
	backArr := make([]domain.ChainInfo, 0)

	b, err := json.Marshal(chains)
	if err != nil {
		return i, global.OperationErrorText(err.Error())
	}

	err = json.Unmarshal(b, &backArr)
	if err != nil {
		return i, global.OperationErrorText(err.Error())
	}

	return backArr, err
}

//SearchSubCoinList 代币列表
func SearchSubCoinList() (i interface{}, err error) {
	var coins []base.CoinInfo
	coins, err = base.FindCoins()
	if err != nil {
		return i, global.DaoError(err)
	}
	backArr := make([]domain.CoinInfo, 0)

	b, err := json.Marshal(coins)
	if err != nil {
		return i, global.OperationErrorText(err.Error())
	}

	err = json.Unmarshal(b, &backArr)
	if err != nil {
		return i, global.OperationErrorText(err.Error())
	}

	return backArr, err
}

//SearchSubCoinList 代币列表
func SearchSubCoinListByIds(ids []string) (i interface{}, err error) {
	var coins []base.CoinInfo
	coins, err = base.FindCoinsByIds(ids)
	if err != nil {
		return i, global.DaoError(err)
	}
	backArr := make([]domain.CoinInfo, 0)

	b, err := json.Marshal(coins)
	if err != nil {
		return i, global.OperationErrorText(err.Error())
	}

	err = json.Unmarshal(b, &backArr)
	if err != nil {
		return i, global.OperationErrorText(err.Error())
	}

	return backArr, err
}

//GetCoinById 代币
func GetCoinById(id int) (*base.CoinInfo, error) {
	byId, err := base.FindCoinsById(id)
	if err != nil {
		return nil, err
	}
	return byId, nil
}

//GetChainByCId 代币获取主链币
func GetChainByCId(id int, chainName string) (*base.ChainInfo, error) {
	byId, err := base.GetChainByCId(id, chainName)
	if err != nil {
		return nil, err
	}
	return byId, nil
}
