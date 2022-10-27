package entity

import (
	"time"
)

type FcHuobiOrdersMatchresults struct {
	Id             int64     `json:"id" xorm:"pk autoincr comment('订单成交记录ID') BIGINT(20)"`
	OrderId        int64     `json:"order_id" xorm:"not null comment('订单 ID') BIGINT(20)"`
	MatchId        int64     `json:"match_id" xorm:"not null comment('撮合ID') BIGINT(20)"`
	Symbol         string    `json:"symbol" xorm:"not null comment('交易对,btcusdt, bchbtc, rcneth ...') VARCHAR(20)"`
	Type           string    `json:"type" xorm:"not null comment('订单类型,buy-market：市价买, sell-market：市价卖, buy-limit：限价买, sell-limit：限价卖, buy-ioc：IOC买单, sell-ioc：IOC卖单') VARCHAR(20)"`
	Source         string    `json:"source" xorm:"not null comment('订单来源,api') VARCHAR(15)"`
	Price          string    `json:"price" xorm:"not null comment('成交价格') DECIMAL(65,20)"`
	FilledAmount   string    `json:"filled_amount" xorm:"not null comment('成交数量') DECIMAL(65,20)"`
	FilledFees     string    `json:"filled_fees" xorm:"not null comment('成交手续费') DECIMAL(65,20)"`
	FilledPoints   string    `json:"filled_points" xorm:"not null comment('文档上没有此字段说明，') VARCHAR(28)"`
	CreatedAt      string    `json:"created_at" xorm:"not null comment('成交时间，抓取的数据') index VARCHAR(15)"`
	CreateDatetime time.Time `json:"create_datetime" xorm:"not null default '0000-00-00 00:00:00' comment('成交时间-日期格式,created_at进行了转换便于查看') DATETIME"`
	Lastmodify     time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最近修改日期') TIMESTAMP"`
}
