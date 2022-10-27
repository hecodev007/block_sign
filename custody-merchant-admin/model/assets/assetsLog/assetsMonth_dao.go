package assetsLog

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

func (u *AssetsMonth) FindAssetsMonth(startTime, endTime string, id int64) ([]AsTime, error) {
	var (
		//build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(startTime, endTime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}

	//build.StringBuild("select service_id,coin_id,sum(nums) as nums,sum(assets_month.freeze) as freeze,").
	//	AddString("DATE_FORMAT(create_time,'%Y-%m') as create_time from assets_month").
	//	StringBuild(" where ( select count(1) from service_audit_role where service_audit_role.sid = assets_month.service_id ")
	//if id != 0 {
	//	build.StringBuild(" and service_audit_role.uid = %d", id)
	//}
	//build.AddString(" limit 1) > 0")
	//if startTime == "" {
	//	startTime = time.Now().Local().Format(global.YyyyMmDd)
	//}
	//build.StringBuild(" and create_time >= '%s 00:00:00'", startTime).
	//	StringBuild(" and create_time <= date_sub('%s 00:00:00',interval -12 month)", startTime)
	//
	//if endTime != "" {
	//	build.StringBuild(" and create_time <= '%s 23:59:59.9999'", endTime)
	//}
	//
	//build.StringBuild(" group by create_time,service_id,coin_id order by create_time asc ")
	//

	db := model.DB().Table("assets_month")
	db.Select("service_id,coin_id,sum(nums) as nums,sum(assets_month.freeze) as freeze,DATE_FORMAT(create_time,'%Y-%m') as create_time")
	db.Where(" (select count(1) from coin_info where coin_info.id = assets_month.coin_id limit 1) > 0 ")
	if id != 0 {
		db.Where("( select count(1) from service_audit_role where service_audit_role.sid = assets_month.service_id and service_audit_role.uid =? limit 1) > 0", id)
	} else {
		db.Where("( select count(1) from service_audit_role where service_audit_role.sid = assets_month.service_id  limit 1) > 0")
	}
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	startTime = fmt.Sprintf("%s 00:00:00", startTime)
	db.Where(" create_time >=? ", startTime).Where(" create_time <= date_sub(?,interval -12 month)", startTime)
	if endTime != "" {
		endTime = fmt.Sprintf("%s 23:59:59.9999", endTime)
		db.Where("create_time <= ?", endTime)
	}
	db.Group("create_time,service_id,coin_id").Order("create_time asc ").Find(&assets).Debug()

	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *AssetsMonth) FindAssetsMonthStart(startTime string, id int64) ([]AsTime, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(startTime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	build.StringBuild("select service_id,coin_id,sum(nums) as nums,sum(assets_month.freeze) as freeze,").
		AddString("DATE_FORMAT(create_time,'%Y-%m') as create_time from assets_month").
		StringBuild(" where (select count(1) from coin_info where coin_info.id = assets_month.coin_id limit 1) > 0 and ( select count(1) from service_audit_role where service_audit_role.sid = assets_month.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0")
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	build.StringBuild(" and create_time < '%s 00:00:00'", startTime).
		StringBuild(" and create_time <= date_sub('%s 00:00:00',interval 1 month)", startTime).
		AddString(" group by create_time,service_id,coin_id order by create_time asc ")

	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *AssetsMonth) FindAssetsYear(startTime, endTime string, id int64) ([]AsTime, error) {

	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	if model.FilteredSQLInject(startTime, endTime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	build.AddString("select service_id,coin_id,sum(nums) as nums,sum(assets_month.freeze) as freeze,DATE_FORMAT(create_time,'%Y') as create_time from assets_month").
		AddString(" where (select count(1) from coin_info where coin_info.id = assets_month.coin_id limit 1) > 0 and ( select count(1) from service_audit_role where service_audit_role.sid = assets_month.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0").
		StringBuild(" and create_time >= '%s 00:00:00'", startTime).
		StringBuild(" and create_time <= date_sub('%s 00:00:00',interval -12 year)", startTime)
	if endTime != "" {
		build.StringBuild(" and create_time <= '%s 23:59:59.9999'", endTime)
	}
	build.StringBuild(" group by create_time,service_id,coin_id order by create_time asc ")

	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *AssetsMonth) FindAssetsYearStart(startTime string, id int64) ([]AsTime, error) {

	var (
		build  = new(xkutils.StringBuilder)
		assets []AsTime
	)
	if model.FilteredSQLInject(startTime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	if startTime == "" {
		startTime = time.Now().Local().Format(global.YyyyMmDd)
	}
	build.AddString("select service_id,coin_id,sum(nums) as nums,sum(assets_month.freeze) as freeze,DATE_FORMAT(create_time,'%Y') as create_time from assets_month").
		AddString(" where (select count(1) from coin_info where coin_info.id = assets_month.coin_id limit 1) > 0 and ( select count(1) from service_audit_role where service_audit_role.sid = assets_month.service_id ")
	if id != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", id)
	}
	build.AddString(" limit 1) > 0").
		StringBuild(" and create_time < '%s 00:00:00'", startTime).
		StringBuild(" and create_time <= date_sub('%s 00:00:00',interval 1 year)", startTime).
		AddString(" group by create_time,service_id,coin_id order by create_time asc ")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	return assets, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *AssetsMonth) GetAssetsMonth(coinId, serviceId int, createTime string) (AssetsMonth, error) {
	var (
		build  = new(xkutils.StringBuilder)
		assets []AssetsMonth
	)
	if model.FilteredSQLInject(createTime) {
		return AssetsMonth{}, fmt.Errorf(global.MsgWarnSqlInject)
	}
	build.StringBuild("select assets_month.* from assets_month  where assets_month.id > 0")
	if coinId > 0 {
		build.StringBuild(" and assets_month.coin_id = %d", coinId)
	}
	if serviceId > 0 {
		build.StringBuild(" and assets_month.service_id = %d", serviceId)
	}
	if createTime != "" {
		build.StringBuild(" and assets_month.create_time = '%s'", createTime)
	}
	build.StringBuild(" order by create_time desc")
	db := model.DB().Raw(build.ToString()).Scan(&assets)
	err := model.ModelError(db, global.MsgWarnModelNil)
	if err != nil {
		return AssetsMonth{}, err
	}
	if len(assets) > 0 {
		return assets[0], nil
	}
	return AssetsMonth{}, err
}
