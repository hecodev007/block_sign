package dao

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"strconv"
	"strings"
)

func FcTransfersApplyCreate(ta *entity.FcTransfersApply, tacs []*entity.FcTransfersApplyCoinAddress) (int64, error) {

	session := db.Conn.NewSession()
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

func FcTransfersApplyFindByAgree() ([]*entity.FcTransfersApply, error) {
	results := make([]*entity.FcTransfersApply, 0)
	err := db.Conn.Where("error_num <= ?", 3).In("status", []int{1, 8}).OrderBy("id asc").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcTransfersApplyUpdateStatusById(id, status int) error {
	_, err := db.Conn.Table("fc_transfers_apply").Where("id =?", id).Update(map[string]interface{}{"status": status})
	return err
}

func FcTransfersApplyUpdateStatusAndRemarkById(id, status int, remark string) error {
	_, err := db.Conn.Table("fc_transfers_apply").Where("id =?", id).Update(map[string]interface{}{"status": status, "remark": remark})
	return err
}

func FcTransfersApplyUpdateOrderIdAndStatus(id, status int, orderId string, sort int) error {
	_, err := db.Conn.Table("fc_transfers_apply").Where("id =?", id).Update(map[string]interface{}{"status": status, "order_id": orderId, "sort": sort})
	return err
}

func FcTransfersApplyUpdateStatusForRollback(id, status int) (int64, error) {
	return db.Conn.Table("fc_transfers_apply").Where("id = ? AND status = 49", id).Update(map[string]interface{}{"status": status})
}

func FcTransfersApplyUpdateStatusForRollbackForMultiAddr(id, status int) (int64, error) {
	return db.Conn.Table("fc_transfers_apply").Where("id = ?", id).Update(map[string]interface{}{"status": status})
}

//修改错误状态，并且添加错误次数
func FcTransfersApplyFail(id, status int) error {
	res, err := db.Conn.Exec("update fc_transfers_apply set status = ?,error_num = error_num + '1'  where id = ? ", status, id)
	if err != nil {
		log.Errorf("修改错误状态失败，ID：%d,修改状态：%d,error:%s", id, status, err.Error())
		return err
	}
	affected, err2 := res.RowsAffected()
	if err2 != nil {
		log.Errorf("修改错误状态失败，ID：%d,修改状态：%d,error:%s", id, status, err2.Error())
		return err2
	}
	if affected == 0 {
		err = fmt.Errorf("修改错误状态失败，ID：%d,修改状态：%d,没有影响行数", id, status)
		return err
	}
	return err
}

//更改状态增加错误计数
func FcTransfersApplyUpdateByOutNOAddErr(outorderNo string, status int) error {
	res, err := db.Conn.Exec("update fc_transfers_apply set status = ?,error_num = error_num + '1'  where out_orderid = ? ", status, outorderNo)
	if err != nil {
		log.Errorf("修改错误状态失败，outOrderNo：%s,修改状态：%d,error:%s", outorderNo, status, err.Error())
		return err
	}
	affected, err2 := res.RowsAffected()
	if err2 != nil {
		log.Errorf("修改错误状态失败，outOrderNo：%s,修改状态：%d,error:%s", outorderNo, status, err2.Error())
		return err2
	}
	if affected == 0 {
		err = fmt.Errorf("修改错误状态失败，outOrderNo：%s,修改状态：%d,没有影响行数", outorderNo, status)
		return err
	}
	return err
}

func FcTransfersApplyUpdateStatusAddErr(id, status int, orderId string) error {
	res, err := db.Conn.Exec("update fc_transfers_apply set status = ?,order_id = ?,error_num = error_num + '1'  where id = ? ", status, orderId, id)
	if err != nil {
		log.Errorf("修改错误状态失败，ID：%d,修改状态：%d,error:%s", id, status, err.Error())
		return err
	}
	affected, err2 := res.RowsAffected()
	if err2 != nil {
		log.Errorf("修改错误状态失败，ID：%d,修改状态：%d,error:%s", id, status, err2.Error())
		return err2
	}
	if affected == 0 {
		err = fmt.Errorf("修改错误状态失败，ID：%d,修改状态：%d,没有影响行数", id, status)
		return err
	}
	return err
}

//修改错误次数
func FcTransfersApplyUpdateErrNum(id, num int64) error {
	res, err := db.Conn.Exec("update fc_transfers_apply set error_num = ?  where id = ? ", num, id)
	if err != nil {
		log.Errorf("修改错误状态失败，ID：%d,修改错误次数：%d,error:%s", id, num, err.Error())
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		log.Errorf("修改错误状态失败，ID：%d,修改错误次数：%d,error:%s", id, num, err.Error())
		return err
	}
	// 如果 原始数据本来就是要修改的的数据 affected会填充0

	//if affected == 0 {
	//	err = fmt.Errorf("修改错误状态失败，ID：%d,修改错误次数：%d,没有影响行数", id, num)
	//	return err
	//}
	return err
}

func FcTransfersApplyUpdateRemark(id int64, remark string) error {
	res, err := db.Conn.Exec("update fc_transfers_apply set remark = ?  where id = ? ", remark, id)
	if err != nil {
		log.Errorf("修改transfer_apply备注失败，ID：%d,error:%s", id, err.Error())
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		log.Errorf("修改transfer_apply备注失败，ID：%d,error:%s", id, err.Error())
		return err
	}
	return err
}

//out_orderid 是唯一键，这就得要求创建的时候加商户标识了
func FcTransfersApplyById(id int) (*entity.FcTransfersApply, error) {
	result := new(entity.FcTransfersApply)
	if has, err := db.Conn.Id(id).Get(result); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

//out_orderid 是唯一键，这就得要求创建的时候加商户标识了
func FcTransfersApplyByOutOrderNo(out_orderid string) (*entity.FcTransfersApply, error) {
	result := new(entity.FcTransfersApply)
	if has, err := db.Conn.Where("out_orderid = ?", out_orderid).Get(result); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}
func FcTransfersApplyByOutOrderNoAndApplyId(out_orderid string, appId int) (*entity.FcTransfersApply, error) {
	result := new(entity.FcTransfersApply)
	if has, err := db.Conn.Where("out_orderid = ? and app_id = ?", out_orderid, appId).Get(result); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

//out_orderid 是唯一键，这就得要求创建的时候加商户标识了
func FcTransfersApplyByOrderNo(orderNO string) (*entity.FcTransfersApply, error) {
	result := new(entity.FcTransfersApply)
	if has, err := db.Conn.Where("order_id = ?", orderNO).Get(result); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

//查询有效的订单，状态为1,
func FcTransfersApplyFindValidOrder(limit int) ([]*entity.FcTransfersApply, error) {
	results := make([]*entity.FcTransfersApply, 0)
	err := db.Conn.Where("status = ? and type = ?", entity.ApplyStatus_AuditOk, "cz").OrderBy("sort desc").Limit(limit).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//查询有效的订单，状态为1,
func FcTransfersApplyGroupValidOrder(limit int) ([]*entity.FcTransfersApply, error) {
	results := make([]*entity.FcTransfersApply, 0)
	//err := db.Conn.Where("status = ? and type = ?", entity.ApplyStatus_AuditOk, "cz").GroupBy("coin_name").OrderBy("sort desc").Limit(limit).Find(&results)

	//query := "`id`, `username`, `department`, " +
	//	"`applicant`, `out_orderid`, `order_id`, `operator`, `coin_name`, " +
	//	"`type`, `purpose`, `status`, `call_back`, `error_num`, `createtime`, " +
	//	"`lastmodify`, `memo`, `fee`, `eostoken`, `eoskey`, `app_id`, `is_ding`, " +
	//	"`remark`, `source`, `code`, `is_examine`, `isforce`, `sort` "
	//
	//sqlStr := fmt.Sprintf("SELECT %s FROM (SELECT %s FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz') ORDER BY  sort desc limit ?)r  GROUP BY r.coin_name")

	err := db.Conn.SQL("SELECT * FROM (SELECT * FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz' and coin_name not in('hsc','heco','bsc','trx','eth')) ORDER BY  sort desc limit ?)r  GROUP BY r.coin_name ORDER BY createtime", limit).Find(&results)

	if err != nil {
		return nil, err
	}
	return FetchCanProcessOrders(results), nil
}

////查询有效的订单，状态为1,
//func FcTransfersApplyGroupValidOrder(limit int) ([]*entity.FcTransfersApply, error) {
//	results := make([]*entity.FcTransfersApply, 0)
//	//err := db.Conn.Where("status = ? and type = ?", entity.ApplyStatus_AuditOk, "cz").GroupBy("coin_name").OrderBy("sort desc").Limit(limit).Find(&results)
//
//	//query := "`id`, `username`, `department`, " +
//	//	"`applicant`, `out_orderid`, `order_id`, `operator`, `coin_name`, " +
//	//	"`type`, `purpose`, `status`, `call_back`, `error_num`, `createtime`, " +
//	//	"`lastmodify`, `memo`, `fee`, `eostoken`, `eoskey`, `app_id`, `is_ding`, " +
//	//	"`remark`, `source`, `code`, `is_examine`, `isforce`, `sort` "
//	//
//	//sqlStr := fmt.Sprintf("SELECT %s FROM (SELECT %s FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz') ORDER BY  sort desc limit ?)r  GROUP BY r.coin_name")
//
//	err := db.Conn.SQL("SELECT * FROM (SELECT * FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz' and coin_name not in('hsc','heco','bsc','trx')) ORDER BY  sort desc limit ?)r  GROUP BY r.coin_name ORDER BY createtime", limit).Find(&results)
//
//	if err != nil {
//		return nil, err
//	}
//	return results, nil
//}

//查询有效的订单，状态为1,
func FcTransfersApplyGroupValidOrderHSC(limit int) ([]*entity.FcTransfersApply, error) {
	results := make([]*entity.FcTransfersApply, 0)
	//err := db.Conn.Where("status = ? and type = ?", entity.ApplyStatus_AuditOk, "cz").GroupBy("coin_name").OrderBy("sort desc").Limit(limit).Find(&results)

	//query := "`id`, `username`, `department`, " +
	//	"`applicant`, `out_orderid`, `order_id`, `operator`, `coin_name`, " +
	//	"`type`, `purpose`, `status`, `call_back`, `error_num`, `createtime`, " +
	//	"`lastmodify`, `memo`, `fee`, `eostoken`, `eoskey`, `app_id`, `is_ding`, " +
	//	"`remark`, `source`, `code`, `is_examine`, `isforce`, `sort` "
	//
	//sqlStr := fmt.Sprintf("SELECT %s FROM (SELECT %s FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz') ORDER BY  sort desc limit ?)r  GROUP BY r.coin_name")
	err := db.Conn.SQL("SELECT * FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz' and coin_name = 'hsc' and app_id = 1 ) ORDER BY  createtime  limit ? ", limit).Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

//查询有效的订单，状态为1,
func FcTransfersApplyGroupValidOrderByNameDrop102(coinName string, limit int) ([]*entity.FcTransfersApply, error) {
	results := make([]*entity.FcTransfersApply, 0)
	err := db.Conn.SQL("SELECT * FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz' and coin_name = ? and app_id != 102) ORDER BY  createtime  limit ? ", coinName, limit).Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

//查询有效的订单，状态为1,
func FcTransfersApplyGroupValidOrderByName(coinName string, limit int) ([]*entity.FcTransfersApply, error) {
	results := make([]*entity.FcTransfersApply, 0)
	err := db.Conn.SQL("SELECT * FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz' and coin_name = ? ) ORDER BY  createtime  limit ? ", coinName, limit).Find(&results)
	if err != nil {
		return nil, err
	}

	return FetchCanProcessOrders(results), nil
}

// FetchCanProcessOrders 挑选需要优先处理的订单
// orders 本次需要处理的所有订单
// return 本次需要可以处理的订单
func FetchCanProcessOrders(orders []*entity.FcTransfersApply) []*entity.FcTransfersApply {
	if len(orders) == 0 {
		return orders
	}

	priorityExistList, err := FcOrderPriorityList()
	if err != nil {
		log.Errorf("调用FcOrderPriorityByApplyIds失败:%v", err)
		return orders // 直接返回，不要影响正常逻辑
	}
	if len(priorityExistList) == 0 {
		return orders
	}

	//存放优先处理订单
	// key：商户ID+链名+币种
	// value：优先处理订单的单号
	priorityMap := map[string]string{}
	ordersCanProcess := make([]*entity.FcTransfersApply, 0)

	convertMapKey := func(mchId int, chain, coinCode string) string {
		return fmt.Sprintf("%d-%s-%s", mchId, chain, coinCode)
	}

	for _, p := range priorityExistList {
		priorityMap[convertMapKey(p.MchId, p.ChainName, p.CoinCode)] = p.OuterOrderNo
	}

	for _, o := range orders {
		v, ok := priorityMap[convertMapKey(o.AppId, o.CoinName, o.Eoskey)]
		if !ok {
			// 同一商户的同一币种 没有优先订单的记录，所以本次所有订单都可以执行
			ordersCanProcess = append(ordersCanProcess, o)
		} else {
			// 如果这笔订单就是优先处理的订单，可以执行
			if o.OutOrderid == v {
				ordersCanProcess = append(ordersCanProcess, o)
			} else {
				log.Infof("商户ID=%d 链=%s 币种=%s 发现优先处理订单(%s)，所以本次订单(%s)不能正常出账，需要等到优先订单出账完成", o.AppId, o.CoinName, o.Eoskey, v, o.OutOrderid)
			}
		}
	}
	return ordersCanProcess
}

func FcTransfersApplyGroupValidOrderByNameForHold(coinName string, limit int, holdTime int64, coinCode string) ([]*entity.FcTransfersApply, error) {
	var err error
	results := make([]*entity.FcTransfersApply, 0)
	if coinCode != "" {
		err = db.Conn.SQL("SELECT * FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz' and coin_name = ?  and createtime < ? and eoskey = ?) ORDER BY  createtime  limit ? ", coinName, holdTime, limit).Find(&results)
	} else {
		err = db.Conn.SQL("SELECT * FROM `fc_transfers_apply` WHERE (status = 41 and type = 'cz' and coin_name = ?  and createtime < ?) ORDER BY  createtime  limit ? ", coinName, holdTime, limit).Find(&results)
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

func holdTransaction(coin string) (int64, string) {
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		return 0, ""
	}
	defer redisHelper.Close()
	cacheCoins, _ := redisHelper.Get("holdtransfer")
	msg := strings.Split(cacheCoins, ",")
	timeStamp := int64(0)
	coinCode := ""
	for _, s := range msg {
		arr := strings.Split(s, ":")
		if len(arr) != 3 {
			continue
		}
		if coin == arr[0] {
			coinCode = arr[1] // 币种
			timeStamp, _ = strconv.ParseInt(arr[2], 10, 64)
			break
		}
	}
	return timeStamp, coinCode
}

//查询需要重推的的订单，状态为8,
func FcTransfersApplyFindRetryOrder(limit int, errorNum int) ([]*entity.FcTransfersApply, error) {
	return nil, nil
	//results := make([]*entity.FcTransfersApply, 0)
	//err := db.Conn.Where("status = ? and error_num < ?", entity.ApplyStatus_CreateRetry, errorNum).OrderBy("id asc").Limit(limit).Find(&results)
	//if err != nil {
	//	return nil, err
	//}
	//return results, nil
}

//查询构建中的订单，状态为7,
func FcTransfersApplyFindCreateOrder(limit int, errorNum int) ([]*entity.FcTransfersApply, error) {
	results := make([]*entity.FcTransfersApply, 0)
	err := db.Conn.Where("status = ? and error_num < ?", entity.ApplyStatus_CreateOk, errorNum).OrderBy("id asc").Limit(limit).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//废弃订单
func FcTransfersApplyAbandoned(outOrderNo, orderNo string, walletType status.WalletType) error {

	if walletType != status.WalletType_Hot && walletType != status.WalletType_Cold {
		return fmt.Errorf("错误的钱包类型:%v", walletType)
	}

	session := db.Conn.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	apply := entity.FcTransfersApply{
		Status: int(entity.ApplyStatus_Rollback),
	}
	_, err = session.Cols("status").Where("order_id = ? and out_orderid = ?", orderNo, outOrderNo).Update(&apply)
	if err != nil {
		session.Rollback()
		return err
	}

	if walletType == status.WalletType_Cold {
		order := entity.FcOrder{
			Status: status.RollbackTransaction.Int(),
		}
		_, err = session.Cols("status").Where("order_no = ? and outer_order_no =? ", orderNo, outOrderNo).Update(&order)
		if err != nil {
			session.Rollback()
			return err
		}
	} else {
		order := entity.FcOrderHot{
			Status: status.RollbackTransaction.Int(),
		}
		_, err = session.Cols("status").Where("order_no = ? and outer_order_no =? ", orderNo, outOrderNo).Update(&order)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
	}
	return nil
}

func DeleteTransferApplyById(id int64) error {
	result := new(entity.FcTransfersApply)
	_, err := db.Conn.Where("id = ?", id).Delete(result)
	if err != nil {
		return err
	}
	return nil
}
