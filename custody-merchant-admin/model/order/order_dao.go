package order

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/util/sql"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"github.com/shopspring/decimal"
	"time"
)

func (o *Orders) CreateOrderInfo(info *domain.OrderInfo) (int64, error) {

	orders := Orders{
		SerialNo:    info.SerialNo,
		TxId:        info.TxId,
		CoinId:      info.CoinId,
		ServiceId:   info.ServiceId,
		MerchantId:  info.MerchantId,
		Phone:       info.Phone,
		Type:        info.Type,
		OrderResult: 0,
		State:       0,
		FromAddr:    info.FromAddr,
		ReceiveAddr: info.ReceiveAddr,
		Nums:        info.Nums,
		Fee:         info.Fee,
		UpChainFee:  info.UpChainFee,
		BurnFee:     info.BurnFee,
		DestroyFee:  info.DestroyFee,
		RealNums:    info.RealNums,
		Memo:        info.Memo,
		CreateUser:  info.CreateUser,
		CreateTime:  time.Now().Local(),
	}
	creatOrder := model.DB().Omit("chain_id", "audit_result", "audit_type", "update_time", "chain_name", "coin_name", "service_name").Create(&orders)
	return orders.Id, model.ModelError(creatOrder, global.MsgWarnModelAdd)
}

func (o *Orders) FindOrderInfo(sid int) ([]Orders, error) {
	var order []Orders
	findOrder := model.DB().Where("service_id=? and order_result = 0", sid).Find(&order)
	return order, model.ModelError(findOrder, global.MsgWarnModelNil)
}

func (o *Orders) UpdateOrdersInfo(oId int64, mp map[string]interface{}) (int, error) {
	db := model.DB().Table("orders").Where("id = ? ", oId).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	return xkutils.ThreeDo(err != nil, 1, 0).(int), err
}

func (o *Orders) UpdateOrdersInfoBySerialNo(serialNo string, mp map[string]interface{}) error {
	db := model.DB().Table("orders").Where("serial_no = ? ", serialNo).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	return err
}

