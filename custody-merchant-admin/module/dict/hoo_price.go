package dict

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/unitUsdt"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"reflect"
	"strings"
	"time"
)

type CoinPriceInfo struct {
	CoinPrice map[string]string
}

type HooFeeInfo struct {
	TokenName string `json:"token_name"`
	ChainName string `json:"chain_name"`
	Fee       string `json:"fee"`
	FeeUnit   string `json:"fee_unit"`
}

func SyncHooPrice() map[string]CoinPriceInfo {
	hooConf := Conf.Price["hoo"]
	res, err := xkutils.Get(hooConf.Url)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	resMap := map[string]interface{}{}
	json.Unmarshal(res, &resMap)
	data := resMap["data"].(map[string]interface{})
	ticker := data["ticker"].([]interface{})
	var ourMap = map[string]CoinPriceInfo{}
	for _, t := range ticker {
		priceMap := t.(map[string]interface{})
		name := priceMap["name"].(string)
		sm := map[string]string{}
		for key, value := range priceMap {
			sm[strings.ToUpper(key)] = value.(string)
		}
		ourMap[strings.ToUpper(name)] = CoinPriceInfo{
			CoinPrice: sm,
		}
	}
	cache.GetRedisClientConn().Set(global.CustodyPriceHoo, ourMap, time.Hour*24)
	InitCoinInfo()
	InitChainInfo()
	InitUniInfo()
	return ourMap
}

func SyncHooGeekPrice() map[string]CoinPriceInfo {
	hooGeekConf := Conf.Price["hoogeek"]
	res, err := xkutils.Get(hooGeekConf.Url)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	var ourMap = map[string]CoinPriceInfo{}
	cache.GetRedisClientConn().Get(global.CustodyPriceHoo, &ourMap)
	if ourMap == nil {
		ourMap = map[string]CoinPriceInfo{}
	}
	resMap := map[string]interface{}{}
	json.Unmarshal(res, &resMap)
	data := resMap["data"].(map[string]interface{})
	for key, d := range data {
		priceMap := d.(map[string]interface{})
		sm := map[string]string{}
		for k, value := range priceMap {
			retType := reflect.TypeOf(value)
			if retType.Name() == "string" {
				continue
			}
			sm[k] = decimal.NewFromFloat(value.(float64)).String()
		}
		if _, ok := ourMap[key]; !ok {
			ourMap[key] = CoinPriceInfo{
				CoinPrice: sm,
			}
		}
	}

	cache.GetRedisClientConn().Set(global.CustodyPriceHoo, ourMap, time.Hour*24)
	InitCoinInfo()
	InitUniInfo()
	return ourMap
}

func GetHooPriceByName(coinName, unit string) decimal.Decimal {
	coinName = strings.ToUpper(coinName)
	unit = strings.ToUpper(unit)
	var (
		coinPrice map[string]CoinPriceInfo
		price     string
	)
	cache.GetRedisClientConn().Get(global.CustodyPriceHoo, &coinPrice)
	if coinPrice == nil {
		coinPrice = SyncHooPrice()
	}
	if coinInfo, ok := coinPrice[coinName]; ok {
		price = coinInfo.CoinPrice[unit]
	}
	fromString, err := decimal.NewFromString(price)
	if err != nil {
		return decimal.Decimal{}
	}
	return fromString
}

func InitCoinInfo() {
	clist, err := base.FindCoins()
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, info := range clist {
		price := GetHooPriceByName(info.Name, "usd")
		go func(ifs base.CoinInfo) {
			if !price.IsZero() {
				err = base.UpdateCoinsById(ifs.Id, map[string]interface{}{"price_usd": price})
				if err != nil {
					log.Error(err.Error())
					return
				}
			}
		}(info)
		time.Sleep(20)
	}
}

func InitChainInfo() {
	clist, err := base.FindAllChainCoins()
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, info := range clist {

		price := GetHooPriceByName(info.Name, "usd")

		go func(ifs base.ChainInfo) {
			if !price.IsZero() {
				err = base.UpdateChainsById(ifs.Id, map[string]interface{}{"price_usd": price})
				if err != nil {
					return
				}
			}
		}(info)
	}
}

func InitUniInfo() {
	dao := new(unitUsdt.UnitUsdt)
	ulist, err := dao.GetUnitUsdtList()
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, info := range ulist {
		ratio := GetHooPriceByName("usdt", info.Name)
		go func(ifs unitUsdt.UnitUsdt) {
			err = dao.UpdateUnitById(ifs.Id, map[string]interface{}{"ratio": ratio})
			if err != nil {
				return
			}
		}(info)
	}
}

func InitHooFee() map[string]HooFeeInfo {
	hlst := map[string]HooFeeInfo{}
	res, err := xkutils.Get(Conf.Fee.Url)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	resMap := map[string]interface{}{}
	err = json.Unmarshal(res, &resMap)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	data := resMap["data"].([]interface{})
	for _, d := range data {
		dt := d.(map[string]interface{})
		hlst[dt["token_name"].(string)] = HooFeeInfo{
			TokenName: dt["token_name"].(string),
			ChainName: dt["chain_name"].(string),
			Fee:       dt["fee"].(string),
			FeeUnit:   dt["fee_unit"].(string),
		}
	}
	//fmt.Printf("%v", hlst)
	return hlst
}
