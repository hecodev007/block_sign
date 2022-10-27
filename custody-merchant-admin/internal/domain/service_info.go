package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type ServiceInfo struct {
	WithdrawalStatus int       `json:"withdrawal_status"`
	State            int       `json:"state"`
	Name             string    `json:"name"`
	Remark           string    `json:"remark"`
	CreateTime       time.Time `json:"create_time"`
	UpdateTime       time.Time `json:"update_time"`
}

type ChainAndName struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ChainCoinsList struct {
	Chains []ChainAndName `json:"chains"`
	Coins  []ChainAndName `json:"coins"`
}

type ServicInfoList struct {
	Id          int             `json:"id" gorm:"column:id; PRIMARY_KEY"`
	ServiceId   int             `json:"service_id,omitempty" gorm:"column:service_id"`
	ServiceName string          `json:"service_name,omitempty" gorm:"column:service_name"`
	Assets      decimal.Decimal `json:"assets,omitempty" gorm:"column:assets"`
	AssetsId    int64           `json:"assets_id,omitempty" gorm:"column:assets_id"`
	State       int             `json:"state" gorm:"column:state"`
	ChainName   string          `json:"chain_name,omitempty" gorm:"column:chain_name"`
	CoinName    string          `json:"coin_name,omitempty" gorm:"column:coin_name"`
	ChainId     int             `json:"chain_id,omitempty" gorm:"column:chain_id"`
	CoinId      int             `json:"coin_id,omitempty" gorm:"column:coin_id"`
	UpChainCoin string          `json:"up_chain_coin,omitempty" gorm:"column:up_chain_coin"`
	CreateTime  string          `json:"create_time" gorm:"column:create_time"`
	ChainAddr   string          `json:"chain_addr" gorm:"column:chain_addr"`
	IsSameAddr  int             `json:"is_same_addr" gorm:"column:is_same_addr"`
	IsClose     int             `json:"is_close" gorm:"column:is_close"`
	IsTransfer  int             `json:"is_transfer" gorm:"column:is_transfer"`
	AuditId     int             `json:"audit_id" gorm:"column:audit_id"`
	FinanceId   int             `json:"finance_id" gorm:"column:finance_id"`
	VisitorId   int             `json:"visitor_id" gorm:"column:visitor_id"`
}

type ServiceSelect struct {
	ServiceId int `json:"service_id,omitempty"`
	State     int `json:"state,omitempty"`
	Limit     int `json:"limit"`
	Offset    int `json:"offset"`
}

type ServiceHaveUser struct {
	RoleId       int                  `json:"role_id"`
	RoleName     string               `json:"role_name"`
	ServiceRoles []ServiceUserAndRole `json:"service_roles"`
	ServiceUsers []ServiceUserAndRole `json:"service_users"`
	Money        Money                `json:"money"`
}

type Money struct {
	NumEach  string `json:"num_each" description:"每笔限额"`
	NumDay   string `json:"num_day" description:"每日限额"`
	NumWeek  string `json:"num_week" description:"每周限额"`
	NumMonth string `json:"num_month" description:"每月限额"`
}

type ServiceUsers struct {
	RoleId          int     `json:"role_id"`
	ServiceId       int     `json:"service_id"`
	ServiceHaveUser []int64 `json:"service_have_user"`
	ServiceUsers    []int64 `json:"service_users"`
}

type ServiceHaveAudit struct {
	ServiceId          int                     `json:"service_id"`
	ServiceHaveUser    []int64                 `json:"service_have_user"`
	ServiceAuditLevels []ServiceHaveAuditLevel `json:"service_audit_levels"`
}

type ServiceHaveAuditLevel struct {
	Level        int     `json:"level"`
	ServiceUsers []int64 `json:"service_users"`
}

type ServiceUserAndRole struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	State    int    `json:"state"`
}

type ServiceThisUser struct {
	UserId    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	NameAndId string `json:"name_and_id"`
}

type UserHaveServiceAuditLevel struct {
	RoleId           int               `json:"role_id"`
	RoleName         string            `json:"role_name"`
	ServiceAndLevels []ServiceAndLevel `json:"service_and_levels"`
	UServices        []UService        `json:"u_services"`
}

type ServiceAndLevel struct {
	Level     int    `json:"level"`
	LevelName string `json:"level_name"`
	UService
}

type UService struct {
	ServiceId   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type MerchantService struct {
	Id          int64 `json:"id"`
	HaveService []int `json:"have_service"`
	AddService  []int `json:"add_service"`
}

type MerchantServiceList struct {
	HaveService []UServiceList `json:"have_service"`
	ServiceList []UServiceList `json:"service_list"`
}

type UServiceList struct {
	ServiceId       int    `json:"service_id"`
	ServiceName     string `json:"service_name"`
	ServiceRole     string `json:"service_role"`
	UserId          int64  `json:"user_id"`
	MerchantId      int64  `json:"merchant_id"`
	UserName        string `json:"user_name"`
	MerchantName    string `json:"merchant_name"`
	ServiceMerchant string `json:"service_merchant"`
}
