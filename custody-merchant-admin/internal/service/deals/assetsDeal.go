package deals

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/assets"
	"custody-merchant-admin/model/assets/assetsLog"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/financeAssets"
	"custody-merchant-admin/model/unitUsdt"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

// FindServiceAssetsList
// 商户查询业务线资产
func FindServiceAssetsList(as *domain.AssetsSelect, id int64) ([]map[string]interface{}, int64, error) {

	var (
		unit   = new(unitUsdt.UnitUsdt)
		ats    = new(assets.Assets)
		asList = make([]map[string]interface{}, 0)
	)
	list, err := ats.FindServiceAssetsList(as, id)
	if err != nil {
		return asList, 0, err
	}
	count, err := ats.CountServiceAssetsList(as, id)
	if err != nil {
		return asList, 0, err
	}
	rname := "cny"
	usdtInfo, err := unit.GetUnitUsdtById(as.UnitId)
	if err != nil {
		return asList, 0, err
	}
	if usdtInfo != nil {
		rname = usdtInfo.Name
	}
	ratio := dict.GetHooPriceByName("USDT", rname)
	for i := 0; i < len(list); i++ {
		coins, err := base.FindCoinsByName(list[i].CoinName)
		if err != nil {
			return asList, 0, err
		}
		coinPrice := coins.PriceUsd
		nums := list[i].Nums
		valuation := nums.Mul(coinPrice)
		asList = append(asList, map[string]interface{}{
			"chainName": list[i].ChainName,
			"coinName":  list[i].CoinName,
			"nums":      nums,
			"coinPrice": coinPrice.Round(6),
			"valuation": valuation.Round(6),
			"reduced":   valuation.Mul(ratio).Round(6),
		})
	}
	return asList, count, nil
}

// FindAssetsList
// 查询资产
func FindAssetsList(as *domain.AssetsSelect, id int64) ([]domain.AssetsInfo, int64, error) {

	var (
		unit   = new(unitUsdt.UnitUsdt)
		ats    = new(assets.Assets)
		asList = []domain.AssetsInfo{}
	)
	list, err := ats.FindAssetsList(as, id)
	if err != nil {
		return asList, 0, err
	}
	count, err := ats.CountAssetsList(as, id)
	if err != nil {
		return asList, 0, err
	}

	rname := "cny"
	usdtInfo, err := unit.GetUnitUsdtById(as.UnitId)
	if err != nil {
		return asList, 0, err
	}
	if usdtInfo != nil {
		rname = usdtInfo.Name
	}
	ratio := dict.GetHooPriceByName("USDT", rname)

	for i := 0; i < len(list); i++ {
		coins, err := base.FindCoinsByName(list[i].CoinName)
		if err != nil {
			return asList, 0, err
		}
		coinPrice := coins.PriceUsd
		nums := list[i].Nums
		if as.CoinState == -1 {
			nums = list[i].Nums.Add(list[i].Freeze)
		}
		if as.CoinState == 1 {
			nums = list[i].Freeze
		}
		valuation := nums.Mul(coinPrice)
		asList = append(asList, domain.AssetsInfo{
			ServiceId: list[i].ServiceId,
			CoinId:    list[i].CoinId,
			ChainName: list[i].ChainName,
			CoinName:  list[i].CoinName,
			Nums:      nums,
			CoinPrice: coinPrice.Round(6),
			Valuation: valuation.Round(6),
			Freeze:    list[i].Freeze.Round(6),
			Reduced:   valuation.Mul(ratio).Round(6),
		})
	}

	return asList, count, nil
}

