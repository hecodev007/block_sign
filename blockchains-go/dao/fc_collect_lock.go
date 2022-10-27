package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"time"
)

func FcCollectLockIds() ([]int64, error) {
	results := make([]int64, 0)
	err := db.Conn.Table("fc_collect_lock").Cols("address_amount_id").
		Where("is_lock = ? AND update_at >= ?", 1, time.Now().Unix()-60*10).
		Find(&results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		results = append(results, 0)
	}
	return results, nil
}
