package fullYear

import (
	"custody-merchant-admin/module/log"
)

func (e *Entity) InsertNewItem() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf(" insert fullYear error: %v", err)
	}
	return
}

func (e *Entity) FindItemById(uid, pid int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("account_id = ? and package_id = ?", uid, pid).Find(e).Error
	if err != nil {
		log.Errorf("find fullYear error: %v", err)
	}
	if e.Id == 0 {
		err = e.Db.Table(e.TableName()).Where("account_id = ? and package_id = 0", uid).Find(e).Error
		if err != nil {
			log.Errorf("find fullYear error: %v", err)
		}
	}
	return
}

func (e *Entity) DeleteItemById(id int) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ?", id).Delete(e).Error
	if err != nil {
		log.Errorf(" del fullYear error: %v", err)
	}
	return
}