func FindAssetsRing(id int64) (domain.RingObj, error) {
	var (
		//asList []domain.AssetsRingInfo
		ats        = new(assets.Assets)
		totalPrice = decimal.Zero
		mapList    = map[string]domain.AssetsRingInfo{}
		ringObj    = domain.RingObj{}
		ringItems  = []domain.RingItems{}
		ringInfo   = []domain.RingInfo{}
		SortList   = []string{}
	)
	ringObj.RingInfo = ringInfo
	ringObj.RingItems = ringItems

	list, err := ats.FindAssetsListGroup(id)
	if err != nil {
		return ringObj, err
	}
	for i := 0; i < len(list); i++ {
		coinPrice := dict.GetHooPriceByName(list[i].CoinName, "usd")
		cnyPrice := dict.GetHooPriceByName(list[i].CoinName, "cny")
		haveNums := list[i].Nums.Add(list[i].Freeze)
		if i < 10 {
			totalPrice = totalPrice.Add(haveNums.Mul(coinPrice))
		}
		SortList = append(SortList, list[i].CoinName)
		mapList[list[i].CoinName] = domain.AssetsRingInfo{
			Nums:      haveNums,                         // 币数量
			Price:     haveNums.Mul(coinPrice).Round(6), // 默认usdt
			Valuation: haveNums.Mul(cnyPrice).Round(6),  // 折合人名币
		}
	}

	i := 0
	for _, sl := range SortList {
		if i < 5 {
			ringInfo = append(ringInfo, domain.RingInfo{
				Name:      sl,
				Value:     mapList[sl].Price.Round(6),
				ItemStyle: domain.ItemStyle{Color: dict.Colors[i]},
			})
			v := decimal.Zero
			if !totalPrice.IsZero() && !mapList[sl].Price.IsZero() {
				v = mapList[sl].Price.Div(totalPrice)
			}
			ringItems = append(ringItems, domain.RingItems{
				Name:  sl,
				Value: v.Round(6),
				Color: dict.Colors[i],
			})
		}
		i++
		//asList = append(asList, domain.AssetsRingInfo{
		//	CoinName:  key,
		//	Nums:      mapList[key].Nums,      // 币数量
		//	Price:     mapList[key].Price,     // 默认usdt
		//	Valuation: mapList[key].Valuation, // 折合人名币
		//})
	}
	ringObj.RingItems = ringItems
	ringObj.RingInfo = ringInfo
	return ringObj, nil
}

// FindAssetsByHour
// 查询24小时资产
func FindAssetsByHour(selectTime, nowTime string, id int64) ([]domain.AssetsTimeInfo, error) {
	dao := new(assetsLog.AssetsHours)
	day, err := dao.FindAssetsHours(selectTime, nowTime, id)
	if err != nil {
		return nil, err
	}
	return DealAssetsLine(day)
}

// FindAssetsByDay
// 查询每天资产
func FindAssetsByDay(startTime, endTime string, id int64) ([]domain.AssetsTimeInfo, error) {
	var res []domain.AssetsTimeInfo
	dao := new(assetsLog.AssetsDay)
	day, err := dao.FindAssetsDay(startTime, endTime, id)
	if err != nil {
		return nil, err
	}
	startDay, err := dao.FindAssetsDayByStart(startTime, id)
	if err != nil {
		return nil, err
	}
	data, err := TwoAssetsData(startDay, day)
	if err != nil {
		return nil, err
	}
	ln := len(data)
	zero, err := AddDataZero(ln, 1, startTime, endTime, global.YyyyMmDd)
	if err != nil {
		return nil, err
	}
	data = append(data, zero...)
	for i := 0; i < len(data); i++ {
		dates := strings.Split(data[i].CreateTime, "-")
		res = append(res, domain.AssetsTimeInfo{
			Scale:      data[i].Scale,
			Price:      data[i].Price,
			CreateTime: dates[1] + "-" + dates[2],
			Freeze:     data[i].Freeze,
		})
	}
	return res, err
}

