package chainBill

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"gorm.io/gorm"
)

func (e *Entity) CreateChainBill() error {
	db := model.DB().Begin()
	db.Model(&Entity{}).
		Omit("updated_at", "deleted_at").
		Create(e)
	if db.Error != nil {
		db.Rollback()
		return db.Error
	}
	db.Commit()
	return nil
}

func (e *Entity) FindChainBillBySerialNo(serialNo string) error {
	db := model.DB()
	db.Model(&Entity{}).Where("serial_no=?", serialNo).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindChainBillById(id int64) error {
	db := model.DB()
	db.Model(&Entity{}).Where("id=?", id).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) UpdatesChainBillBySerialNo(serialNo string, mp map[string]interface{}) error {
	db := model.DB().Begin()
	db.Model(&Entity{}).Where("serial_no =?", serialNo).Updates(mp)
	if db.Error != nil {
		db.Rollback()
		return db.Error
	}
	db.Commit()
	return nil
}

func (e *Entity) UpdatesChainBill(id int64, mp map[string]interface{}) error {
	db := model.DB().Begin()
	db.Model(&Entity{}).Where("id =?", id).Updates(mp)
	if db.Error != nil {
		db.Rollback()
		return db.Error
	}
	db.Commit()
	return nil
}

func (e *Entity) UpdatesChainBillCommit(db *gorm.DB, id int64, mp map[string]interface{}) error {
	db.Model(&Entity{}).Where("id =?", id).Updates(mp)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (e *Entity) FindBillLimit(limit int, chain, coin int) ([]Entity, error) {
	var list = []Entity{}
	db := model.DB().Table("chain_bill")
	db.Where(" up_chain_fee is not null and (tx_type = 1 or tx_type = 4) and chain_id=? and coin_id =?", chain, coin).
		Limit(limit).
		Order("id desc").
		Find(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindChainBillNoUpChain() ([]Entity, error) {
	var list = []Entity{}
	db := model.DB().Table("chain_bill")
	db.Where(" is_wallet_deal = 0 and (tx_type = 0 or tx_type = 3 ) ").Find(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindChainBillNoUpChainTxid(Txid string) (Entity, error) {
	var list = Entity{}
	db := model.DB().Table("chain_bill")
	db.Where(" is_wallet_deal = 0 and tx_id =? ", Txid).First(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindChainBillChainTxid(Txid string) (Entity, error) {
	var list = Entity{}
	db := model.DB().Table("chain_bill")
	db.Where(" tx_id =? ", Txid).First(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindChainBillChainSerialNo(serialNo string) (Entity, error) {
	var list = Entity{}
	db := model.DB().Table("chain_bill")
	db.Where(" serial_no =?", serialNo).First(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindChainBillNoUpChainSerialNo(serialNo string) (Entity, error) {
	var list = Entity{}
	db := model.DB().Table("chain_bill")
	db.Where(" is_wallet_deal = 0 and serial_no =? ", serialNo).First(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindList(chain *domain.ChainBillSelect) ([]ChainBillLists, int64, error) {

	clist := []ChainBillLists{}
	var count int64
	db := model.DB().Table("chain_bill").
		Select(" chain_bill.*," +
			"chain.id as chain_id, " +
			"chain.name as chain_name, " +
			"s.name as service_name, " +
			"c.name as coin_name, " +
			"user.is_test as is_test, " +
			"user.phone_code as phone_code ").
		Joins(" left join service as s on s.id = chain_bill.service_id").
		Joins(" left join coin_info as c on c.id = chain_bill.coin_id").
		Joins(" left join chain_info as chain on chain.id = c.chain_id").
		Joins(" left join user_info as user on user.id = chain_bill.merchant_id ").
		Where(" chain_bill.state = 0 ")
	if chain.Phone != "" {
		db.Where("chain_bill.phone = ? ", chain.Phone)
	}
	if chain.MerchantId != 0 {
		db.Where("chain_bill.merchant_id = ? ", chain.MerchantId)
	}
	if chain.AddressOrMemo != "" {
		db.Where(" chain_bill.tx_to_addr = ? or chain_bill.memo = ? ", chain.AddressOrMemo, chain.AddressOrMemo)
	}
	if chain.ConfirmStartTime != "" {
		db.Where("chain_bill.confirm_time >= ? ", chain.ConfirmStartTime)
	}
	if chain.ConfirmEndTime != "" {
		db.Where("chain_bill.confirm_time =< ?", chain.ConfirmEndTime)
	}
	if chain.StartTime != "" {
		db.Where("chain_bill.created_at >= ? ", chain.StartTime)
	}
	if chain.EndTime != "" {
		db.Where("chain_bill.created_at =< ?", chain.EndTime)
	}
	if chain.TxType != -1 {
		db.Where("chain_bill.tx_type = ?", chain.TxType)
	}
	if chain.IsReback != -1 {
		db.Where("chain_bill.is_reback = ? ", chain.IsReback)
	}
	if chain.Offset == 0 && chain.Limit > 10 {
		db.Offset(chain.Offset).Limit(chain.Limit).Find(&clist)
	} else {
		db.Offset(chain.Offset).Limit(chain.Limit).Find(&clist).Offset(-1).Limit(-1).Count(&count)
	}
	return clist, count, model.ModelError(db, global.MsgWarnModelNil)
}
