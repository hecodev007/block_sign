package dao

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func FcTransfersApplyCoinAddressCreate(o *entity.FcTransfersApplyCoinAddress) (int64, error) {
	return db.Conn.InsertOne(o)
}

// 根据applyid获取相关信息
// Deprecated: 单个查询是可以，但是有些状况是多个返回.
func GetApplyAddressByApplyCoinId(id int64, typed string) (*entity.FcTransfersApplyCoinAddress, error) {
	result := new(entity.FcTransfersApplyCoinAddress)
	if has, err := db.Conn.Where("apply_id = ? and address_flag = ?", id, typed).Get(result); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf("don't find for any %T", result)
	}
	return result, nil
}

//查询某个订单的地址（一般是只有接收地址，否则需要在create阶段插入找零地址和出站个地址）
//type :comment('出账地址,找零地址,接收地址') ENUM('change','from','to')
func FcTransfersApplyCoinAddressFindAddr(applyId int, addrType string) ([]string, error) {
	results := make([]string, 0)
	err := db.Conn.Table("fc_transfers_apply_coin_address").Cols("address").
		Where("apply_id = ? and address_flag = ?", applyId, addrType).
		Find(&results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("don't find any address")
	}
	return results, nil

	return results, nil
}

//查询某个订单的地址（一般是只有接收地址，否则需要在create阶段插入找零地址和出站个地址）
//type :comment('出账地址,找零地址,接收地址') ENUM('change','from','to')
func FcTransfersApplyCoinAddressFindAddrInfo(applyId int, addrType string) ([]*entity.FcTransfersApplyCoinAddress, error) {
	results := make([]*entity.FcTransfersApplyCoinAddress, 0)
	err := db.Conn.Where("apply_id = ? and address_flag = ?", applyId, addrType).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//根据订单，币种，获取出账地址
func FcGenerateAddressListFindToAddr(mchId int, coinName string, applyId int) ([]string, error) {
	results := make([]string, 0)
	err := db.Conn.Table("fc_transfers_apply_coin_address").Cols("address").
		Where("platform_id = ? and coin_name = ? and is_change = 1", mchId, coinName).
		Find(&results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("don't find any address")
	}
	return results, nil
}

func DeleteTransfersApplyCoinAddressByApplyId(applyId int64) error {
	result := new(entity.FcTransfersApplyCoinAddress)
	_, err := db.Conn.Where("apply_id = ?", applyId).Delete(result)
	if err != nil {
		return err
	}
	return nil
}
