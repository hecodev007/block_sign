package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

//根据appid coinid获取余额信息
func FcMchAmountGetByACId(appId, coinId int) (*entity.FcMchAmount, error) {
	result := new(entity.FcMchAmount)
	has, err := db.Conn.Where("app_id = ? and coin_id = ?", appId, coinId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

//func FcMchAmountGetByACId(appId, coinId int, coinName string) (*entity.FcMchAmount, error) {
//	result := new(entity.FcMchAmount)
//	has, err := db.Conn.Where("app_id = ? and coin_id = ? and coin_type = ?", appId, coinId, coinName).Get(result)
//	if err != nil {
//		return nil, err
//	}
//	if !has {
//		return nil, errors.New("Not Fount!")
//	}
//	return result, nil
//}

//获取余额信息
func FcMchAmountGetInfo(appId int, coinName string) (*entity.FcMchAmount, error) {
	result := new(entity.FcMchAmount)
	has, err := db.Conn.Where("app_id = ? and coin_type = ?", appId, coinName).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}
