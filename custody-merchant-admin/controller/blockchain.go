package controller

import (
	"bytes"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/router/web/handler"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"strings"
)

//BlockchainCallback 提现上链回调
func BlockchainCallback(c *handler.Context) error {

	req := new(domain.InComeBack)
	err := c.DefaultBinder(req)
	if err != nil {
		log.Errorf("提现上链回调 err %+v\n", err)
		return handler.NewError(c, err.Error())
	}

	if err != nil {
		log.Errorf("bind err %v \n", err)
		return handler.NewError(c, err.Error())
	}

	WithdrawCallBack(*req)
	res := handler.NewResult(0, "")
	return res.ResultOk(c)
}

//BlockchainIncomeCallback 充值回调
func BlockchainIncomeCallback(c *handler.Context) error {
	//InComeBack
	data, _ := ioutil.ReadAll(c.Request().Body)
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(data))
	req := new(domain.InComeBack)
	err := c.Binder(req)
	if err != nil {
		log.Errorf("钱包回调 err %+v \n", err)
		return handler.NewError(c, err.Error())
	}
	log.Infof("钱包回调 data %s", string(data))
	var reqM map[string]interface{}
	json.Unmarshal(data, &reqM)

	if err != nil {
		log.Errorf("bind err %v \n", err)
		return handler.NewError(c, err.Error())
	}

	if req.IsIn == 2 {
		log.Errorf("提现回调 %+v\n", req)
		err = WithdrawCallBack(*req)
	} else {
		log.Errorf("充值回调 %+v\n", req)
		coins, err := base.FindCoinsByName(strings.ToUpper(req.Coin))
		if err != nil {
			return handler.NewError(c, err.Error())
		}
		if req.Confirmations < coins.Confirm {
			// 更新资产
			log.Errorf("充值回调 接收 更新数小于设置值，无后续处理 %v\n", req)
			res := handler.NewResult(0, "")
			return res.ResultOk(c)
		}
		service.InComeCallBack(req)
	}
	if err != nil {
		log.Errorf("钱包回调 接收 err %v\n", err)
	}
	res := handler.NewResult(0, "")
	return res.ResultOk(c)
}

func WithdrawCallBack(req domain.InComeBack) error {
	bills, err := service.FindBillByOrderId(req.OutOrderId)
	if err != nil {
		log.Errorf("FindBillByOrderId err %v \n", err)
		return err
	}
	coins, err := base.FindCoinsById(bills.CoinId)
	if err != nil {
		return err
	}
	// 提现到账，更新
	wallets := domain.MqWalletInfo{}
	wallets.SerialNo = req.OutOrderId
	wallets.TxId = req.Txid
	wallets.ConfirmNums = int(req.Confirmations)
	wallets.Height = int(req.BlockHeight)
	wallets.RealNums = decimal.NewFromFloat(req.Amount)
	wallets.Nums = bills.Nums
	wallets.MinerFee = decimal.NewFromFloat(req.Fee)
	if req.Confirmations == 0 {
		err = errors.New("更新数为0")
		log.Errorf("FreezeBillDetailState %v \n", err)
		return err
	}
	if req.Confirmations >= coins.Confirm {
		// 更新资产
		err = service.FreezeBillDetailState(wallets, 1)
		log.Errorf("FreezeBillDetailState %v \n", err)
		return err
	} else {
		// 更新链上订单
		err = service.UpdateConfirmNums(wallets)
		log.Errorf("UpdateConfirmNums %v \n", err)
		return err
	}
}
