package service

import (
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/shopspring/decimal"
)

type BalanceService interface {
	//获取主链币余额
	GetMchBalance(coinName string, mchId int) (decimal.Decimal, error)

	//获取某个主链币的代币余额
	GetMchTokenBalance(coinName string, tokenName string, mchId int) (decimal.Decimal, error)

	//所有币种余额
	GetMchAllBalance(mchId int) ([]*model.CoinBalance, error)

	GetMchAllBalanceV2(mchId int) ([]*model.CoinBalance, error)

	// 获取指定币种单地址最大余额
	GetMchCoinMaxBalance(coinName, tokenName string, mchId int) (decimal.Decimal, string, error)

	//获取商户的可用余额
	GetMchActivityBalance(coinName string, contractAddress string, mchId int) (decimal.Decimal, error)

	//获取商户前20个地址的总余额
	GetTopsTwentyAddresses(coinName string, contractAddress string, mchId int) (decimal.Decimal, error)
}
