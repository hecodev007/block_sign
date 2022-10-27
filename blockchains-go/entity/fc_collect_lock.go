package entity

type FcCollectLock struct {
	Id              int64 `json:"id" xorm:"not null pk autoincr BIGINT(20)"`
	AddressAmountId int64 `json:"address_amount_id" xorm:"default 0 comment('address') BIGINT(20)"`
	IsLock          bool  `json:"is_lock" xorm:"comment('是否锁定') BIT"`
	UpdateAt        int64 `json:"update_at" xorm:"not null default 0 BIGINT(20)"`
}
