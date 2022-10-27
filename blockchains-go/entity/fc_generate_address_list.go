package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"time"
	"xorm.io/builder"
)

type AddrStatus int

type AddrType int

const (
	//状态, 0-删除, 1-未分配, 2-已分配
	AddrDelStatus   AddrStatus = 0 // 0-删除
	AddrUnUseStatus AddrStatus = 1 // 1-未分配
	AddrUsedStatus  AddrStatus = 2 // 2-已分配

	AddrOuterType   AddrType = 0 //0 外部地址
	AddrCollectType AddrType = 1 //1 归集地址（冷地址）
	AddrUserType    AddrType = 2 //2 用户地址
	AddrFeeType     AddrType = 3 //3 手续费地址
	AddrHotType     AddrType = 4 //4 热地址
	AddrBalanceType AddrType = 5 //5 商户余额地址
	AddrReceiveType AddrType = 6 //6 接收地址
)

type FcGenerateAddressList struct {
	Id                int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(10)"`
	ApplyId           int       `json:"apply_id" xorm:"not null default 0 comment('所属申请ID') INT(10)"`
	TaskId            int       `json:"task_id" xorm:"not null comment('所属任务ID') INT(10)"`
	PlatformId        int       `json:"platform_id" xorm:"not null default 0 comment('商户ID') INT(11)"`
	CoinId            int       `json:"coin_id" xorm:"not null default 0 comment('币种id') INT(10)"`
	CoinName          string    `json:"coin_name" xorm:"not null default '' comment('币种名称') unique(coin_addr) VARCHAR(16)"`
	Address           string    `json:"address" xorm:"not null default '' comment('生产的地址') unique(coin_addr) VARCHAR(255)"`
	CompatibleAddress string    `json:"compatible_address" xorm:"comment('双地址,暂时只有bch用到') VARCHAR(255)"`
	Status            int       `json:"status" xorm:"not null default 1 comment('状态, 0-删除, 1-正常, 2-已分配') TINYINT(2)"`
	Type              int       `json:"type" xorm:"not null default 0 comment('地址类型 1 归集地址（冷地址）  2 用户地址  3 手续费地址  4 热地址  5 商户余额地址,6是接收地址') TINYINT(3)"`
	OutOrderid        string    `json:"out_orderid" xorm:"not null default '' comment('合作方订单ID') index VARCHAR(64)"`
	Createtime        int       `json:"createtime" xorm:"not null comment('创建时间') INT(11)"`
	Lastmodify        time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	IsReg             int       `json:"is_reg" xorm:"not null default 0 comment('是否已注册') TINYINT(1)"`
	IsChange          int       `json:"is_change" xorm:"not null default 0 comment('0：非找零地址1：找零地址') TINYINT(1)"`
	Json              string    `json:"json" xorm:"comment('综合字段，存JSON字符串，需要自己解析') TEXT"`
}

func (o *FcGenerateAddressList) Add() (int64, error) {
	return db.Conn.InsertOne(o)
}
func (o *FcGenerateAddressList) Get(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Desc("id").Get(o)
}
func (o FcGenerateAddressList) Update(cond builder.Cond) (int64, error) {
	return db.Conn.Where(cond).Update(o)
}
func (o FcGenerateAddressList) Find(cond builder.Cond) ([]*FcGenerateAddressList, error) {
	res := make([]*FcGenerateAddressList, 0)
	if err := db.Conn.Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
func (o FcGenerateAddressList) FindAddress(cond builder.Cond) ([]string, error) {
	res := make([]string, 0)
	if err := db.Conn.Table("fc_generate_address_list").Cols("address").Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
