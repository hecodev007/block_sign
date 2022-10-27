package base

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/entity"
)

//==========================================db==========================================

//btc特定查询使用，查询末尾金额。过滤指定金额，因为现有逻辑btc和usdt地址混在一起了，唯一区分只能是3开头地址
//params appId: 	商户ID
//params limit: 	需要的记录数
func FindMergeAddr(appId int64, limit uint) ([]*entity.FcAddressAmount, error) {
	result := make([]*entity.FcAddressAmount, 0)
	//err := Conn.Where("address like '3%' and type = 2 and coin_type = 'btc' and amount >= ? and amount <= ? and app_id = ?",
	//	CoinCfg.MinAmount.String(),
	//	CoinCfg.MaxAmount.String(),
	//	appId).Find(&result)

	//先暂时写死
	err := Conn.Where("address like '3%' and type in(1,2) and address != '3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw' and coin_type = 'btc' and amount >= ? and amount <= ? and app_id = ?",
		CoinCfg.MinAmount.String(),
		CoinCfg.MaxAmount.String(),
		appId).OrderBy("amount desc").Limit(int(limit)).Find(&result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

//btc特定查询使用，查询末尾金额。过滤指定金额，因为现有逻辑btc和usdt地址混在一起了，唯一区分只能是3开头地址
//params appId: 	商户ID
//params limit: 	需要的记录数
func FindBTCMergeAddr(appId int64, limit uint) ([]*entity.FcAddressAmount, error) {
	result := make([]*entity.FcAddressAmount, 0)
	err := Conn.Where("address like '3%' and type in (1,2) and coin_type = 'btc' and amount >= ? and amount <= ? and app_id = ?",
		CoinCfg.MinAmount.String(),
		CoinCfg.MaxAmount.String(),
		appId).OrderBy("amount desc").Limit(int(limit)).Find(&result)

	////先暂时写死
	//err := Conn.Where("address like '3%' and type in(1,2) and address != '3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw' and coin_type = 'btc' and amount >= ? and amount <= ? and app_id = ?",
	//	CoinCfg.MinAmount.String(),
	//	CoinCfg.MaxAmount.String(),
	//	appId).Find(&result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

//根据币种获取出账冷地址
func FindBTCMergeAddrStrs(mchId int64) ([]string, error) {
	results := make([]string, 0)
	err := Conn.Table("fc_generate_address_list").Cols("address").
		Where("address like '3%' and type = 1 and status = 2 and platform_id = ? and coin_name = 'btc'", mchId).
		Find(&results)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("don't find any address")
	}
	return results, nil
}

func FindMchIds(coinName string) ([]*entity.FcMchService, error) {
	results := make([]*entity.FcMchService, 0)
	err := Conn.Where("status = 0 and coin_name = ?", coinName).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FindMch(id int64) (*entity.FcMch, error) {
	result := new(entity.FcMch)
	has, err := Conn.ID(id).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

//根据币种获取出账冷地址
func FindToAddr(typed, status int, mchId int64, coinName string) ([]string, error) {
	results := make([]string, 0)
	err := Conn.Table("fc_generate_address_list").Cols("address").
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

//status 有效
func FindCoin(coinName string) (*entity.FcCoinSet, error) {
	result := new(entity.FcCoinSet)
	has, err := Conn.Where("name = ? and  status = 1", coinName).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil

}

func SaveApplyOrder(ta *entity.FcTransfersApply, tacs []*entity.FcTransfersApplyCoinAddress) (int64, error) {

	session := Conn.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return -1, err
	}
	_, err = session.InsertOne(ta)
	if err != nil {
		session.Rollback()
		return -1, err
	}

	if len(tacs) > 0 {
		for _, tac := range tacs {
			tac.ApplyId = int64(ta.Id)
		}
		_, err = session.Insert(tacs)
		if err != nil {
			session.Rollback()
			return -1, err
		}
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
		return -1, err
	}
	return int64(ta.Id), nil
}

func FindWorkers() ([]*entity.FcWorker, error) {
	results := make([]*entity.FcWorker, 0)
	err := Conn.Where("status = ? ", 1).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
