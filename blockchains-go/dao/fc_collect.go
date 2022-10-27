package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"time"
)

func FcCollectInsert(fc *entity.FcCollect) (id int64, err error) {
	_, err = db.Conn.Insert(fc)

	id = fc.Id
	log.Infof("FcCollectInsert 返回ID：%d", id)
	return id, err
}

func FcCollectUpdateStateById(id int64, status entity.FcCollectStatus) error {
	fc := entity.FcCollect{
		Status:     status,
		UpdateTime: time.Now().Unix(),
	}
	_, err := db.Conn.Id(id).Cols("status", "update_time").Update(&fc)
	return err
}

func FcCollectUpdateStateSuccess(id int64, txid string) error {
	fc := entity.FcCollect{
		Txid:       txid,
		Status:     entity.FcCollectSuccess,
		UpdateTime: time.Now().Unix(),
	}
	_, err := db.Conn.Id(id).Cols("txid", "status", "update_time").Update(&fc)
	return err
}

//更新为失败订单并且表明错误信息
func FcCollectUpdateFailOrder(id int64, errorData string) error {
	fc := entity.FcCollect{
		Status:     entity.FcCollectFail,
		ErrData:    errorData,
		UpdateTime: time.Now().Unix(),
	}
	_, err := db.Conn.Id(id).Cols("status", "update_time", "err_data").Update(&fc)
	return err
}

//查询需要归集的订单
func FcCollectFindCollect(limit int) ([]*entity.FcCollect, error) {
	results := make([]*entity.FcCollect, 0)
	err := db.Conn.Where("status = ?", entity.FcCollectCreate).Asc("create_time").Limit(limit).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
