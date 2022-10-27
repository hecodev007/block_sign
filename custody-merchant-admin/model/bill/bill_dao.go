package bill

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/util/sql"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

func (b *BillDetail) CreateBillDetail(db *gorm.DB, info *domain.BillInfo) (string, error) {

	bill := &BillDetail{
		CoinId:       info.CoinId,
		ChainId:      info.ChainId,
		ServiceId:    info.ServiceId,
		SerialNo:     info.SerialNo,
		MerchantId:   info.MerchantId,
		Phone:        info.Phone,
		Nums:         info.Nums,
		Fee:          info.Fee,
		BurnFee:      info.BurnFee,
		DestroyFee:   info.DestroyFee,
		RealNums:     info.RealNums,
		TxType:       info.TxType,
		BillStatus:   info.BillStatus,
		TxFromAddr:   info.TxFromAddr,
		TxToAddr:     info.TxToAddr,
		TxId:         info.TxId,
		FromId:       info.FromId,
		ToId:         info.ToId,
		Remark:       info.Remark,
		Memo:         info.Memo,
		CreateByUser: info.CreateByUser,
		CreatedAt:    time.Now(),
		TxTime:       time.Now(),
	}

	db.Model(&BillDetail{}).
		Omit("real_nums",
			"destroy_fee",
			"burn_fee",
			"destroy_fee",
			"updated_at",
			"deleted_at",
			"audit_time",
			"confirm_time",
			"service_name",
			"coin_name",
			"chain_name").
		Create(bill)

	err := model.ModelError(db, global.MsgWarnModelAdd)
	if err != nil {
		return "", err
	}
	return bill.SerialNo, nil
}

func (b *BillDetail) InsertBillDetail() (string, error) {

	db := model.DB().Begin()
	db.Model(&BillDetail{}).
		Omit("updated_at",
			"deleted_at",
			"audit_time",
			"confirm_time",
			"service_name",
			"coin_name",
			"chain_name").
		Create(b)
	err := model.ModelError(db, global.MsgWarnModelAdd)
	if err != nil {
		db.Rollback()
		return "", err
	}
	db.Commit()
	return b.SerialNo, nil
}

func (b *BillDetail) UpdateBillDetail(info *domain.BillInfo) (int, error) {
	if model.FilteredSQLInject(info.SerialNo) {
		return 0, errors.New(global.MsgWarnSqlInject)
	}
	mp := map[string]interface{}{"bill_status": info.BillStatus}
	db := model.DB().Model(&BillDetail{}).Where("serial_no=?", info.SerialNo).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (b *BillDetail) UpdateBill(txId string, mp map[string]interface{}) (int64, error) {

	db := model.DB().Model(&BillDetail{}).Where("tx_id = ?", txId).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return 0, err
	}
	return db.RowsAffected, nil
}

func (b *BillDetail) UpdateBillBySerialNo(serialNo string, mp map[string]interface{}) error {

	db := model.DB().Model(&BillDetail{}).Where("serial_no = ?", serialNo).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return err
	}
	return nil
}

