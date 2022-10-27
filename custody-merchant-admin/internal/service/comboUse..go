package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model/businessPackage"
	"custody-merchant-admin/model/comboUse"
	"errors"
	"github.com/shopspring/decimal"
	"time"
)

// SumAddrNums
// 统计套餐地址数
func SumAddrNums(nums, sid int64) error {
	cu := comboUse.NewEntity()
	bp := businessPackage.NewEntity()
	bps, err := bp.FindBPItemBySId(sid)
	if err != nil {
		return err
	}
	if bps.Id == 0 {
		return errors.New("业务线没开通套餐")
	}
	if bps.AddrNums == 0 {
		return nil
	}
	sumCombo, err := cu.SumComboUserDayByCId(bps.AccountId, bps.PackageId)
	if err != nil {
		return err
	}
	if sumCombo == nil || sumCombo.Id == 0 {
		return nil
	}
	if (sumCombo.UsedAddrDay + nums) > int64(bps.AddrNums) {
		return errors.New("套餐地址数不足")
	}
	return nil
}

// UpAddrComboUse
// 更新地址套餐使用
func UpAddrComboUse(nums, sid int64) error {

	cu := comboUse.NewEntity()
	bp := businessPackage.NewEntity()
	bps, err := bp.FindBPItemBySId(sid)
	if err != nil {
		return err
	}
	if bps.Id == 0 {
		return errors.New("业务线没开通套餐")
	}

	comboInfo, err := cu.FindComboUserDayByCId(bps.AccountId, bps.PackageId, time.Now().Local().Format(global.YyyyMmDd))
	if err != nil {
		return err
	}
	if bps.TypeName == "地址收费套餐" {
		_, err = bp.UpdateBPItemUsdBySIdByMap(bps.BusinessId, map[string]interface{}{
			"had_used": bps.HadUsed.Add(decimal.NewFromInt(nums)),
		})
	} else {
		return nil
	}
cLoop:
	if comboInfo != nil {
		up, err := cu.UpDateComboUserDayByCId(bps.AccountId, bps.PackageId,
			time.Now().Local().Format(global.YyyyMmDd), comboInfo.Version,
			map[string]interface{}{
				"used_addr_day": nums + comboInfo.UsedAddrDay,
				"version":       comboInfo.Version + 1,
			})
		if err != nil {
			return err
		}
		if up == 0 {
			goto cLoop
		}
	} else {
		_, err := cu.CreateComboUserDay(comboUse.Entity{
			ComboUserId: bps.PackageId,
			UsedAddrDay: nums,
			UsedLineDay: decimal.Zero,
			CreateTime:  time.Now().Local(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
