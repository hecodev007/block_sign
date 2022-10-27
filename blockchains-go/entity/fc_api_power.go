package entity

import (
	"time"
)

type FcApiPower struct {
	Id            int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	ApiId         int       `json:"api_id" xorm:"default 0 comment('apilist的ID') index(api_id) INT(11)"`
	UserId        int       `json:"user_id" xorm:"default 0 comment('商户的ID') index(api_id) INT(11)"`
	CoinId        int       `json:"coin_id" xorm:"default 0 comment('coin_set表id') INT(11)"`
	CoinName      string    `json:"coin_name" xorm:"comment('币种名称') index(api_id) VARCHAR(15)"`
	Content       string    `json:"content" xorm:"comment('商户描述') VARCHAR(255)"`
	Ip            string    `json:"ip" xorm:"comment('API白名单') VARCHAR(255)"`
	Url           string    `json:"url" xorm:"comment('回调地址') VARCHAR(255)"`
	Status        int       `json:"status" xorm:"default 1 comment('1：待审核2：审核通过3：审核驳回') TINYINT(3)"`
	UserDel       int       `json:"user_del" xorm:"default 1 comment('商户删除1：正常2：删除') TINYINT(3)"`
	AdminDel      int       `json:"admin_del" xorm:"default 1 comment('后台删除：1正常2禁用') TINYINT(3)"`
	StatusContent string    `json:"status_content" xorm:"comment('后台审核备注') TEXT"`
	DelContent    string    `json:"del_content" xorm:"comment('后台禁用备注') TEXT"`
	Createtime    int       `json:"createtime" xorm:"default 0 comment('申请时间') INT(11)"`
	Lastmodify    time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
