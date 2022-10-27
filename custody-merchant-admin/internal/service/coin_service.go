package service

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

//UpdateCoinDB 更新主链币表
func UpdateCoinDB(list []domain.BCCoinInfo) (err error) {
	//查询已有的
	var had []base.ChainInfo
	had, err = base.FindAllChainCoins()

	lMap := make(map[string]int)        //需要insert
	change := make([]base.ChainInfo, 0) //需要update
	del := make([]base.ChainInfo, 0)    //需要删除的

	for i, item := range list {
		lMap[strings.ToUpper(item.Name)] = i
	}

	for _, i := range had {
		var ishave bool
		for _, j := range list {
			iName := strings.ToUpper(i.Name)
			jName := strings.ToUpper(j.Name)
			jfName := strings.ToUpper(j.FullName)
			if (iName == jName || iName == jfName) && iName != "" {
				ishave = true
				var isChange bool
				delete(lMap, strings.ToUpper(j.Name))
				if j.State == int64(i.State) {
					if i.State == 0 {
						i.State = 1
					} else if i.State == 1 {
						i.State = 0
					} else {
						i.State = 0
					}
					i.PriceUsd = xkutils.StringToDecimal(j.PriceUsd)
					i.Name = strings.ToUpper(i.Name)
					isChange = true
				}
				p, _ := decimal.NewFromString(j.PriceUsd)
				if i.PriceUsd.Cmp(p) != 0 {
					i.PriceUsd = p
					isChange = true
				}
				if isChange {
					i.Name = strings.ToUpper(i.Name)
					change = append(change, i)
				}
				break
			}
		}
		//旧数据多余的币，需要删除的
		if !ishave {
			del = append(del, i)
		}
	}

	//insert
	iArr := make([]base.ChainInfo, 0)
	for _, v := range lMap {
		item := list[v]
		newI := base.ChainInfo{
			Name:       strings.ToUpper(item.Name),
			PriceUsd:   xkutils.StringToDecimal(item.PriceUsd),
			CreateTime: time.Now().Local(),
			UpdateTime: time.Now().Local(),
		}
		if item.Name != "" {
			iArr = append(iArr, newI)
		}
	}
	//insert
	if len(iArr) > 0 {
		err = base.InsertChainCoins(iArr)
		if err != nil {
			log.Error(" 更新主链币表 insert err:", err.Error())
			return
		}
	}
	//update
	if len(change) > 0 {
		err = base.UpdateChainCoins(change)
		if err != nil {
			log.Error(" 更新主链币表 update err:", err.Error())
			return
		}
	}
	////delete
	//if len(del) > 0 {
	//	err = base.DelChainCoins(del)
	//	if err != nil {
	//		log.Error(" 更新主链币表 del err:", err.Error())
	//		return
	//	}
	//}

	return err

}

//UpdateSubCoinDB 更新代币表
func UpdateSubCoinDB(list []domain.BCCoinInfo) (err error) {
	//查询已有的
	var chainhad []base.ChainInfo
	chainhad, err = base.FindAllChainCoins()
	chainMap := make(map[string]int)
	for _, item := range chainhad {
		chainMap[strings.ToUpper(item.Name)] = item.Id
	}

	var had []base.CoinInfo
	had, err = base.FindAllCoins()

	lMap := make(map[string]int)       //需要insert
	change := make([]base.CoinInfo, 0) //需要update
	del := make([]base.CoinInfo, 0)    //需要删除的

	for i, item := range list {
		lMap[strings.ToUpper(item.Name)] = i
	}

	for _, i := range had {
		var ishave bool
		for _, j := range list {
			iName := strings.ToUpper(i.Name)
			ifName := strings.ToUpper(i.FullName)
			jName := strings.ToUpper(j.Name)
			jfName := strings.ToUpper(j.FullName)
			if (iName == jName && iName != "") || (iName == jfName && iName != "") || (ifName == jfName && ifName != "") { //可用状态变了
				ishave = true
				delete(lMap, strings.ToUpper(j.Name))
				var isChange bool
				if j.State == int64(i.State) {
					if i.State == 0 {
						i.State = 1
					} else if i.State == 1 {
						i.State = 0
					} else {
						i.State = 0
					}
					isChange = true
				}

				if i.Token != j.Token { //合约地址变了
					i.Token = j.Token
					isChange = true
				}
				fname := j.Father
				if j.Father == "" {
					fname = j.Name
				}
				fId := chainMap[strings.ToUpper(fname)]
				if fId != i.ChainId { //主链id变了
					i.ChainId = fId
					isChange = true
				}
				p, _ := decimal.NewFromString(j.PriceUsd)
				if i.PriceUsd.Cmp(p) != 0 {
					i.PriceUsd = p
					isChange = true
				}

				if j.Confirm != i.Confirm {
					i.Confirm = j.Confirm
					isChange = true
				}
				if isChange {
					i.Name = strings.ToUpper(i.Name)
					change = append(change, i)
				}
				break
			}
		}
		//旧数据多余的币，需要删除的
		if !ishave {
			del = append(del, i)
		}
	}

	//insert
	iArr := make([]base.CoinInfo, 0)
	for _, v := range lMap {
		item := list[v]
		var fId int
		if item.Father != "" {
			fId = chainMap[strings.ToUpper(item.Father)]
		}

		newI := base.CoinInfo{
			Name:       strings.ToUpper(item.Name),
			ChainId:    fId,
			FullName:   item.FullName,
			Token:      item.Token,
			Confirm:    item.Confirm,
			CreateTime: time.Now().Local(),
			UpdateTime: time.Now().Local(),
		}
		if item.Name != "" {
			iArr = append(iArr, newI)
		}

	}

	//insert
	if len(iArr) > 0 {

		err = base.InsertSubCoins(iArr)
		if err != nil {
			log.Error(" 更新coin_info表 insert err:", err.Error())
			return err
		}
	}
	//update
	if len(change) > 0 {
		err = base.UpdateSubCoins(change)
		if err != nil {
			log.Error(" 更新coin_info表 update err:", err.Error())
			return err
		}
	}

	////delete
	//if len(del) > 0 {
	//	err = base.DelSubCoins(del)
	//	if err != nil {
	//		log.Error(" 更新coin_info表 del err:", err.Error())
	//		return err
	//	}
	//}
	return err

}

func GetCoinInfo(coinName string) (info base.CoinInfo, err error) {
	return base.FindCoinsByName(coinName)
}

func GetDeductCoinName(coinName string) (arr string) {
	c := strings.Split(coinName, ",")
	items, _ := base.FindCoinsInIds(c)
	s := make([]string, 0)
	for _, item := range items {
		s = append(s, item.Name)
	}
	arr = strings.Join(s, ",")
	return
}

func GetDeductCoinId(coinName string) (arr string) {
	c := strings.Split(coinName, ",")
	items, _ := base.FindCoinsInName(c)
	s := make([]string, 0)
	for _, item := range items {
		s = append(s, item.Name)
	}
	arr = strings.Join(s, ",")
	return
}
