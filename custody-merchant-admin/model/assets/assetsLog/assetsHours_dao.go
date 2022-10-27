package assetsLog

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (u *AssetsHours) FindAssetsHours(startTime, endTime string, id int64) ([]AsTime, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(startTime, endTime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString("select service_id,coin_id,sum(freeze) as freeze, sum(nums) as nums,").
		AddString(" DATE_FORMAT(create_time,'%Y-%m-%d %T') as create_time from assets_hours ").
		AddString(" where (select count(1) from coin_info where coin_info.id = assets_hours.coin_id limit 1) > 0 and (select count(1) from service_audit_role where service_audit_role.sid = assets_hours.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0")
	if startTime == "" {
		startTime = strings.Split(time.Now().Local().Format(global.YyyyMmDdHhMmSs), ":")[0] + ":00:00"
	} else {
		startTime += " 00:00:00"
	}
	build.StringBuild(" and create_time >= '%s'", startTime).
		StringBuild(" and create_time <= date_sub('%s',interval -24 hour)", startTime)
	if endTime != "" {
		build.StringBuild(" and create_time <= '%s 23:59:59.9999'", endTime)
	}
	build.StringBuild(" group by create_time,service_id,coin_id order by create_time ")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *AssetsHours) GetAssetsHours(CoinId, ServiceId int, createTime string) (AssetsHours, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AssetsHours
		asset  AssetsHours
	)
	if model.FilteredSQLInject(createTime) {
		return asset, fmt.Errorf(global.MsgWarnSqlInject)
	}
	build.StringBuild("select assets_hours.* from assets_hours  where assets_hours.id > 0")
	if CoinId > 0 {
		build.StringBuild(" and assets_hours.coin_id = %d", CoinId)
	}
	if ServiceId > 0 {
		build.StringBuild(" and assets_hours.service_id = %d", ServiceId)
	}
	if createTime != "" {
		build.StringBuild(" and assets_hours.create_time = '%s'", createTime)
	}
	build.AddString(" order by create_time desc")
	build.AddString(" limit 0,1 ")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	err := model.ModelError(db, global.MsgWarnModelNil)
	if len(assets) > 0 {
		asset = assets[0]
	}
	return asset, err
}
