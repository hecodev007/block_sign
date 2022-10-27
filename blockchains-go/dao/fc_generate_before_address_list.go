package dao

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"strings"
	"time"
	"xorm.io/xorm"
)

type AddrStatus int

type AddrType int

const (
	//状态, 0-删除, 1-未分配, 2-已分配
	AddrDelStatus   AddrStatus = 0 // 0-删除
	AddrUnUseStatus AddrStatus = 1 // 1-未分配
	AddrUsedStatus  AddrStatus = 2 // 2-已分配

	AddrOuterType   AddrType = 0 //0 外部地址
	AddrCollectType AddrType = 1 //1 归集地址（冷地址）
	AddrUserType    AddrType = 2 //2 用户地址
	AddrFeeType     AddrType = 3 //3 手续费地址
	AddrHotType     AddrType = 4 //4 热地址
	AddrBalanceType AddrType = 5 //5 商户余额地址
	AddrReceiveType AddrType = 6 //6 接收地址
)

//给商户分配地址
func AssignMchAddrs(mchId int64,mchName, coinName string, outerOrderId string, num int64) ([]*entity.FcGenerateAddressList,int, error) {
	var (
		err error
		bas []*entity.FcGenerateBeforeAddressList
		as  []*entity.FcGenerateAddressList
		remain int
	)
	tx := db.Conn.NewSession()
	defer func() {
		if r := recover(); r != nil {
			//log要记录错误
			tx.Rollback()
		}
		tx.Close()
	}()
	if err = tx.Begin(); err != nil {
		return nil,0, fmt.Errorf("db begin error: %w", err)
	}
	//这个语句效率慢
	//if bas, err = GetAndMarkUnuseAddr(mchId, coinName, outerOrderId, num, tx); err != nil {
	//	tx.Rollback()
	//	return nil, fmt.Errorf("GetAndMarkUnuseAddr error: %w", err)
	//}
	if bas,remain, err = GetAndMarkUnuseAddr2(mchId, coinName, outerOrderId, int(num), tx); err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(),"no enough") {
			return nil,remain, err
		}
		return nil,remain, fmt.Errorf("GetAndMarkUnuseAddr error: %w", err)
	}

	if as, err = WriteMchAddrs(bas, tx); err != nil {
		tx.Rollback()
		return nil,remain, fmt.Errorf("WriteMchAddrs error: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil,remain, fmt.Errorf("db commit error: %w", err)
	}
	return as,remain, nil
}

//获取未使用的地址信息
func GetAndMarkUnuseAddr(mchId int64, coinName string, outerOrderId string, num int64, tx *xorm.Session) ([]*entity.FcGenerateBeforeAddressList, error) {
	var (
		counts       int64
		ids          []int
		bas          []*entity.FcGenerateBeforeAddressList
		err          error
		rowsAffected int64
	)
	//get
	if tx != nil {
		//该事务做了两次查询，并且由于表的索引不当引起慢查询
		if counts, err = tx.Where("platform_id = ? and coin_name = ? and status = ?", mchId, coinName, AddrUnUseStatus).Limit(int(num)).FindAndCount(&bas); err != nil {
			return nil, err
		}

	} else {
		if counts, err = db.Conn.Where("platform_id = ? and coin_name = ? and status = ?", mchId, coinName, AddrUnUseStatus).Limit(int(num)).FindAndCount(&bas); err != nil {
			return nil, err
		}
	}

	if counts < num {
		return nil, errors.New("unuse address no enough supply")
	}

	for _, ba := range bas {
		ids = append(ids, ba.Id)
	}

	if tx != nil {
		if rowsAffected, err = tx.In("id", ids).Update(&entity.FcGenerateBeforeAddressList{Status: int(AddrUsedStatus), Type: int(AddrUserType), OutOrderid: outerOrderId}); err != nil {
			return nil, fmt.Errorf("mark unuse address error: %w", err)
		}
	} else {
		if rowsAffected, err = db.Conn.In("id", ids).Update(&entity.FcGenerateBeforeAddressList{Status: int(AddrUsedStatus), Type: int(AddrUserType), OutOrderid: outerOrderId}); err != nil {
			return nil, fmt.Errorf("mark unuse address error: %w", err)
		}
	}

	if rowsAffected != int64(len(bas)) {
		return nil, fmt.Errorf("address list must be mark %d,but mark %d", len(bas), rowsAffected)
	}

	for i, _ := range bas {
		bas[i].Status = int(AddrUsedStatus)
		bas[i].Type = int(AddrUserType)
		bas[i].OutOrderid = outerOrderId
	}
	return bas, nil
}

