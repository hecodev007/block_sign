package entity

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"time"
	"xorm.io/builder"
)

type FcCoinSet struct {
	Id               int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Pid              int       `json:"pid" xorm:"default 0 comment('上级id') INT(11)"`
	Name             string    `json:"name" xorm:"not null comment('英文简写') unique(idx_name) VARCHAR(25)"`
	Title            string    `json:"title" xorm:"not null comment('中文名') VARCHAR(25)"`
	Connect          string    `json:"connect" xorm:"not null comment('客户端url') VARCHAR(255)"`
	Status           int       `json:"status" xorm:"default 1 comment('币种状态') TINYINT(4)"`
	WStatus          int       `json:"w_status" xorm:"not null default 1 comment('提现状态') TINYINT(4)"`
	RStatus          int       `json:"r_status" xorm:"not null default 1 comment('充值状态') TINYINT(4)"`
	Sort             int       `json:"sort" xorm:"not null default 0 comment('排序值') TINYINT(4)"`
	Addtime          time.Time `json:"addtime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	Type             int       `json:"type" xorm:"not null default 0 comment('类型  1 主链  2 侧链（代币）') TINYINT(4)"`
	Token            string    `json:"token" xorm:"comment('合约地址') unique(idx_name) VARCHAR(255)"`
	Decimal          int       `json:"decimal" xorm:"not null default 0 comment('精度位数') TINYINT(3)"`
	Num              string    `json:"num" xorm:"not null default 0.0000000000 comment('最小提现数量') DECIMAL(23,10)"`
	HugeNum          string    `json:"huge_num" xorm:"not null default 0.000000000000000000 comment('巨额提现临界数量') DECIMAL(40,18)"`
	Rate             float32   `json:"rate" xorm:"not null default 0.000000 comment('流水费率') FLOAT(12,6)"`
	ServiceFee       string    `json:"service_fee" xorm:"not null default 0.0000000000 comment('年服务费(usdt)') DECIMAL(23,10)"`
	Code             int       `json:"code" xorm:"not null default 0 comment('0不启用多签1启用多签') TINYINT(1)"`
	FeeLimit         string    `json:"fee_limit" xorm:"not null default 0.0000000000 comment('矿工费准备金阀值') DECIMAL(23,10)"`
	Maximum          string    `json:"maximum" xorm:"not null default 0.0000000000 comment('最大余额') DECIMAL(23,10)"`
	Confirm          int       `json:"confirm" xorm:"default 6 comment('币种确认数') INT(11)"`
	Asset            string    `json:"asset" xorm:"VARCHAR(255)"`
	Price            string    `json:"price" xorm:"not null default 0.0000000000 comment('币种相对于rmb的价格') DECIMAL(23,10)"`
	CollectFee       string    `json:"collect_fee" xorm:"not null default 0.000000000000000000 comment('归集手续费阈值') DECIMAL(40,18)"`
	IsOpen           int       `json:"is_open" xorm:"not null default 0 comment('是否为公链云开放币种 1 开放 0 未开放') TINYINT(1)"`
	ApiListId        string    `json:"api_list_id" xorm:"comment('权限关联') VARCHAR(100)"`
	HugeFee          string    `json:"huge_fee" xorm:"not null default 0.000000000000000000 comment('手续费临界值') DECIMAL(40,18)"`
	HugeExtra        string    `json:"huge_extra" xorm:"not null default 0.000000000000000000 comment('部分代币附加的主币临界值') DECIMAL(40,18)"`
	PatternType      int       `json:"pattern_type" xorm:"not null default 1 comment('1.账户模型；2.utxo模型') TINYINT(4)"`
	ContrastConfirm  int       `json:"contrast_confirm" xorm:"not null default 6 comment('对账确认数') TINYINT(4)"`
	SuportRegister   int       `json:"suport_register" xorm:"not null default 0 TINYINT(4)"`
	V2GjMchName      string    `json:"v2_gj_mch_name" xorm:"not null default '' comment('需要归集的商户，不填则没有商户需要归集') VARCHAR(255)"`
	V2GjUserNum      string    `json:"v2_gj_user_num" xorm:"not null default 0.00000000000000000000 comment('触发归集阀值，大于这个值才归集') DECIMAL(40,20)"`
	V2GjChangeNum    string    `json:"v2_gj_change_num" xorm:"not null default 0.00000000000000000000 comment('手续费告警阀值,垫资地址少于这个值,告警') DECIMAL(40,20)"`
	V2GjUserFee      string    `json:"v2_gj_user_fee" xorm:"not null default 0.00000000000000000000 comment('用户地址打手续费阀值') DECIMAL(40,20)"`
	V2GjUserFeeNum   string    `json:"v2_gj_user_fee_num" xorm:"not null default 0.00000000000000000000 comment('用户地址每次打多少手续费') DECIMAL(40,20)"`
	V2GjFeeCoin      string    `json:"v2_gj_fee_coin" xorm:"not null default '' comment('消耗手续费币种') VARCHAR(20)"`
	V2GjDate         int       `json:"v2_gj_date" xorm:"not null default 0 comment('归集时间') INT(11)"`
	V2IsGj           int       `json:"v2_is_gj" xorm:"not null default 0 comment('0:不需要归集 1：需要归集') TINYINT(4)"`
	V2GjMchFee       string    `json:"v2_gj_mch_fee" xorm:"not null default 0.00000000000000000000 comment('出账地址打手续费阀值') DECIMAL(40,20)"`
	V2GjMchFeeNum    string    `json:"v2_gj_mch_fee_num" xorm:"not null default 0.00000000000000000000 comment('出账地址每次打多少手续费') DECIMAL(40,20)"`
	V2FeeType        int       `json:"v2_fee_type" xorm:"not null default 0 comment('0:一对一打手续费 1：一对多打手续费') TINYINT(4)"`
	HoldAmount       string    `json:"hold_amount" xorm:"hold_amount"` //拦截金额预警
	CollectThreshold string    `json:"collect_threshold"`              //归集阈值
	StaThreshold     string    `json:"sta_threshold"`                  //归集阈值
	IsCollect        int       `json:"is_collect"`                     //是否归集,0false 1true
	UnionId          int       `json:"union_id" xorm:"default 0 comment('全局id，与交易所统一') INT(11)"`
}

func (o *FcCoinSet) Add() (int64, error) {
	return db.Conn.InsertOne(o)
}
func (o *FcCoinSet) Get(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Desc("id").Get(o)
}
func (o FcCoinSet) Update(cond builder.Cond) (int64, error) {
	return db.Conn.Where(cond).Update(o)
}
func (o FcCoinSet) Find(cond builder.Cond) ([]*FcCoinSet, error) {
	res := make([]*FcCoinSet, 0)
	if err := db.Conn.Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
func (o *FcCoinSet) GetDecimal(cond builder.Cond) (int, error) {
	c := 0
	if has, err := db.Conn.Table("fc_coin_set").Select("decimal").Where(cond).Get(&c); err != nil {
		return -1, err
	} else if !has {
		return -1, fmt.Errorf("dont't find coin set")
	}
	return c, nil
}