// FindAssetsByWeek
// 查询每周资产
func FindAssetsByWeek(selectTime, endTime string, id int64) ([]domain.AssetsTimeInfo, error) {
	dao := new(assetsLog.AssetsDay)
	day, err := dao.FindAssetsWeek(selectTime, endTime, id)
	if err != nil {
		return nil, err
	}
	startDay, err := dao.FindAssetsWeekStart(selectTime, id)
	if err != nil {
		return nil, err
	}
	data, err := TwoAssetsData(startDay, day)
	if err != nil {
		return nil, err
	}
	ln := len(data)
	zero, err := AddDataZero(ln, 7, selectTime, endTime, global.YyyyMmDd)
	if err != nil {
		return nil, err
	}
	data = append(data, zero...)
	return data, err
}

// FindAssetsByMonth
// 查询每月资产
func FindAssetsByMonth(selectTime, endTime string, id int64) ([]domain.AssetsTimeInfo, error) {
	dao := new(assetsLog.AssetsMonth)
	month, err := dao.FindAssetsMonth(selectTime, endTime, id)
	if err != nil {
		return nil, err
	}
	start, err := dao.FindAssetsMonthStart(selectTime, id)
	if err != nil {
		return nil, err
	}
	data, err := TwoAssetsData(start, month)
	if err != nil {
		return nil, err
	}
	ln := len(data)
	zero, err := AddDataZero(ln, 30, selectTime, endTime, global.YyyyMmDd)
	if err != nil {
		return nil, err
	}
	data = append(data, zero...)
	return data, err
}

// FindAssetsByYears
// 查询每年资产
func FindAssetsByYears(selectTime, endTime string, id int64) ([]domain.AssetsTimeInfo, error) {
	dao := new(assetsLog.AssetsMonth)
	year, err := dao.FindAssetsYear(selectTime, endTime, id)
	if err != nil {
		return nil, err
	}
	start, err := dao.FindAssetsYearStart(selectTime, id)
	if err != nil {
		return nil, err
	}
	data, err := TwoAssetsData(start, year)
	if err != nil {
		return nil, err
	}
	ln := len(data)
	zero, err := AddDataZero(ln, 365, selectTime, endTime, global.YyyyMmDd)
	if err != nil {
		return nil, err
	}
	data = append(data, zero...)
	return data, err
}

// GetScale
// first 前一天
// second 后一天
func GetScale(first, second decimal.Decimal) string {
	// 资产 == 0
	if !first.IsZero() {
		if !second.IsZero() {
			// 波动 = (第二天-前一天)/前一天
			bf := decimal.NewFromInt(100)
			sub := second.Sub(first)
			if sub.IsZero() {
				return "0"
			}
			return (sub.Div(first)).Mul(bf).Round(2).String()
		} else {
			return "-100"
		}
	} else {
		return "100"
	}
}

func DealAssetsLine(astime []assetsLog.AsTime) ([]domain.AssetsTimeInfo, error) {

	var (
		dayList = []domain.AssetsTimeInfo{}
		mapList = map[string]domain.AssetsTimeInfo{}
		keyList = []string{}
	)
	for i := 0; i < len(astime); i++ {
		coin, err := base.FindCoinsById(astime[i].CoinId)
		if err != nil {
			return dayList, err
		}
		coinPrice := decimal.Zero
		if coin != nil {
			coinPrice = coin.PriceUsd
		}
		price := astime[i].Nums.Mul(coinPrice)        // 未冻结资产 = 未冻结数*单位
		freezeNums := astime[i].Freeze.Mul(coinPrice) // 冻结资产 = 冻结数*单位
		// 根据时间进行分组
		if _, ok := mapList[astime[i].CreateTime]; ok {
			// map 已经存在该时间
			dl := mapList[astime[i].CreateTime]
			mapList[astime[i].CreateTime] = domain.AssetsTimeInfo{
				CreateTime: astime[i].CreateTime,
				Price:      dl.Price.Add(price).Add(freezeNums), // 资产价值 += 未冻结资产+冻结资产
				Freeze:     dl.Freeze.Add(freezeNums),           // 冻结资产+
			}
		} else {
			keyList = append(keyList, astime[i].CreateTime)
			// map 未存在该时间
			mapList[astime[i].CreateTime] = domain.AssetsTimeInfo{
				CreateTime: astime[i].CreateTime,
				Price:      price.Add(freezeNums),
				Freeze:     freezeNums,
			}
		}
	}
	dayList, err := AssetsTimeInfoResult(keyList, mapList)
	if err != nil {
		return []domain.AssetsTimeInfo{}, err
	}
	return dayList, err
}

