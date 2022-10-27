package deals

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/limit"
	"custody-merchant-admin/model/order"
	"custody-merchant-admin/model/services"
	"github.com/shopspring/decimal"
	"time"
)

// WithdrawalLimit
// 业务线提币判断
// 5分钟内提币数量、次数限制
// 1小时内提币数量、次数限制
func WithdrawalLimit(sId int, addr string, nums decimal.Decimal) error {
	var (
		withdrawalDao = new(limit.LimitWithdrawal)
		orderDao      = new(order.Orders)
	)
	withdrawal, err := withdrawalDao.FindLimitWithdrawal(sId)
	if err != nil {
		return err
	}
	if withdrawal.Id > 0 {
		nowTime := time.Now().Local()
		startTime := nowTime.Format(global.YyyyMmDdHhMmSs)
		endTime := nowTime.Add(-(time.Duration(5) * time.Minute)).Format(global.YyyyMmDdHhMmSs)
		mNums, err := orderDao.FindOrderByTimeNums(startTime, endTime, addr, sId)
		if err != nil {
			return err
		}
		endTime = nowTime.Add(-(time.Duration(1) * time.Hour)).Format(global.YyyyMmDdHhMmSs)
		hNums, err := orderDao.FindOrderByTimeNums(startTime, endTime, addr, sId)
		if err != nil {
			return err
		}
		if !withdrawal.LineMinutes.IsZero() && withdrawal.LineMinutes.LessThanOrEqual(mNums.Nums.Add(nums)) {
			return global.WarnMsgError(global.MsgWithdrawLimitMinuteNums)
		}
		if !withdrawal.LineHours.IsZero() && withdrawal.LineHours.LessThanOrEqual(hNums.Nums.Add(nums)) {
			return global.WarnMsgError(global.MsgWithdrawLimitHourNums)
		}
		if withdrawal.NumMinutes != 0 && withdrawal.NumMinutes <= mNums.Counts {
			return global.WarnMsgError(global.MsgWithdrawLimitMinuteCount)
		}
		if withdrawal.NumHours != 0 && withdrawal.NumHours <= hNums.Counts {
			return global.WarnMsgError(global.MsgWithdrawLimitHourCount)
		}
	}
	return nil
}

// ServiceConfigLimit
// 业务线门槛判断
func ServiceConfigLimit(sId, aid int, nums, dNums, wNums, mNums decimal.Decimal) error {
	s := new(services.ServiceAuditConfig)
	transfer, err := s.GetServiceConfigBySLid(sId, aid)
	if err != nil {
		return err
	}
	if transfer.Id > 0 {
		if !transfer.NumEach.IsZero() && transfer.NumEach.LessThanOrEqual(nums) {
			return global.WarnMsgError(global.MsgTransferLimitEachNums)
		}
		if !transfer.NumDay.IsZero() && transfer.NumDay.LessThanOrEqual(dNums) {
			return global.WarnMsgError(global.MsgTransferLimitDayNums)
		}
		if !transfer.NumWeek.IsZero() && transfer.NumWeek.LessThanOrEqual(wNums) {
			return global.WarnMsgError(global.MsgTransferLimitWeekNums)
		}
		if !transfer.NumMonth.IsZero() && transfer.NumMonth.LessThanOrEqual(mNums) {
			return global.WarnMsgError(global.MsgTransferLimitMonthNums)
		}
	}
	return nil
}

// 查询日周月的转帐金额
func FindLimitNums(startTime, addr string, sid int) (domain.OrderLimitNums, error) {
	var (
		orderDao = new(order.Orders)
		oln      = domain.OrderLimitNums{}
	)
	dNums, err := orderDao.FindOrderByDayNums(startTime, addr, sid)
	if err != nil {
		return oln, err
	}
	oln.DNums = dNums
	wNums, err := orderDao.FindOrderByWeekNums(startTime, addr, sid)
	if err != nil {
		return oln, err
	}
	oln.WNums = wNums
	mNums, err := orderDao.FindOrderByMonthNums(startTime, addr, sid)
	if err != nil {
		return oln, err
	}
	oln.MNums = mNums
	return oln, err
}

// TransferLimit
// 业务线转出判断
func TransferLimit(sId int, addr string, nums decimal.Decimal) error {

	var (
		orderDao            = order.Orders{}
		transferDao         = limit.LimitTransfer{}
		dNums, wNums, mNums decimal.Decimal
		err                 error
	)
	nowTime := time.Now().Local()
	startTime := nowTime.Format(global.YyyyMmDdHhMmSs)

	transfer, err := transferDao.FindLimitTransfer(sId)
	if err != nil {
		return err
	}
	if transfer.Id > 0 {
		dNums, err = orderDao.FindOrderByDayNums(startTime, addr, sId)
		if err != nil {
			return err
		}
		wNums, err = orderDao.FindOrderByWeekNums(startTime, addr, sId)
		if err != nil {
			return err
		}
		mNums, err = orderDao.FindOrderByMonthNums(startTime, addr, sId)
		if err != nil {
			return err
		}
		if transfer.NumEach.LessThanOrEqual(nums) {
			return global.WarnMsgError(global.MsgTransferLimitEachNums)
		}
		if transfer.NumDay.LessThanOrEqual(dNums) {
			return global.WarnMsgError(global.MsgTransferLimitDayNums)
		}
		if transfer.NumWeeks.LessThanOrEqual(wNums) {
			return global.WarnMsgError(global.MsgTransferLimitWeekNums)
		}
		if transfer.NumMonth.LessThanOrEqual(mNums) {
			return global.WarnMsgError(global.MsgTransferLimitMonthNums)
		}
	}
	return nil
}
