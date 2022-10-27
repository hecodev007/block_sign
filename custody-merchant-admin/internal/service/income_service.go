package service

import (
	"bytes"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/businessOrder"
	"custody-merchant-admin/model/businessPackage"
	"custody-merchant-admin/model/chainBill"
	"custody-merchant-admin/model/incomeAccount"
	"errors"
	"fmt"
	"github.com/onethefour/common/xutils"
	"github.com/shopspring/decimal"
	"time"
)

func FindIncomePage(search *domain.SearchIncome) (domain.IncomeList, int64, error) {
	return deals.FindIncomePage(search)
}

func FindIncomeChart(search *domain.SearchIncome) (domain.IncomeList, error) {
	return deals.FindIncomeChart(search)
}

func FindIncomeExcelExport(search *domain.SearchIncome) (bytes.Buffer, error) {
	return deals.FindIncomeExcelExport(search)
}

func SaveIncome(billData *chainBill.Entity, state int) error {
	var (
		iDao = incomeAccount.NewEntity()
		bp   = businessPackage.NewEntity()
		bo   = businessOrder.NewEntity()
	)

	// 先查业务线用的套餐Id
	err := bp.FindBPItemByBusinessId(int64(billData.ServiceId))
	if err != nil {
		return err
	}
	if bp.Id == 0 {
		return errors.New(fmt.Sprintf(global.DataBusinessComboIsNil, billData.ServiceId))
	}
	// 加锁防止数据量过大
	keys := fmt.Sprintf("Income-%d%d%d%d", bp.Id, billData.MerchantId, billData.ServiceId, billData.CoinId)
	if err := xutils.LockMax(keys, 1000); err != nil {
		fmt.Println(global.DataIsMore)
		return errors.New(global.DataIsMore)
	}
	defer xutils.UnlockDelay(keys, 1)

updateLoop:
	income, err := bo.SumNumsBusinessOrder(billData.CoinId, int64(billData.ServiceId), int64(bp.Id), billData.MerchantId)
	if err != nil {
		return err
	}
	err = iDao.FindInfo(&domain.SearchIncome{
		ComboId:    bp.Id,
		MerchantId: billData.MerchantId,
		ServiceId:  billData.ServiceId,
		CoinId:     billData.CoinId,
	})
	if err != nil {
		return err
	}
	if iDao.Id != 0 {
		iDao.MinerFee = billData.UpChainFee
		// 提现
		if state == 0 {
			iDao.WithdrawIncome = iDao.WithdrawIncome.Add(billData.Fee.Sub(billData.UpChainFee)) // 提现收益
			iDao.WithdrawNums += 1                                                               // 提现笔数
			iDao.WithdrawPrice = iDao.WithdrawPrice.Add(billData.Nums)                           // 提现金额
			iDao.WithdrawFee = iDao.WithdrawFee.Add(billData.Fee)                                // 提现手续费
			iDao.WithdrawDestroy = iDao.WithdrawDestroy.Add(billData.DestroyFee)                 // 提现销毁
		}
		// 充值
		if state == 1 {
			iDao.TopUpIncome = iDao.TopUpIncome.Add(billData.Fee.Sub(billData.UpChainFee)) // 充值收益
			iDao.TopUpPrice = iDao.TopUpPrice.Add(billData.Nums)                           // 充值金额
			iDao.TopUpNums += 1                                                            // 充值笔数
			iDao.ToUpFee = iDao.ToUpFee.Add(billData.Fee)                                  // 充值手续费
			iDao.ToUpDestroy = iDao.ToUpDestroy.Add(billData.DestroyFee)                   // 充值销毁
		}

		iDao.ComboIncome = income // 因为还没得知上链费多少，以所收到的手续费为基础
		iDao.MinerFee = iDao.MinerFee.Add(billData.UpChainFee)
		mp := map[string]interface{}{
			"top_up_nums":      iDao.TopUpNums,
			"withdraw_price":   iDao.WithdrawPrice,
			"top_up_price":     iDao.TopUpPrice,
			"miner_fee":        iDao.MinerFee,
			"withdraw_destroy": iDao.WithdrawDestroy,
			"combo_income":     income,
			"top_up_destroy":   iDao.ToUpDestroy,
			"withdraw_income":  iDao.WithdrawIncome,
			"withdraw_nums":    iDao.WithdrawNums,
			"top_up_income":    iDao.TopUpIncome,
			"withdraw_fee":     iDao.WithdrawFee,
			"top_up_fee":       iDao.ToUpFee,
			"version":          iDao.Version + 1,
		}

		// 更新
		count, err := iDao.UpdateInfoByVersion(iDao.Id, iDao.Version, mp)
		if err != nil {
			return err
		}
		if count == 0 {
			goto updateLoop
		}
	} else {
		iDao.ComboId = bp.Id
		iDao.ComboModelName = bp.ModelName
		iDao.ComboTypeName = bp.TypeName
		iDao.ComboIncome = income // 因为还没得知上链费多少，以所收到的手续费为基础
		iDao.ServiceId = billData.ServiceId
		iDao.MerchantId = billData.MerchantId
		iDao.CoinId = billData.CoinId
		// 提现
		if state == 0 {
			iDao.WithdrawIncome = billData.Fee.Sub(billData.UpChainFee) // 提现收益，暂时无收益，因为还没得知上链费多少，以所收到的手续费为基数
			iDao.WithdrawNums = 1                                       // 提现笔数
			iDao.WithdrawPrice = billData.Nums                          // 提现金额
			iDao.WithdrawFee = billData.Fee                             // 提现手续费
			iDao.WithdrawDestroy = decimal.Zero                         // 提现销毁
		}
		// 充值
		if state == 1 {
			iDao.TopUpIncome = billData.Fee.Sub(billData.UpChainFee) // 充值收益，暂时无收益，因为还没得知上链费多少，以所收到的手续费
			iDao.TopUpPrice = billData.Nums                          // 充值金额
			iDao.TopUpNums = 1                                       // 充值笔数
			iDao.ToUpFee = billData.Fee                              // 充值手续费
			iDao.ToUpDestroy = decimal.Zero                          // 充值销毁
		}
		iDao.MinerFee = billData.UpChainFee // 矿工费
		iDao.CreatedAt = time.Now().Local()
		// 创建
		err = iDao.CreateIncome()
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveComboIncome
// 续费、开通的时候添加套餐收益户
// comboNums 收益的主链币数量
// merchantId 商户Id
// serviceId 业务线Id
// chainName 主链名
func SaveComboIncome(comboNums decimal.Decimal, merchantId, serviceId int64, chainName string) error {
	var (
		iDao = incomeAccount.NewEntity()
		bp   = businessPackage.NewEntity()
		bo   = businessOrder.NewEntity()
	)

	coin, err := base.FindCoinsByName(chainName)
	if err != nil {
		return err
	}
	// 先查业务线用的套餐Id
	err = bp.FindBPItemByBusinessId(serviceId)
	if err != nil {
		return err
	}
	if bp.Id == 0 {
		return errors.New(fmt.Sprintf(global.DataBusinessComboIsNil, serviceId))
	}
	// 加锁防止数据量过大
	keys := fmt.Sprintf("Income-%d%d%d%d", bp.Id, merchantId, serviceId, coin.Id)
	if err := xutils.LockMax(keys, 1000); err != nil {
		fmt.Println(global.DataIsMore)
		return errors.New(global.DataIsMore)
	}
	defer xutils.UnlockDelay(keys, 1)
	// 更新轮询，防止数据被其他线程抢占
updateLoop:
	income, err := bo.SumNumsBusinessOrder(int(coin.Id), serviceId, int64(bp.Id), merchantId)
	if err != nil {
		return err
	}
	err = iDao.FindInfo(&domain.SearchIncome{
		ComboId:    bp.Id,
		MerchantId: merchantId,
		ServiceId:  int(serviceId),
		CoinId:     int(coin.Id),
	})
	if err != nil {
		return err
	}
	if iDao.Id != 0 {
		mp := map[string]interface{}{
			"combo_income": income.Add(comboNums),
			"version":      iDao.Version + 1,
		}
		// 更新
		count, err := iDao.UpdateInfoByVersion(iDao.Id, iDao.Version, mp)
		if err != nil {
			return err
		}
		if count == 0 {
			goto updateLoop
		}
	} else {
		iDao.ComboId = bp.Id
		iDao.ComboModelName = bp.ModelName
		iDao.ComboTypeName = bp.TypeName
		iDao.ComboIncome = comboNums // 因为还没得知上链费多少，以所收到的手续费为基础
		iDao.ServiceId = int(serviceId)
		iDao.MerchantId = merchantId
		iDao.CoinId = int(coin.Id)
		// 套餐收益
		iDao.WithdrawIncome = decimal.Zero  // 提现收益，暂时无收益，因为还没得知上链费多少，以所收到的手续费为基数
		iDao.WithdrawNums = 0               // 提现笔数
		iDao.WithdrawPrice = decimal.Zero   // 提现金额
		iDao.WithdrawFee = decimal.Zero     // 提现手续费
		iDao.WithdrawDestroy = decimal.Zero // 提现销毁
		iDao.MinerFee = decimal.Zero
		iDao.CreatedAt = time.Now().Local()
		// 创建
		err = iDao.CreateIncome()
		if err != nil {
			return err
		}
	}
	return nil
}