func (b *BillDetail) GetBillBySerialNo(serialNo string) (*BillDetail, error) {

	bill := new(BillDetail)
	db := model.DB().Model(&BillDetail{}).Where("serial_no = ?", serialNo).First(bill)
	return bill, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) GetBillByTxId(txId string) (*BillDetail, error) {

	bill := new(BillDetail)
	db := model.DB().Model(&BillDetail{}).Where("tx_id = ?", txId).First(bill)
	return bill, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) GetBillByTxIdState(txId string, state int) (*BillDetail, error) {

	bill := new(BillDetail)
	db := model.DB().Model(&BillDetail{}).Where("tx_id = ? and tx_type= ? ", txId, state).First(bill)
	return bill, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) FindBillByStatus(id int64, state int) ([]BillNums, error) {
	var (
		build = new(sql.SqlBuilder)
		bn    = []BillNums{}
	)
	build.SqlAdd(" select sum(bill_detail.nums) as nums,bill_detail.coin_id as coin_id from bill_detail ").
		SqlAdd(" where (select count(1) from user_service where user_service.uid = ? and user_service.sid = bill_detail.service_id limit 1) > 0 ").
		SqlAdd(" and bill_detail.bill_status = ? and bill_detail.state = 0 group by bill_detail.coin_id")
	db := model.DB().Raw(build.ToSqlString(), id, state).Scan(&bn)
	return bn, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) FindBillByTimeNums(startTime, endTime, addr string, sid int) (*WithdrawalOrderInfo, error) {
	var (
		build  = new(sql.SqlBuilder)
		assets WithdrawalOrderInfo
	)
	if model.FilteredSQLInject(startTime, endTime, addr) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	build.SqlAdd(" select sum(nums) as nums,count(id) as counts from bill_detail").
		SqlWhere("  create_time >= ?", startTime, true).
		SqlWhereVars(" and create_time <= ? ", endTime, true).
		SqlAdd(" and tx_to_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)

	return &assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) FindBillByWeekNums(startTime, addr string, sid int) (decimal.Decimal, error) {
	var (
		build  = new(sql.SqlBuilder)
		assets ConfigNums
	)
	build.SqlAdd(" select sum(nums) as nums from bill_detail").
		SqlEnd(" where DATE_FORMAT(create_time,'%Y-%u') = DATE_FORMAT(?,'%Y-%u')", startTime, true).
		SqlAdd(" and tx_to_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)
	return assets.Nums, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) FindBillByDayNums(startTime, addr string, sid int) (decimal.Decimal, error) {
	var (
		build  = new(sql.SqlBuilder)
		assets ConfigNums
	)
	build.SqlAdd(" select sum(nums) as nums from bill_detail").
		SqlWhere(" DATE_FORMAT(create_time,'%Y-%m-%d') = DATE_FORMAT(?,'%Y-%m-%d')", startTime, true).
		SqlAdd(" and tx_to_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)
	return assets.Nums, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) FindBillByMonthNums(startTime, addr string, sid int) (decimal.Decimal, error) {

	var (
		build  = new(sql.SqlBuilder)
		assets ConfigNums
	)
	build.SqlAdd(" select sum(nums) as nums from bill_detail").
		SqlWhere(" DATE_FORMAT(create_time,'%Y-%m') = DATE_FORMAT(?,'%Y-%m')", startTime, true).
		SqlAdd(" and tx_to_addr = ? and service_id = ? group by create_time")
	db := model.DB().Raw(build.ToSqlString(), addr, sid).Scan(&assets)
	return assets.Nums, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) FindBillDetailList(info *domain.BillSelect) ([]BillLists, error) {

	var (
		build = new(xkutils.StringBuilder)
		blst  []BillLists
	)
	if model.FilteredSQLInject(info.Address, info.TxStartTime, info.TxEndTime, info.Phone) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString(" select bill_detail.*,o.order_result as order_result, " +
		"chain.id as chain_id,chain.name as chain_name, " +
		"s.name as service_name,c.name as coin_name from bill_detail ").
		AddString(" left join orders as o on o.serial_no = bill_detail.serial_no ").
		AddString(" left join service as s on s.id = bill_detail.service_id ").
		AddString(" left join coin_info as c on c.id = bill_detail.coin_id ").
		AddString(" left join chain_info as chain on chain.id = c.chain_id ").
		AddString(" where bill_detail.state = 0 ")
	if info.MerchantId != 0 {
		build.StringBuild(" and bill_detail.merchant_id = %d", info.MerchantId)
	}
	if info.Phone != "" {
		build.StringBuild(" and bill_detail.phone = '%s'", info.Phone)
	}
	if info.CoinId != 0 {
		build.StringBuild(" and bill_detail.coin_id = %d", info.CoinId)
	}
	if info.ServiceId != 0 {
		build.StringBuild(" and bill_detail.service_id = %d", info.ServiceId)
	}
	if info.TxType != 0 {
		build.StringBuild(" and bill_detail.tx_type = %d", info.TxType)
	}
	if info.BillStatus > 0 {
		build.StringBuild(" and bill_detail.bill_status = %d", info.BillStatus)
	}
	if info.Address != "" {
		build.StringBuild(" and (bill_detail.tx_to_addr = '%s' or bill_detail.tx_from_addr = '%s'  or bill_detail.tx_id = '%s') ", info.Address, info.Address, info.Address)
	}
	if info.TxStartTime != "" {
		build.StringBuild(" and bill_detail.tx_time >= '%s 00:00:00'", info.TxStartTime)
	}
	if info.TxEndTime != "" {
		build.StringBuild(" and bill_detail.tx_time <= '%s 23:59:59.9999'", info.TxEndTime)
	}
	if info.ConfirmStartTime != "" {
		build.StringBuild(" and bill_detail.confirm_time >= '%s 00:00:00'", info.ConfirmStartTime)
	}
	if info.ConfirmEndTime != "" {
		build.StringBuild(" and bill_detail.confirm_time <= '%s 23:59:59.9999'", info.ConfirmEndTime)
	}

	build.StringBuild(" order by bill_detail.id desc  limit %d,%d", info.Offset, info.Limit)
	sql := build.ToString()
	db := model.DB().Raw(sql).Scan(&blst)
	return blst, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) CountBillDetailList(info *domain.BillSelect) (int64, error) {
	var (
		build = new(xkutils.StringBuilder)
		count int64
	)
	build.StringBuild(" select count(1) from bill_detail  where state = 0 ")
	if info.MerchantId != 0 {
		build.StringBuild(" and bill_detail.merchant_id = %d", info.MerchantId)
	}
	if info.Phone != "" {
		build.StringBuild(" and bill_detail.phone = '%s'", info.Phone)
	}
	if info.CoinId != 0 {
		build.StringBuild(" and coin_id = %d", info.CoinId)
	}
	if info.ServiceId != 0 {
		build.StringBuild(" and service_id = %d", info.ServiceId)
	}
	if info.TxType != 0 {
		build.StringBuild(" and tx_type = %d", info.TxType)
	}
	if info.BillStatus > 0 {
		build.StringBuild(" and bill_status = %d", info.BillStatus)
	}
	if info.Address != "" {
		build.StringBuild(" and (tx_to_addr = '%s' or tx_from_addr = '%s'  or tx_id = '%s')", info.Address, info.Address, info.Address)
	}

	if info.TxStartTime != "" {
		build.StringBuild(" and bill_detail.tx_time >= '%s 00:00:00'", info.TxStartTime)
	}
	if info.TxEndTime != "" {
		build.StringBuild(" and bill_detail.tx_time <= '%s 23:59:59.9999'", info.TxEndTime)
	}

	if info.ConfirmStartTime != "" {
		build.StringBuild(" and bill_detail.confirm_time >= '%s 00:00:00'", info.ConfirmStartTime)
	}
	if info.ConfirmEndTime != "" {
		build.StringBuild(" and bill_detail.confirm_time <= '%s 23:59:59.9999'", info.ConfirmEndTime)
	}
	db := model.DB().Raw(build.ToString()).Count(&count)
	return count, model.ModelError(db, global.MsgWarnModelNil)
}

func (b *BillDetail) FindBillDetailBySerialNo(serialNo string) (BillLists, error) {

	var (
		build = new(xkutils.StringBuilder)
		blst  BillLists
	)
	if model.FilteredSQLInject(serialNo) {
		return BillLists{}, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString(" select bill_detail.*,o.order_result as order_result, " +
		"chain.id as chain_id,chain.name as chain_name, " +
		"s.name as service_name,c.name as coin_name from bill_detail ").
		AddString(" left join orders as o on o.serial_no = bill_detail.serial_no ").
		AddString(" left join service as s on s.id = bill_detail.service_id ").
		AddString(" left join coin_info as c on c.id = bill_detail.coin_id ").
		AddString(" left join chain_info as chain on chain.id = c.chain_id ").
		AddString(" where bill_detail.state = 0 ")

	if serialNo != "" {
		build.StringBuild(" and bill_detail.serial_no = '%s'", serialNo)
	}
	build.StringBuild(" order by bill_detail.id desc  limit 1")
	sql := build.ToString()
	db := model.DB().Raw(sql).Scan(&blst)
	return blst, model.ModelError(db, global.MsgWarnModelNil)
}
