package entity

type FcTrade struct {
	Id      int64  `json:"id" xorm:"pk autoincr BIGINT(20)"`
	Action  string `json:"action" xorm:"unique(action) VARCHAR(255)"`
	TradeId string `json:"trade_id" xorm:"unique(action) VARCHAR(255)"`
	TradeSn string `json:"trade_sn" xorm:"not null index VARCHAR(20)"`
}
