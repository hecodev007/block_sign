package entity

type FcApiRecord struct {
	Id       int64  `json:"id" xorm:"pk autoincr BIGINT(20)"`
	AppId    int    `json:"app_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	ApiId    int    `json:"api_id" xorm:"not null default 0 comment('接口类型') TINYINT(3)"`
	CoinName string `json:"coin_name" xorm:"default '0' comment('币种') VARCHAR(20)"`
	OrderNo  string `json:"order_no" xorm:"comment('订单号') VARCHAR(100)"`
	Url      string `json:"url" xorm:"VARCHAR(255)"`
	Param    string `json:"param" xorm:"TEXT"`
	AddTime  int    `json:"add_time" xorm:"not null default 0 comment('请求时间') INT(10)"`
	Ip       string `json:"ip" xorm:"default '来源ip' VARCHAR(32)"`
	Status   int    `json:"status" xorm:"not null default 0 comment('请求状态 1 成功 2 失败  3 超时') TINYINT(3)"`
	ErrorMsg string `json:"error_msg" xorm:"comment('错误信息') VARCHAR(255)"`
}
