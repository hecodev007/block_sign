package service

import "github.com/group-coldwallet/blockchains-go/model"

type CoinService interface {
	//查询币种列表
	GetCoinList() ([]*model.CoinList, error)
	CustodyGetCoinList(limit int,offset int) ([]*model.CustodyCoinList, error) //托管后台所需币列表信息
	CloseAllCollect(pid int) error
	CloseCollect(name string) error
	OpenCollect(name string, bof string) error
}
