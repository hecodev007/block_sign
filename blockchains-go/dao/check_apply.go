package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func CheckApplyFindByApplyId(applyId int64) (*entity.CheckApply, error) {
	result := new(entity.CheckApply)
	if has, err := db.Conn2.Where("apply_id= ?", applyId).Get(result); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}
func CheckApplyDeleteByApplyId(applyId int64) error {
	result := new(entity.CheckApply)
	_, err := db.Conn2.Where("apply_id = ?", applyId).Delete(result)
	if err != nil {
		return err
	}
	return nil
}
