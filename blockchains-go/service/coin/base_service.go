package coin

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/service"
)

type CoinBaseService struct {
}

func (c *CoinBaseService) CloseAllCollect(pid int) error {
	return dao.FcCoinSetCloseAllCollect(pid)
}

func (c *CoinBaseService) CloseCollect(name string) error {
	return dao.FcCoinSetCloseCollect(name)
}

func (c *CoinBaseService) OpenCollect(name string, bof string) error {
	return dao.FcCoinSetOpenCollect(name, bof)
}

func (c *CoinBaseService) GetCoinList() ([]*model.CoinList, error) {
	datas, err := dao.FcCoinSetFindByStatus(1)
	if err != nil {
		return nil, err
	}
	results := make([]*model.CoinList, 0)
	for _, v := range datas {
		if v.Pid == 0 {
			results = append(results, &model.CoinList{
				Father:  "",
				Name:    v.Name,
				Token:   "",
				Decimal: v.Decimal,
			})
		} else {
			result, err := dao.FcCoinSetGetByStatus(v.Pid, 1)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, fmt.Errorf("coin :%s,miss parent info")
			}
			results = append(results, &model.CoinList{
				Father:  result.Name,
				Name:    v.Name,
				Token:   v.Token,
				Decimal: v.Decimal,
			})
		}
	}
	return results, nil
}


func (c *CoinBaseService) CustodyGetCoinList(limit int,offset int) ([]*model.CustodyCoinList, error) {
	datas, err := dao.FcCoinList(limit,offset)
	if err != nil {
		return nil, err
	}
	results := make([]*model.CustodyCoinList, 0)
	for _, v := range datas {
		if v.Pid == 0 {
			results = append(results, &model.CustodyCoinList{
				Father:  "",
				Name:    v.Name,
				Token:   "",
				Decimal: v.Decimal,
				State: v.Status,
				PriceUsd: v.Price,
			})
		} else {
			result, err := dao.FcCoinSetGetByStatus(v.Pid, 1)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, fmt.Errorf("coin :%s,miss parent info")
			}
			results = append(results, &model.CustodyCoinList{
				Father:  result.Name,
				Name:    v.Name,
				Token:   v.Token,
				Confirm: v.Confirm,
				Decimal: v.Decimal,
				State: v.Status,
				PriceUsd: v.Price,
			})
		}
	}
	return results, nil
}

func NewCoinBaseService() service.CoinService {
	return &CoinBaseService{}
}
