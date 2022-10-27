package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func FcWorkerGetByCoinName(coinName string) ([]*entity.FcWorker, error) {
	results := make([]*entity.FcWorker, 0)
	err := db.Conn.Where("status = ? and coin_name not like ?", 1, "%"+coinName+"%").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcWorkerFind() ([]*entity.FcWorker, error) {
	results := make([]*entity.FcWorker, 0)
	err := db.Conn.Where("status = ? ", 1).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
