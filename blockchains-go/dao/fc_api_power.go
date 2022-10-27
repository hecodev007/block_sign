package dao

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"strings"
	"time"
)

//查询商户有效的数据
func FcApiPowerFindsValid(mchId int) ([]*entity.FcApiPower, error) {
	results := make([]*entity.FcApiPower, 0)
	err := db.Conn.Where("user_id = ? and status = 2 and user_del = 1 and aadmin_del = 1", mchId).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//查询商户有效的数据
func FcApiPowerFindsValidCoin(coinId int, mchId int) ([]*entity.FcApiPower, error) {
	results := make([]*entity.FcApiPower, 0)
	err := db.Conn.Where("coin_id = ? and user_id = ? and status = 2 and user_del = 1 and admin_del = 1", coinId, mchId).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//InsertCustobyPowerItem 托管插入
func InsertCustobyPowerItem(apiIds []int,coinId,mchId int, cName,address, whiteIp string)(err error){
	selectStr := "insert into fc_api_power ( `api_id`, `user_id`, `coin_id`, `coin_name`," +
		"`ip`, `url`, `status`, `user_del`, `admin_del`, `createtime`) values"
	t := time.Now().Unix()
	values := make([]string,0)
	for _, item := range apiIds {
		value := fmt.Sprintf("(%v,%v,%v,'%v','%v','%v',%v,%v,%v,%v)",item,
			mchId,coinId,cName,whiteIp,address,2,1,1,t)
		values = append(values, value)
	}
	selectStr = selectStr + strings.Join(values, ",")
	_, err = db.Conn.Exec(selectStr)
	return
}
