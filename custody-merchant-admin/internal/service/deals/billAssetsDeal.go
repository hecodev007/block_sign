package deals

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model/assets"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/bill"
	"custody-merchant-admin/model/chainBill"
	"custody-merchant-admin/model/serviceChains"
	"custody-merchant-admin/module/log"
	"errors"
	"fmt"
	"github.com/onethefour/common/xutils"
	"github.com/shopspring/decimal"
)

func FindAssetBySIdAndCId(sid, cid int) (*assets.Assets, error) {
	asDao := assets.NewEntity()

	asset, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return asset, err
	}
	return asset, nil
}

// WithdrawalFreezeAssets
// 提现账单，资产冻结
func WithdrawalFreezeAssets(sid, cid int, nums, fee decimal.Decimal) error {

	asDao := assets.NewEntity()
	asset, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	if asset == nil || asset.CoinName == "" {
		return errors.New("业务线币种不存在")
	}
	keys := fmt.Sprintf("Assets-%s%d%d%d", asset.CoinName, asset.Id, asset.ServiceId, asset.CoinId)
	if err := xutils.LockMax(keys, 1000); err != nil {
		fmt.Println(global.DataIsMore)
		return errors.New(global.DataIsMore)
	}
	defer xutils.UnlockDelay(keys, 10)

asUpdateLoop:
	asInfo, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	assetsNums := asInfo.Nums.Sub(nums).Sub(fee)
	up, err := asDao.UpDateAssetsBySCId(sid, cid, asInfo.Version, map[string]interface{}{
		"nums":           assetsNums,
		"freeze":         asInfo.Freeze.Add(nums),
		"finance_freeze": asInfo.FinanceFreeze.Add(fee),
		"version":        asInfo.Version + 1,
	})
	if err != nil {
		return err
	}
	if up == 0 {
		goto asUpdateLoop
	}
	return nil
}

// WithdrawalConfirmAssets
// 提现账单确认
// sid 业务线ID
// cid 币种(代币)ID
// nums 确认的实际数量
func WithdrawalConfirmAssets(sid, cid int, nums decimal.Decimal) error {

	asDao := assets.NewEntity()
	asset, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	if asset == nil || asset.CoinName == "" {
		return errors.New("业务线币种不存在")
	}
	keys := fmt.Sprintf("Assets-%s%d%d%d", asset.CoinName, asset.Id, asset.ServiceId, asset.CoinId)
	if err := xutils.LockMax(keys, 1000); err != nil {
		fmt.Println(global.DataIsMore)
		return errors.New(global.DataIsMore)
	}
	defer xutils.UnlockDelay(keys, 10)
asUpdateLoop:
	asInfo, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	up, err := asDao.UpDateAssetsBySCId(sid, cid, asInfo.Version, map[string]interface{}{
		"freeze":  asInfo.Freeze.Sub(nums),
		"version": asInfo.Version + 1,
	})
	if err != nil {
		return err
	}
	if up == 0 {
		goto asUpdateLoop
	}

	return nil
}

// ReceiveFreezeAssets
// 接收账单，资产冻结
// nums 实际到账
// fee 手续费
func ReceiveFreezeAssets(sid, cid int, nums, fee decimal.Decimal) error {

	asDao := assets.NewEntity()
	asset, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	if asset == nil || asset.CoinName == "" {
		return errors.New("业务线币种不存在")
	}
	keys := fmt.Sprintf("Assets-%s%d%d%d", asset.CoinName, asset.Id, asset.ServiceId, asset.CoinId)
	if err := xutils.LockMax(keys, 1000); err != nil {
		fmt.Println(global.DataIsMore)
		return errors.New(global.DataIsMore)
	}
	defer xutils.UnlockDelay(keys, 10)

asUpdateLoop:
	asInfo, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	assetsNums := asInfo.Nums.Sub(fee)
	up, err := asDao.UpDateAssetsBySCId(sid, cid, asInfo.Version, map[string]interface{}{
		"nums":           assetsNums,
		"freeze":         asInfo.Freeze.Add(nums),
		"finance_freeze": asInfo.FinanceFreeze.Add(fee),
		"version":        asInfo.Version + 1,
	})
	if err != nil {
		return err
	}
	if up == 0 {
		goto asUpdateLoop
	}
	return nil
}