func TwoAssetsData(startDay, day []assetsLog.AsTime) ([]domain.AssetsTimeInfo, error) {

	one, err := DealAssetsLine(startDay)
	if err != nil {
		return nil, err
	}
	two, err := DealAssetsLine(day)
	if err != nil {
		return nil, err
	}
	if one != nil && two != nil {
		two[0] = domain.AssetsTimeInfo{
			Scale:      GetScale(one[0].Price, two[0].Price),
			Price:      two[0].Price,
			CreateTime: two[0].CreateTime,
			Freeze:     two[0].Freeze,
		}
	}
	return two, err
}

func AssetsTimeInfoResult(keyList []string, mapList map[string]domain.AssetsTimeInfo) ([]domain.AssetsTimeInfo, error) {
	var (
		dayList []domain.AssetsTimeInfo
	)
	blo := decimal.Zero

	for i, _ := range keyList {
		// 遍历map为数组
		dayList = append(dayList, domain.AssetsTimeInfo{
			Scale:      GetScale(blo, mapList[keyList[i]].Price),
			Price:      mapList[keyList[i]].Price,
			CreateTime: mapList[keyList[i]].CreateTime,
			Freeze:     mapList[keyList[i]].Freeze,
		})
		blo = mapList[keyList[i]].Price
	}

	return dayList, nil
}

func AddDataZero(ln, units int, startTime, endTime, layout string) ([]domain.AssetsTimeInfo, error) {
	data := []domain.AssetsTimeInfo{}
	if startTime == "" || ln == 0 {
		creatTime := time.Now().Add(-time.Hour * 24 * 12 * time.Duration(units)).Format(layout)
		if startTime != "" {
			creatTime = startTime
		}

		dates := 12
		parse, err := time.Parse(layout, creatTime)
		if err != nil {
			return nil, err
		}
		ds := time.Hour * 24 * time.Duration(units)
		for i := 0; i < dates; i++ {
			etime := ""
			if i != 0 {
				parse = parse.Add(ds)
			}
			etime = parse.Format(layout)
			fmt.Printf("%s", etime)
			if units == 365 {
				etime = fmt.Sprintf("%d", parse.Year())
			}
			if units == 7 {
				etime = fmt.Sprintf("%d-%d", parse.Year(), xkutils.WeekByDate(parse))
			}
			if units == 31 {
				etime = fmt.Sprintf("%d-%d", parse.Year(), parse.Month())
			}
			data = append(data, domain.AssetsTimeInfo{
				Scale:      "0",
				CreateTime: fmt.Sprintf("%s", etime),
				Price:      decimal.Zero,
				Freeze:     decimal.Zero,
				Nums:       decimal.Zero,
				Valuation:  decimal.Zero,
			})
			if etime == endTime {
				fmt.Printf("%s", etime)
				return data, nil
			}
		}
	}
	return data, nil
}

func FindFinanceAssetsList(as *domain.AssetsSelect) (map[string]interface{}, error) {

	asDao := assets.NewEntity()
	financeDao := financeAssets.NewEntity()
	merchantAsset, count, err := asDao.FindFinanceAssetsList(as)
	if err != nil {
		return nil, err
	}
	list, err := financeDao.FindFinanceAssetList(as.CoinId)
	if err != nil {
		return nil, err
	}
	mp := map[string]interface{}{
		"merchantsList":  merchantAsset,
		"merchantsTotal": count,
		"financeList":    list,
	}
	return mp, err
}
