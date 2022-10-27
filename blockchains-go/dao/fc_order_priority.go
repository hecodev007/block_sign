package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"xorm.io/builder"
)

func FcOrderPriorityByApplyIds(applyIds []int) ([]entity.FcOrderPriority, error) {
	results := make([]entity.FcOrderPriority, 0)
	err := db.Conn.Where(builder.In("apply_id", applyIds)).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcOrderPriorityList() ([]entity.FcOrderPriority, error) {
	results := make([]entity.FcOrderPriority, 0)
	err := db.Conn.Where("status = ?", entity.OrderPriorityStatusProcessing).OrderBy("create_time ASC").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcOrderPriorityByChain(chain, coinCode string, mchId int) ([]entity.FcOrderPriority, error) {
	results := make([]entity.FcOrderPriority, 0)
	err := db.Conn.Where("chain_name = ? and coin_code = ? and mch_id = ? and status = ?", chain, coinCode, mchId, entity.OrderPriorityStatusProcessing).OrderBy("create_time ASC").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcOrderPriorityByApplyId(applyId int) (*entity.FcOrderPriority, error) {
	order := &entity.FcOrderPriority{}
	has, err := db.Conn.Where("apply_id = ?", applyId).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return order, nil
}

func FcOrderPriorityByOuterOrderNo(outerOrderNo string) (*entity.FcOrderPriority, error) {
	order := &entity.FcOrderPriority{}
	has, err := db.Conn.Where("outer_order_no = ?", outerOrderNo).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return order, nil
}

func UpdatePriorityCompletedIfExist(applyId int) {
	priority, _ := FcOrderPriorityByApplyId(applyId)
	if priority != nil {
		FcOrderPriorityUpdateStatusByApplyId(applyId, entity.OrderPriorityStatusCompleted)
	}
}

func DeletePriorityOrder(id int) error {
	order := &entity.FcOrderPriority{}
	_, err := db.Conn.Where("id = ?", id).Delete(order)
	if err != nil {
		return err
	}
	return nil
}

func FcOrderPriorityUpdateStatusByApplyId(applyId int, status entity.OrderPriorityStatus) (int64, error) {
	return db.Conn.Table("fc_order_priority").Where("apply_id = ?", applyId).Update(map[string]interface{}{"status": status})
}
