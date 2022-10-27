package controller

import (
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/runtime/job"
)

func UpdateCoin(c *handler.Context) error {

	j := job.WalletCoinListCallBackJob{}
	j.Run()
	return nil

}
