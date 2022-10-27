package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

//查询商户有效的数据
func FcApiListGetByID(id int) (*entity.FcApiList, error) {
	result := new(entity.FcApiList)
	has, err := db.Conn.Where("id = ? and status = 1", id).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}
