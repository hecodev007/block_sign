package entity

type FcMchMsg struct {
	Id      int    `json:"id" xorm:"not null pk autoincr comment('编号') INT(11)"`
	AppId   int    `json:"app_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	Title   string `json:"title" xorm:"not null comment('标题') VARCHAR(100)"`
	Msg     string `json:"msg" xorm:"not null comment('内容') VARCHAR(500)"`
	AddTime int    `json:"add_time" xorm:"not null default 0 comment('发送时间') INT(10)"`
	IsRead  int    `json:"is_read" xorm:"not null default 0 comment('是否已读(0=未读,1=已读)') TINYINT(3)"`
	Type    int    `json:"type" xorm:"not null default 0 comment('消息类型 1 系统通知 2 服务通知') TINYINT(3)"`
}
