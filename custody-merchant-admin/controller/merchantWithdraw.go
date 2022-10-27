package controller

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/util/xkutils"
	"github.com/shopspring/decimal"

	"custody-merchant-admin/module/blockChainsApi"
	"custody-merchant-admin/router/web/handler"
	"fmt"
)

//直接钱包提现
func WalletWithdraw(c *handler.Context) error {
	req := domain.BCWithDrawReq{
		ApiKey:          "oqwjQIgxnSMxTELIUkkPtGAxZzxZsljt",
		OutOrderId:      xkutils.NewUUId("outOrderId"),
		CoinName:        "ftm",
		Amount:          decimal.NewFromFloat(0.001),
		ToAddress:       "0x32e5a45c81370c81a465052238d3e8ceb0017d01",
		TokenName:       "",
		ContractAddress: "",
		Memo:            "",
	}
	err := blockChainsApi.BlockChainWithdrawCoin(req, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
	fmt.Printf("\n WalletWithdraw req=%+v \n err = %v\n", req, err)
	return nil
}
