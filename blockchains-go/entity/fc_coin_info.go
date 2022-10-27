package entity

import (
	"time"
)

type FcCoinInfo struct {
	Id                int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	Name              string    `json:"name" xorm:"not null comment('英文简写') unique VARCHAR(20)"`
	NameCn            string    `json:"name_cn" xorm:"not null default '' comment('中文名称') VARCHAR(20)"`
	NameEng           string    `json:"name_eng" xorm:"not null default '' comment('币种英文名称') VARCHAR(20)"`
	NameEngFeixiaohao string    `json:"name_eng_feixiaohao" xorm:"not null default '' comment('英文全名，用于拼装url抓数据') VARCHAR(64)"`
	NameEngChaince    string    `json:"name_eng_chaince" xorm:"not null default '' comment('币种英文名称, 在chaince.com上的名称 ') VARCHAR(20)"`
	Url               string    `json:"url" xorm:"not null default '' comment('抓取的url，可以带参数') VARCHAR(255)"`
	Status            int       `json:"status" xorm:"not null default 1 comment('状态,1-正常、0-删除') TINYINT(2)"`
	Description       string    `json:"description" xorm:"not null default '' comment('预留字段') VARCHAR(200)"`
	Lastmodify        time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
