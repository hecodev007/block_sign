package entity

type FcUserCoin struct {
	Id                int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId            int    `json:"coin_id" xorm:"not null index(idx_coinid_address) INT(11)"`
	Key               string `json:"key" xorm:"not null default '' VARCHAR(150)"`
	Address           string `json:"address" xorm:"not null index index(idx_coinid_address) VARCHAR(80)"`
	CoinNum           string `json:"coin_num" xorm:"not null default 0.0000000000 comment('币种数量') DECIMAL(23,10)"`
	WaitNum           string `json:"wait_num" xorm:"not null default 0.0000000000 comment('冻结金额') DECIMAL(23,10)"`
	SharesSum         string `json:"shares_sum" xorm:"not null default 0.0000000000 comment('升值钱包') DECIMAL(23,10)"`
	Addtime           int64  `json:"addtime" xorm:"not null comment('添加时间') BIGINT(20)"`
	AppId             int    `json:"app_id" xorm:"index INT(11)"`
	EntId             int    `json:"ent_id" xorm:"not null INT(11)"`
	ColdCode          int    `json:"cold_code" xorm:"comment('线下冷钱包编号') INT(11)"`
	BatchId           string `json:"batch_id" xorm:"comment('批量id') VARCHAR(64)"`
	Edition           int    `json:"edition" xorm:"default 0 comment('版本') INT(11)"`
	CompatibleAddress string `json:"compatible_address" xorm:"not null default '' comment('兼容地址') VARCHAR(100)"`
}
