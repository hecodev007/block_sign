package dao

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"time"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func FcGenerateAddressListFindColdAddrs(mchId int, coinName string) ([]string, error) {
	return entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   coinName,
	})
}

//根据币种获取出账冷地址
func FcGenerateAddressListFindAddresses(typed, status, mchId int, coinName string) ([]string, error) {
	results := make([]string, 0)
	err := db.Conn.Table("fc_generate_address_list").Cols("address").
		Where("type = ? and status = ? and platform_id = ? and coin_name = ?", typed, status, mchId, coinName).
		Find(&results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("don't find any address")
	}
	return results, nil
}

//根据币种获取地址
func FcGenerateAddressListFindAddressesData(typed, status, mchId int, coinName string) ([]*entity.FcGenerateAddressList, error) {
	results := make([]*entity.FcGenerateAddressList, 0)
	err := db.Conn.Where("type = ? and status = ? and platform_id = ? and coin_name = ?", typed, status, mchId, coinName).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//根据币种获取找零地址
func FcGenerateAddressListFindChangeAddr(mchId int, coinName string) ([]string, error) {
	results := make([]string, 0)
	err := db.Conn.Table("fc_generate_address_list").Cols("address").
		Where("platform_id = ? and coin_name = ? and type = 1 and is_change = 1 and status = 2", mchId, coinName).
		Find(&results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("don't find any address")
	}
	return results, nil
}

func FcGenerateAddressListGetMchAddress(typed, status, mchId int, coinName string) (string, error) {
	var result string
	has, err := db.Conn.Table("fc_generate_address_list").Cols("address").
		Where("type = ? and status = ? and platform_id = ? and coin_name = ?", typed, status, mchId, coinName).
		Get(&result)
	if err != nil {
		return "", err
	}
	if !has {
		return "", fmt.Errorf("don't find any address")
	}
	return result, nil
}

//写入商户使用地址表
func WriteMchAddrs(bas []*entity.FcGenerateBeforeAddressList, tx *xorm.Session) ([]*entity.FcGenerateAddressList, error) {
	var (
		as           []*entity.FcGenerateAddressList
		err          error
		rowsAffected int64
	)
	if len(bas) == 0 {
		return nil, errors.New("before address is empty")
	}
	as = make([]*entity.FcGenerateAddressList, 0, len(bas))
	for _, ba := range bas {
		as = append(as, &entity.FcGenerateAddressList{
			ApplyId:           ba.ApplyId,
			TaskId:            ba.TaskId,
			PlatformId:        ba.PlatformId,
			CoinId:            ba.CoinId,
			CoinName:          ba.CoinName,
			Address:           ba.Address,
			CompatibleAddress: ba.CompatibleAddress,
			Status:            ba.Status,
			Type:              ba.Type,
			OutOrderid:        ba.OutOrderid,
			Createtime:        int(time.Now().Unix()),
			Lastmodify:        util.GetChinaTimeNow(),
			IsReg:             ba.IsReg,
			IsChange:          ba.IsChange,
			Json:              ba.Json,
		})
	}

	if tx != nil {
		if rowsAffected, err = tx.Insert(as); err != nil {
			return nil, fmt.Errorf("write address list error: %w", err)
		}
	} else {
		if rowsAffected, err = db.Conn.Insert(as); err != nil {
			return nil, fmt.Errorf("write address list error: %w", err)
		}
	}

	if rowsAffected != int64(len(as)) {
		return nil, fmt.Errorf("address list must be write %d,but write %d", len(as), rowsAffected)
	}
	return as, nil
}

//查询地址是否已经存在分配
func FcGenerateAddressListIsAssign(coinName string, mchId int) (bool, error) {
	result := new(entity.FcGenerateAddressList)
	has, err := db.Conn.Where("coin_name = ? and type = 1 and platform_id =?", coinName, mchId).Get(result)
	if err != nil {
		return false, err
	}
	if !has {
		return false, errors.New("Not Fount!")
	}
	return true, nil
}

//查询商户订单是否存在
func FcGenerateAddressFindByOutOrderNo(outOrderNo string, appId int) (*entity.FcGenerateAddressList, error) {
	result := &entity.FcGenerateAddressList{}
	has, err := db.Conn.Where("out_orderid = ? and platform_id = ?", outOrderNo, appId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcGenerateAddressGet(address string) (*entity.FcGenerateAddressList, error) {
	result := new(entity.FcGenerateAddressList)
	has, err := db.Conn.Where("address = ? ", address).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcGenerateAddressGetByAddressAndMchId(address string, mchId int) (*entity.FcGenerateAddressList, error) {
	result := new(entity.FcGenerateAddressList)
	has, err := db.Conn.Where("address = ? and platform_id = ?", address, mchId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcGenerateAddressFindIn(address []string) ([]*entity.FcGenerateAddressList, error) {
	results := make([]*entity.FcGenerateAddressList, 0)
	err := db.Conn.In("address", address).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcGenerateAddressColdCount(coinType, address string) (int64, error) {
	result := new(entity.FcGenerateAddressList)
	count, err := db.Conn.Where("coin_name = ? and address = ? and type = ?", coinType, address, AddrCollectType).Count(result)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func IsInsideAddress(addr string) (bool, error) {
	result := new(entity.FcGenerateAddressList)
	exist, err := db.Conn.Table("fc_generate_address_list").Where("address = ?", addr).Exist(&result)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func FcGenerateAddressFindInternal(mchId int, coin string, addresses []string) ([]*entity.FcGenerateAddressList, error) {
	results := make([]*entity.FcGenerateAddressList, 0)
	if err := db.Conn.Table("fc_generate_address_list").Cols("address").Where(builder.In("type", []address.AddressType{address.AddressTypeCold, address.AddressTypeFee}, builder.Eq{"status": address.AddressStatusAlloc, "platform_id": mchId, "coin_name": coin}, builder.In("address", addresses))).Find(&results); err != nil {
		return nil, err
	}
	return results, nil
}

//InsertMchFirstGALAddress 插入第一个  1 归集地址（冷地址）  3 手续费地址 新地址
func InsertMchFirstGALAddress(fc *entity.FcGenerateAddressList, tx *xorm.Session) (err error) {
	_, err =tx.Insert(fc)
	return
}
