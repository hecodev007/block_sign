package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/record"
	"strings"
)

// SearchRecords 搜索套餐列表
func SearchRecords(req *domain.RecordReqInfo) (list []domain.RecordInfo, total int64, err error) {
	var l []record.Entity
	pInfo := record.NewEntity()
	l, total, err = pInfo.FindPackageListByReq(*req)
	if err != nil {
		return list, total, global.DaoError(err)
	}
	list = make([]domain.RecordInfo, 0)
	for _, item := range l {
		operate := OperateName(item.Operate)
		rInfo := domain.RecordInfo{
			OperatorName: item.OperatorName,
			Remark:       item.Remark,
			Operate:      operate,
			CreatedAt:    item.CreatedAt.Local().Format(global.YyyyMmDd),
		}
		list = append(list, rInfo)
	}
	return
}

func SearchFinanceListByReq(req *domain.MerchantReqInfo) (list []domain.FinanceRecordInfo, total int64, err error) {
	pInfo := record.NewEntity()
	items, total, err := pInfo.FindFinanceListByReq(*req)
	list = make([]domain.FinanceRecordInfo, 0)
	for _, item := range items {
		var isLock int
		var isLockFinance int
		//冻结用户和资产lock_user，解冻用户和资产unlock_user，冻结资产lock_asset，，解冻资产unlock_asset
		if strings.Contains(item.Operate, "lock_asset") {
			isLockFinance = 1
		}
		if strings.Contains(item.Operate, "lock_user") {
			isLock = 1
		}
		newItem := domain.FinanceRecordInfo{
			IsLock:        isLock,
			IsLockFinance: isLockFinance,
			OperatorName:  item.OperatorName,
			Remark:        item.Remark,
			CreatedAt:     item.CreatedAt.Format(global.YyyyMmDdHhMmSs),
		}
		list = append(list, newItem)
	}
	return
}