//GetAndMarkUnuseAddr2 获取未使用的地址信息 并返回剩余数量 存在剩余数量<0情况
func GetAndMarkUnuseAddr2(mchId int64, coinName string, outerOrderId string, num int, tx *xorm.Session) (bas []*entity.FcGenerateBeforeAddressList,remain int,err error) {
	var (
		ids          []int
		rowsAffected int64
		total int
	)
	bas = make([]*entity.FcGenerateBeforeAddressList,0)
	//get
	if tx != nil {
		//该事务做了两次查询，并且由于表的索引不当引起慢查询
		list := make([]entity.FcGenerateBeforeAddressList,0)
		err = tx.Select("id").Where("platform_id = ? and coin_name = ? and status = ?", mchId, coinName, AddrUnUseStatus).Find(&list)
		total = len(list)
		err = tx.Where("platform_id = ? and coin_name = ? and status = ?", mchId, coinName, AddrUnUseStatus).Limit(int(num)).Find(&bas)

	} else {
		list := make([]entity.FcGenerateBeforeAddressList,0)
		db.Conn.Select("id").Where("platform_id = ? and coin_name = ? and status = ?", mchId, coinName, AddrUnUseStatus).Find(&list)
		total = len(list)
		err = db.Conn.Where("platform_id = ? and coin_name = ? and status = ?", mchId, coinName, AddrUnUseStatus).Limit(int(num)).Find(&bas)
	}
	if err != nil {
		return
	}
	log.Infof("商户：%d,指定数量：%d,coinName：%v，limit查询数量：%d", mchId, num,coinName, len(bas))
	log.Infof("total:%d",total)
	//剩余数量
	remain = total - num
	if len(bas) < num {
		err = errors.New("unuse address no enough supply")
		return
	}

	for _, ba := range bas {
		ids = append(ids, ba.Id)
	}
	if tx != nil {
		if rowsAffected, err = tx.In("id", ids).Update(&entity.FcGenerateBeforeAddressList{Status: int(AddrUsedStatus), Type: int(AddrUserType), OutOrderid: outerOrderId}); err != nil {
			err =  fmt.Errorf("mark unuse address error: %w", err)
			return
		}
	} else {
		if rowsAffected, err = db.Conn.In("id", ids).Update(&entity.FcGenerateBeforeAddressList{Status: int(AddrUsedStatus), Type: int(AddrUserType), OutOrderid: outerOrderId}); err != nil {
			err =  fmt.Errorf("mark unuse address error: %w", err)
			return
		}
	}

	if rowsAffected != int64(len(bas)) {
		err =  fmt.Errorf("address list must be mark %d,but mark %d", len(bas), rowsAffected)
		return
	}

	for i, _ := range bas {
		bas[i].Status = int(AddrUsedStatus)
		bas[i].Type = int(AddrUserType)
		bas[i].OutOrderid = outerOrderId
	}
	return
}



//GetMchAddr2 获取用户 typ未使用的地址类型  1 归集地址（冷地址）  3 手续费地址
func GetMchAddr2(mchId int64, coinName string) (one,three int,err error) {
	var typeOne  []entity.FcGenerateBeforeAddressList
	var typeThree []entity.FcGenerateBeforeAddressList
	db.Conn.Table("fc_generate_before_address_list").Where("platform_id = ? and coin_name = ? and status = ? and type = ?", mchId, coinName, AddrUsedStatus,AddrCollectType).Find(&typeOne)
	db.Conn.Table("fc_generate_before_address_list").Where("platform_id = ? and coin_name = ? and status = ? and type = ?", mchId, coinName, AddrUsedStatus,AddrFeeType).Find(&typeThree)
	one = len(typeOne)
	three = len(typeThree)
	return
}

//InsertMchFirstAddress 插入第一个  1 归集地址（冷地址）  3 手续费地址 新地址
func InsertMchFirstAddress(fc *entity.FcGenerateBeforeAddressList, tx *xorm.Session) (err error) {
	_, err =tx.Insert(fc)
	return
}

//InsertBatchMchAddress 批量插入冷钱包地址 新地址
func InsertBatchMchAddress(mchId ,coinId int,coinName,outOrderid string,address[]string,t time.Time, tx *xorm.Session) (err error) {
	selectStr := "insert into fc_generate_before_address_list (platform_id,coin_id,coin_name,task_id,address,status,type,out_orderid,createtime) values "
	values := make([]string, 0)
	tNum := t.Unix()
	for _, item := range address {
		value := fmt.Sprintf("(%v,%v,'%v',%v,'%v',%v,%v,'%v',%v)", mchId,coinId, coinName,0,item,1,2,outOrderid,tNum)
		values = append(values, value)
	}
	selectStr = selectStr + strings.Join(values, ",")
	_,err = tx.Exec(selectStr)
	return
}
