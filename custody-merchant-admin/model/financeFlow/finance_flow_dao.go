package financeFlow

import (
	"custody-merchant-admin/module/log"
)

func (e *Entity) InsertNewItem() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavefinanceFlowInfo error: %v", err)
	}
	return
}

func (e *Entity) FindItemById(id int) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ?", id).Find(e).Error
	if err != nil {
		log.Errorf("SavefinanceFlowInfo error: %v", err)
	}
	return
}

func (e *Entity) DeleteItemById(id int) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ?", id).Delete(e).Error
	if err != nil {
		log.Errorf("SavefinanceFlowInfo error: %v", err)
	}
	return
}

func (e *Entity) FindItemByOrderId(orderId string) (err error) {
	err = e.Db.Table(e.TableName()).Where("order_id = ?", orderId).Find(e).Error
	if err != nil {
		log.Errorf("SavefinanceFlowInfo error: %v", err)
	}
	return
}