func (o *Orders) CountOrderStatus(uid int64) ([]CountOrders, error) {

	var (
		counts []CountOrders
		build  = new(sql.SqlBuilder)
	)
	build.SqlAdd("select sum(1) as count,orders.order_result as order_result from orders ").
		SqlAdd(" where (select count(1) from order_audit where order_audit.user_id = ? and order_audit.order_id = orders.id limit 1 ) > 0 ").
		SqlAdd(" group by orders.order_result order by orders.order_result asc")
	db := model.DB().Raw(build.ToSqlString(), uid).Scan(&counts)
	return counts, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderByTimeNums(startTime, endTime, addr string, sid int) (*WithdrawalOrderInfo, error) {
	var (
		build  = new(sql.SqlBuilder)
		assets WithdrawalOrderInfo
	)
	if model.FilteredSQLInject(startTime, endTime, addr) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	build.SqlAdd(" select sum(nums) as nums,count(id) as counts from orders").
		SqlWhere(" orders.create_time >= ?", startTime, true).
		SqlWhereVars(" and orders.create_time <= ? ", endTime, true).
		SqlAdd(" and receive_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)

	return &assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderByWeekNums(startTime, addr string, sid int) (decimal.Decimal, error) {
	var (
		build  = new(sql.SqlBuilder)
		assets ConfigNums
	)
	build.SqlAdd(" select sum(nums) as nums from orders").
		SqlEnd(" where DATE_FORMAT(create_time,'%Y-%u') = DATE_FORMAT(?,'%Y-%u')", startTime, true).
		SqlAdd(" and receive_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)
	return assets.Nums, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderByDayNums(startTime, addr string, sid int) (decimal.Decimal, error) {
	var (
		build  = new(sql.SqlBuilder)
		assets ConfigNums
	)
	build.SqlAdd(" select sum(nums) as nums from orders").
		SqlWhere(" DATE_FORMAT(create_time,'%Y-%m-%d') = DATE_FORMAT(?,'%Y-%m-%d')", startTime, true).
		SqlAdd(" and receive_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)
	return assets.Nums, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderByMonthNums(startTime, addr string, sid int) (decimal.Decimal, error) {

	var (
		build  = new(sql.SqlBuilder)
		assets ConfigNums
	)
	build.SqlAdd(" select sum(nums) as nums from orders").
		SqlWhere(" DATE_FORMAT(create_time,'%Y-%m') = DATE_FORMAT(?,'%Y-%m')", startTime, true).
		SqlAdd(" and receive_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)
	return assets.Nums, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindNoResult(sId int) ([]Orders, error) {
	var a []Orders
	db := model.DB().Where("order_result = 0 and service_id =?", sId).Find(&a)
	return a, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderByOId(oId int64) (*Orders, error) {
	a := new(Orders)
	db := model.DB().Where("id =?", oId).First(&a)
	return a, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderByUId(uid int64) ([]Orders, error) {
	var a []Orders
	sql := " select orders.* from orders left join service_audit_role on service_audit_role.sid = orders.service_id where service_audit_role.aid != 0 and service_audit_role.uid = ? and (orders.order_result = 0 or orders.order_result = 2) "
	db := model.DB().Raw(sql, uid).Scan(&a)
	return a, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderListByState(info *domain.SelectOrderInfo, id int64) ([]Orders, error) {

	var (
		order []Orders
		build = new(xkutils.StringBuilder)
	)
	if model.FilteredSQLInject(info.StartTime, info.EndTime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString("select orders.*, coin_info.name as coin_name, chain_info.name as chain_name,service.name as service_name,service.audit_type as audit_type,order_audit.audit_result as audit_result from orders ").
		AddString(" left join order_audit on orders.id = order_audit.order_id ").
		AddString(" left join coin_info on coin_info.id = orders.coin_id ").
		AddString(" left join chain_info on chain_info.id = coin_info.chain_id ").
		AddString(" left join service on service.id = orders.service_id ").
		StringBuild(" where order_audit.user_id = %d", id).
		AddString(" and orders.order_result = 0")
	if info.Contents != "" {
		oId := xkutils.StrToInt(info.Contents)
		if oId != 0 {
			build.StringBuild(" and orders.id = %d", oId)
		} else {
			build.StringBuild(" and service.name = %s", info.Contents)
		}
	}
	if info.CoinId != 0 {
		build.StringBuild(" and orders.coin_id = %d", info.CoinId)
	}
	if info.ServiceId != 0 {
		build.StringBuild(" and orders.service_id = %d", info.ServiceId)
	}
	if info.StartTime != "" {
		build.StringBuild(" and orders.create_time >= '%s'", info.StartTime)
	}
	if info.EndTime != "" {
		build.StringBuild(" and orders.create_time <= '%s'", info.EndTime)
	}
	db := model.DB().Raw(build.ToString()).Scan(&order)
	return order, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) FindOrderListByServices(info *domain.SelectOrderInfo, id int64) ([]Orders, error) {

	var (
		order []Orders
		build = new(xkutils.StringBuilder)
	)

	if info.Limit == 0 {
		info.Limit = 10
	}

	if model.FilteredSQLInject(info.StartTime, info.EndTime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}

	build.AddString("select orders.*, coin_info.name as coin_name, chain_info.name as chain_name,service.name as service_name,service.audit_type as audit_type,order_audit.audit_result as audit_result from orders ").
		AddString(" left join coin_info on coin_info.id = orders.coin_id ").
		AddString(" left join chain_info on chain_info.id = coin_info.chain_id ").
		AddString(" left join service on service.id = orders.service_id ").
		AddString(" left join order_audit on order_audit.order_id = orders.id")

	if id != 0 {
		build.StringBuild(" where order_audit.user_id = %d ", id)
	}

	if info.OrderResult != -1 {
		build.StringBuild(" and orders.order_result = %d", info.OrderResult)
	}
	if info.Contents != "" {
		oId := xkutils.StrToInt(info.Contents)
		if oId != 0 {
			build.StringBuild(" and orders.id = %d", oId)
		} else {
			build.StringBuild(" and service.name = '%s'", info.Contents)
		}
	}
	if info.CoinId != 0 {
		build.StringBuild(" and orders.coin_id = %d", info.CoinId)
	}
	if info.ServiceId != 0 {
		build.StringBuild(" and orders.service_id = %d", info.ServiceId)
	}
	if info.SerialNo != "" {
		build.StringBuild(" and orders.serial_no = '%s'", info.SerialNo)
	}
	if info.ChainName != "" {
		build.StringBuild(" and chain_info.name = '%s'", info.ChainName)
	}
	if info.StartTime != "" {
		build.StringBuild(" and orders.create_time >= '%s'", info.StartTime)
	}
	if info.EndTime != "" {
		build.StringBuild(" and orders.create_time <= '%s'", info.EndTime)
	}
	build.StringBuild(" order by orders.order_result asc limit %d,%d", info.Offset, info.Limit)
	db := model.DB().Raw(build.ToString()).Scan(&order)
	return order, model.ModelError(db, global.MsgWarnModelNil)
}

func (o *Orders) CountOrderListByServices(info *domain.SelectOrderInfo, id int64) (int, error) {
	var (
		count int64
		build = new(xkutils.StringBuilder)
	)
	if model.FilteredSQLInject(info.StartTime, info.EndTime) {
		return 0, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString("select count(1) from orders ").
		AddString(" left join coin_info on coin_info.id = orders.coin_id").
		AddString(" left join chain_info  on chain_info.id = coin_info.chain_id").
		AddString(" left join service on service.id = orders.service_id ").
		AddString(" left join order_audit on order_audit.order_id = orders.id")
	if id != 0 {
		build.StringBuild(" where order_audit.user_id = %d ", id)
	}
	if info.OrderResult != -1 {
		build.StringBuild(" and orders.order_result = %d", info.OrderResult)
	}
	if info.Contents != "" {
		oId := xkutils.StrToInt(info.Contents)
		if oId != 0 {
			build.StringBuild(" and orders.id = %d", oId)
		} else {
			build.StringBuild(" and service.name = '%s'", info.Contents)
		}
	}
	if info.CoinId != 0 {
		build.StringBuild(" and orders.coin_id = %d", info.CoinId)
	}
	if info.ServiceId != 0 {
		build.StringBuild(" and orders.service_id = %d", info.ServiceId)
	}
	if info.SerialNo != "" {
		build.StringBuild(" and orders.serial_no = '%s'", info.SerialNo)
	}
	if info.ChainName != "" {
		build.StringBuild(" and chain_info.name = '%s'", info.ChainName)
	}
	if info.StartTime != "" {
		build.StringBuild(" and orders.create_time >= '%s'", info.StartTime)
	}
	if info.EndTime != "" {
		build.StringBuild(" and orders.create_time <= '%s'", info.EndTime)
	}
	db := model.DB().Raw(build.ToString()).Count(&count)
	return int(count), model.ModelError(db, global.MsgWarnModelNil)
}
