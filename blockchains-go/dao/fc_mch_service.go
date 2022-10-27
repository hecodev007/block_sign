package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

//查询商户有效的数据
func FcMchServiceFindsValid(mchId int) ([]*entity.FcMchService, error) {
	results := make([]*entity.FcMchService, 0)
	err := db.Conn.Where("mch_id = ? and status = 0", mchId).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcMchServiceGetByMchId(mchId int) (*entity.FcMchService, error) {
	result := &entity.FcMchService{}
	has, err := db.Conn.Where("mch_id = ? and status = 0", mchId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcMchServiceFinds() ([]*entity.FcMchService, error) {
	results := make([]*entity.FcMchService, 0)
	err := db.Conn.Where("status = 0").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
