package entity

import (
	"time"
)

type FcCoinApply struct {
	Id             int64     `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinType       string    `json:"coin_type" xorm:"comment('主链(btcoin,eos..)') VARCHAR(20)"`
	CoinName       string    `json:"coin_name" xorm:"comment('主链币名') VARCHAR(20)"`
	TokenName      string    `json:"token_name" xorm:"comment('代币名(合约名)') VARCHAR(60)"`
	TokenAddress   string    `json:"token_address" xorm:"comment('代币地址(合约地址)') VARCHAR(100)"`
	IsMain         int       `json:"is_main" xorm:"comment('是否上的是代币') TINYINT(1)"`
	Precision      int       `json:"precision" xorm:"comment('币精度') INT(11)"`
	Amount         string    `json:"amount" xorm:"comment('测试币金额') VARCHAR(100)"`
	Txid           string    `json:"txid" xorm:"comment('测试币的交易id') VARCHAR(100)"`
	ToAddress      string    `json:"to_address" xorm:"comment('测试交易接收地址') VARCHAR(255)"`
	FeeType        int       `json:"fee_type" xorm:"default 0 comment('手续费类型(0:消耗代币,1:消耗主链币,2:消耗其他币,3:不消耗手续费)') TINYINT(1)"`
	FeeName        string    `json:"fee_name" xorm:"comment('手续费币名') VARCHAR(20)"`
	FeeAmount      string    `json:"fee_amount" xorm:"comment('测试手续费金额') VARCHAR(100)"`
	LogoPath       string    `json:"logo_path" xorm:"comment('logo图标上传的路径') VARCHAR(200)"`
	OfficialUrl    string    `json:"official_url" xorm:"comment('上币的官方地址') VARCHAR(200)"`
	Proposer       string    `json:"proposer" xorm:"comment('发起人(联系人)') VARCHAR(60)"`
	Cellphone      string    `json:"cellphone" xorm:"comment('发起人(联系人)手机') VARCHAR(20)"`
	Email          string    `json:"email" xorm:"comment('发起人(联系人)邮箱') VARCHAR(100)"`
	ReceiveAddress string    `json:"receive_address" xorm:"comment('测试币接收地址') VARCHAR(255)"`
	ScheduleTime   time.Time `json:"schedule_time" xorm:"comment('计划上币时间') DATETIME"`
	ActualTime     time.Time `json:"actual_time" xorm:"comment('确认上币时间') DATETIME"`
	AuditStatus    int       `json:"audit_status" xorm:"default 0 comment('审核状态(0: 申请中, 1:审核通过,2: 不通过,3: 阻塞)') TINYINT(1)"`
	OperaStatus    int       `json:"opera_status" xorm:"default 0 comment('处理状态 0 待处理  1 测试中 2 测试通过') TINYINT(1)"`
	ProcessRemark  string    `json:"process_remark" xorm:"comment('测试备注(所有测试结果内容都写此)') VARCHAR(200)"`
	CreateAt       time.Time `json:"create_at" xorm:"default CURRENT_TIMESTAMP DATETIME"`
	UpdateAt       time.Time `json:"update_at" xorm:"default CURRENT_TIMESTAMP DATETIME"`
}
