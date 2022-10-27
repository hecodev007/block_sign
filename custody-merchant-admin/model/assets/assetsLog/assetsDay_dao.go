package assetsLog

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"fmt"
	"time"
)

// FindAssetsDay
// 获取年-月-日的资产
func (u *AssetsDay) FindAssetsDay(startTime, endTime string, id int64) ([]AsTime, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(endTime, startTime) {
		return assets, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString("select coin_id,sum(nums) as nums,sum(freeze) as freeze,DATE_FORMAT(create_time,'%Y-%m-%d') as create_time from assets_day").
		StringBuild(" where (select count(1) from coin_info where coin_info.id = assets_day.coin_id limit 1) > 0 and  ( select count(1) from service_audit_role where service_audit_role.sid = assets_day.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0")
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	build.StringBuild(" and create_time >= '%s 00:00:00'", startTime).
		StringBuild(" and create_time <= date_sub('%s 00:00:00',interval -12 day)", startTime)
	if endTime != "" {
		build.StringBuild(" and create_time <= '%s 23:59:59.9999'", endTime)
	}
	build.StringBuild(" group by create_time,coin_id order by create_time")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

// FindAssetsDayByStart
// 获取年-月-日的资产
func (u *AssetsDay) FindAssetsDayByStart(startTime string, id int64) ([]AsTime, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(startTime) {
		return assets, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString("select coin_id,sum(nums) as nums,sum(freeze) as freeze,DATE_FORMAT(create_time,'%Y-%m-%d') as create_time from assets_day").
		StringBuild(" where (select count(1) from coin_info where coin_info.id = assets_day.coin_id limit 1) > 0 and ( select count(1) from service_audit_role where service_audit_role.sid = assets_day.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0")
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	build.StringBuild(" and create_time < date_sub('%s 00:00:00', interval 0 day)", startTime).
		StringBuild(" and create_time >= date_sub('%s 00:00:00',interval 1 day)", startTime).
		StringBuild(" group by create_time,coin_id order by create_time desc")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

// FindAssetsWeek
// 获取年-周的资产
func (u *AssetsDay) FindAssetsWeek(startTime, endTime string, id int64) ([]AsTime, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(endTime, startTime) {
		return assets, fmt.Errorf(global.MsgWarnSqlInject)
	}
	build.AddString("select coin_id,sum(nums) as nums,sum(freeze) as freeze,DATE_FORMAT(create_time,'%Y-%u') as create_time from assets_day").
		StringBuild(" where ( select count(1) from service_audit_role where service_audit_role.sid = assets_day.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0")
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	build.StringBuild(" and create_time >= '%s 00:00:00'", startTime).
		StringBuild(" and create_time <= date_sub('%s 00:00:00',interval -3 month)", startTime)
	if endTime != "" {
		build.StringBuild(" and create_time <= '%s 23:59:59.9999'", endTime)
	}
	build.StringBuild(" group by create_time,coin_id order by create_time asc")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

// FindAssetsWeekStart
// 获取年-周的资产
func (u *AssetsDay) FindAssetsWeekStart(startTime string, id int64) ([]AsTime, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(startTime) {
		return assets, fmt.Errorf(global.MsgWarnSqlInject)
	}
	build.AddString("select coin_id,sum(nums) as nums,sum(freeze) as freeze,DATE_FORMAT(create_time,'%Y-%u') as create_time from assets_day").
		StringBuild(" where ( select count(1) from service_audit_role where service_audit_role.sid = assets_day.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0")
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	build.StringBuild(" and create_time < '%s 00:00:00'", startTime).
		StringBuild(" and create_time >= date_sub('%s 00:00:00',interval -3 month)", startTime).
		StringBuild(" group by create_time,coin_id order by create_time desc")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *AssetsDay) GetAssetsDay(CoinId, ServiceId int, createTime string) (AssetsDay, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AssetsDay
		asset  AssetsDay
	)
	if model.FilteredSQLInject(createTime) {
		return asset, fmt.Errorf(global.MsgWarnSqlInject)
	}
	build.StringBuild("select * from assets_day  where id > 0")
	if CoinId > 0 {
		build.StringBuild(" and coin_id = %d", CoinId)
	}
	if ServiceId > 0 {
		build.StringBuild(" and service_id = %d", ServiceId)
	}
	if createTime != "" {
		build.StringBuild(" and create_time = '%s'", createTime)
	}
	build.StringBuild(" order by create_time desc")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	if len(assets) > 0 {
		asset = assets[0]
	}
	return asset, model.ModelError(db, global.MsgWarnModelNil)
}