// ReceiveConfirmAssets
// 接收账单确认
// sid 业务线ID
// cid 币种(代币)ID
// nums 确认的实际数量
func ReceiveConfirmAssets(sid, cid int, nums decimal.Decimal) error {

	asDao := assets.NewEntity()
	asset, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	if asset == nil || asset.Id == 0 {
		return errors.New("业务线币种不存在")
	}
	keys := fmt.Sprintf("Assets-%s%d%d%d", asset.CoinName, asset.Id, asset.ServiceId, asset.CoinId)
	if err := xutils.LockMax(keys, 1000); err != nil {
		fmt.Println(global.DataIsMore)
		return errors.New(global.DataIsMore)
	}
	defer xutils.UnlockDelay(keys, 10)

asUpdateLoop:
	asInfo, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	up, err := asDao.UpDateAssetsBySCId(sid, cid, asInfo.Version, map[string]interface{}{
		"nums":    asInfo.Nums.Add(nums),
		"freeze":  asInfo.Freeze.Sub(nums),
		"version": asInfo.Version + 1,
	})
	if err != nil {
		return err
	}
	if up == 0 {
		goto asUpdateLoop
	}
	return nil
}

// RollbackBillAssets
// 处理提现账单资产，回滚账单
func RollbackBillAssets(serialNo string) error {
	var (
		billDao   = new(bill.BillDetail)
		chainbill = chainBill.NewEntity()
	)
	// 账单更新 - 0
	bmp := map[string]interface{}{
		"bill_status": 5,
		"tx_type":     5,
	}
	// 更新账单
	err := billDao.UpdateBillBySerialNo(serialNo, bmp)
	if err != nil {
		return err
	}
	bl, err := billDao.GetBillBySerialNo(serialNo)
	if err != nil {
		return err
	}
	err = chainbill.UpdatesChainBillBySerialNo(serialNo, bmp)
	if err != nil {
		return err
	}
	chain, err := base.FindChainsById(bl.ChainId)
	if err != nil {
		return err
	}
	coin, err := base.FindCoinsByName(chain.Name)
	if err != nil {
		return err
	}
	bs := serviceChains.NewEntity()
	err = bs.FindServiceChainsInfo(bl.ServiceId, chain.Name)
	if err != nil {
		return err
	}
	// 先回滚商户主链币的手续费：财务冻结->解冻
	err = RollbackAsset(bl.ServiceId, int(coin.Id), bl.SerialNo, true)
	if err != nil {
		return err
	}
	// 回滚被冻结的代币：解冻
	err = RollbackAsset(bl.ServiceId, bl.CoinId, bl.SerialNo, false)
	if err != nil {
		return err
	}
	return nil
}

// RollbackAsset
// 回滚解冻资产
// sid 业务线ID
// cid 币种(代币)ID
// serialNo 订单号
// isFee 是否回滚手续费
func RollbackAsset(sid, cid int, serialNo string, isFee bool) error {
	var (
		freeze  decimal.Decimal
		nums    decimal.Decimal
		billDao = new(bill.BillDetail)
		asDao   = assets.NewEntity()
	)
	ast, err := asDao.GetDateAssetsBySIdAndCId(sid, cid)
	if err != nil {
		return err
	}
	if ast == nil || ast.CoinName == "" {
		return errors.New("业务线币种不存在")
	}
	keys := fmt.Sprintf("Assets-%s%d%d%d", ast.CoinName, ast.Id, sid, cid)
	if err := xutils.LockMax(keys, 1000); err != nil {
		fmt.Println(global.MsgWarnSysBuss)
		return err
	}
	defer xutils.UnlockDelay(keys, 10)
asRebackLoop:
	// 先查询所有要更新的数据
	billInfo, err := billDao.GetBillBySerialNo(serialNo)
	if err != nil {
		return err
	}
	asset, err := asDao.GetDateAssetsBySIdAndCId(billInfo.ServiceId, cid)
	if err != nil {
		return err
	}
	fmt.Printf("%s", asset.Nums.String())
	mp := map[string]interface{}{}
	if isFee {
		// 回滚主链币的手续费
		mp = map[string]interface{}{
			"nums":           asset.Nums.Add(billInfo.Fee),
			"finance_freeze": asset.FinanceFreeze.Sub(billInfo.Fee),
			"version":        asset.Version + 1,
		}
	} else {
		// 回滚金额
		rebckNums := decimal.Decimal{}
		rebckNums = rebckNums.Add(billInfo.Nums)
		// 减去冻结
		freeze = asset.Freeze.Sub(rebckNums)
		// 增加剩余金额
		nums = asset.Nums.Add(rebckNums)
		// 回滚代币的冻结
		mp = map[string]interface{}{
			"nums":    nums,
			"freeze":  freeze,
			"version": asset.Version + 1,
		}
		coin, err := base.FindCoinsById(cid)
		if err != nil {
			return err
		}
		log.Infof("订单被拒绝，解冻金额并增加剩余金额 %v,该业务线%s,%s币余额 %v", rebckNums, billInfo.ServiceName, coin.Name, nums)
	}
	// 更新资产
	up, err := asDao.UpDateAssetsBySCId(billInfo.ServiceId, cid, asset.Version, mp)
	if err != nil {
		return err
	}
	if up == 0 {
		goto asRebackLoop
	}
	return nil
}
