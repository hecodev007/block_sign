package entity

type FcOrderHistory struct {
	Id          int64  `json:"id" xorm:"pk autoincr BIGINT(11)"`
	ApplyId     int64  `json:"apply_id" xorm:"default 0 comment('申请id') BIGINT(20)"`
	ApplyCoinId int64  `json:"apply_coin_id" xorm:"default 0 comment('申请币id') index BIGINT(20)"`
	OrderId     int64  `json:"order_id" xorm:"default 0 comment('订单id') index BIGINT(11)"`
	Type        int    `json:"type" xorm:"default 0 comment('0:未知,1:create,2:sign:签名,3:push') TINYINT(1)"`
	TypeStr     string `json:"type_str" xorm:"default '' comment('类型描述') VARCHAR(15)"`
	JsonData    string `json:"json_data" xorm:"comment('返回数据') TEXT"`
	Status      int    `json:"status" xorm:"default 0 comment('0:正常,1:异常') TINYINT(1)"`
	CreateAt    int64  `json:"create_at" xorm:"default 0 comment('创建时间') BIGINT(11)"`
	UpdateAt    int64  `json:"update_at" xorm:"default 0 comment('最后修改时间') BIGINT(20)"`
}
