package roleMenu

import (
	"custody-merchant-admin/model"
	"fmt"
)

// 获取所有菜单mid
func SearchAllMid() (mids []string, err error) {
	entity := &Entity{}
	rows, err := model.DB().Table(entity.TableName()).Select("m_id").Rows()
	defer rows.Close()
	if err != nil {
		mids = make([]string, 0)
	}
	for rows.Next() {
		var mid int
		rows.Scan(&mid)
		mids = append(mids, fmt.Sprintf("%v", mid))
	}
	return
}
