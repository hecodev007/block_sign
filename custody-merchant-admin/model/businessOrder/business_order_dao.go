package businessOrder

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/module/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (e *Entity) InsertNewItem() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SaveBusinessOrderInfo error: %v", err)
	}
	return
}

func (e *Entity) DeleteBusinessOrderItem(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Delete(map[string]interface{}{"id": pId}).Error
	if err != nil {
		log.Errorf("DelBusinessOrderInfo error: %v", err)
	}
	return
}

func (e *Entity) FindBusinessOrderItemByOrderId(oId string) (err error) {
	err = e.Db.Table(e.TableName()).Where("order_id = ? ", oId).Find(e).Error
	if err != nil {
		log.Errorf("FindBusinessOrderOneInfo error: %v", err)
	}
	if e.OrderId == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) UpdateBusinessOrderItem() (err error) {
	err = e.Db.Table(e.TableName()).Updates(e).Error
	if err != nil {
		log.Errorf("UpdateBusinessOrderInfo error: %v", err)
	}
	return
}

func (e *Entity) FindBusinessOrderItemById(pId int64) (item Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? ", pId).Find(&item).Error
	if err != nil {
		log.Errorf("FindBusinessOrderOneInfo error: %v", err)
	}
	if item.OrderId == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindLatestBusinessOrderItemByBid(bId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("business_id = ? and deduct_state = 'success'", bId).Order("id desc").First(e).Error
	if err != nil {
		log.Errorf("FindBusinessOrderOneInfo error: %v", err)
	}
	if e.OrderId == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindBusinessOrderListByReq(req domain.OrderReqInfo) (list []Entity, total int64, err error) {
	selectSql := e.Db.Table(e.TableName())
	if req.AccountId != 0 {
		selectSql = selectSql.Where("account_id = ?", req.AccountId)
	}
	if req.OrderId != "" {
		selectSql = selectSql.Where("order_id = ?", req.OrderId)
	}
	if req.ContactStr != "" {
		//通过手机/邮箱获取用户ID
		mInfo := merchant.NewEntity()
		mItem, _ := mInfo.FindMerchantItemByContactStr(req.ContactStr)
		accountId := mItem.Id
		selectSql = selectSql.Where("account_id = ?", accountId)
	}
	selectSql.Count(&total)
	err = selectSql.Limit(req.Limit).Offset(req.Offset).Find(&list).Error
	if err != nil {
		log.Errorf("FindBusinessOrderList error: %v", err)
	}
	return
}

//FindBusinessOrderInfoListByReq 查询订单详细信息
func (e *Entity) FindBusinessOrderInfoListByReq(req *domain.OrderReqInfo) (list []Item, total int64, err error) {

	//mItem := merchant.Entity{}
	//uTableName := mItem.TableName()
	//bItem := business.Entity{}
	//bTableName := bItem.TableName()
	//bpItem := businessPackage.Entity{}
	//bpTableName := bpItem.TableName()
	selectSql := e.Db.Table(e.TableName()).Select("admin_service_order.*,admin_service_order.admin_remark as remark ," +
		"u.name,u.is_test as account_status,u.id as account_id ,b.phone,u.name,b.email," +
		"b.id as business_id,b.coin as coin ,b.sub_coin as sub_coin,b.name as business_name," +
		"bp.type_name as type_name,bp.model_name as model_name,bp.deduct_coin as deduct_coin," +
		"bp.deploy_fee as deploy_fee,bp.custody_fee as custody_fee ,bp.deposit_fee as deposit_fee,bp.cover_fee as cover_fee ")

	selectSql = selectSql.Joins("left join user_info u ON u.id = admin_service_order.account_id")                // 用户表
	selectSql = selectSql.Joins("left join service b ON b.id = admin_service_order.business_id")                 //业务线表
	selectSql = selectSql.Joins("left join service_combo bp ON bp.id = admin_service_order.business_package_id") //业务线套餐表
	if req.AccountId != 0 {
		selectSql = selectSql.Where("admin_service_order.account_id = ?", req.AccountId)
	}
	if req.BusinessId != 0 {
		selectSql = selectSql.Where("admin_service_order.business_id = ?", req.BusinessId)
	}
	if req.OrderId != "" {
		selectSql = selectSql.Where("admin_service_order.order_id = ?", req.OrderId)
	}
	if req.ContactStr != "" {
		//通过手机/邮箱获取用户ID
		mInfo := merchant.NewEntity()
		mItem, _ := mInfo.FindMerchantItemByContactStr(req.ContactStr)
		accountId := mItem.Id
		selectSql = selectSql.Where("admin_service_order.account_id = ?", accountId)
	}
	selectSql.Count(&total)
	err = selectSql.Limit(req.Limit).Offset(req.Offset).Order("id desc").Find(&list).Error
	if err != nil {
		log.Errorf("FindBusinessOrderList error: %v", err)
	}
	return
}

func (e *Entity) SumNumsBusinessOrder(coinId int, bid, bpid, aid int64) (decimal.Decimal, error) {
	selectSql := model.DB().Raw("select order_coin_id,sum(profit_number) as profit_number from admin_service_order "+
		" where deduct_state='success' and order_coin_id = ? and business_id=? and business_package_id=? and account_id=? GROUP BY order_coin_id LIMIT 1",
		coinId, bid, bpid, aid).Scan(e)
	return e.ProfitNumber, model.ModelError(selectSql, global.MsgWarnModelNil)
}

func (e *Entity) FindLatestPackageType(bId int) (err error) {
	err = e.Db.Table("admin_service_order").Where("business_id = ?", bId).Order("id desc").First(e).Error
	//err = selectSql.Find(&list).Error
	if err != nil {
		log.Errorf("FindPackageList error: %v", err)
	}
	return
}
