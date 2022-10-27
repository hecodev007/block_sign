package dict

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/adminPermission/api"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/unitUsdt"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

var (
	HooFee = map[string]HooFeeInfo{}
	// SysRoleNameList
	//  0: "超级管理员",
	//	1: "管理员",
	//	2: "审核员",
	//	3: "财务员",
	//	4: "游客",
	SysRoleNameList = []string{"超级管理员", "管理员", "审核员", "财务员", "游客"}

	// SysAdminRoleNameList
	//  0: "超级管理员",
	//	1: "管理员",
	//	2: "财务员",
	//	3: "游客",
	SysAdminRoleNameList = []string{"超级管理员", "管理员", "财务员", "游客"}

	// IsTestNameList
	//  0: "正式账号",
	//	1: "测试账号"
	IsTestNameList = []string{"正式账号", "测试账号"}
	// AuditLevelName
	//  0: "初级审核员",
	//	1: "中级审核员",
	//	2: "高级审核员",
	//	3: "超级审核员",
	AuditLevelName = []string{"初级审核员", "中级审核员", "高级审核员", "超级审核员"}

	// SysRoleTagList
	//  0: "administrator",
	//	1: "admin",
	//	2: "audit",
	//	3: "finance",
	//	4: "visitor",
	SysRoleTagList = []string{"administrator", "admin", "finance", "visitor"}

	// SysMerchantRoleTagList
	//  0: "administrator",
	//	1: "admin",
	//	2: "audit",
	//	3: "finance",
	//	4: "visitor",
	SysMerchantRoleTagList = []string{"administrator", "admin", "audit", "finance", "visitor"}

	// AuditTypeList
	//  0: "一审终审制",
	//	1: "二审终审制",
	//	2: "三审终审制",
	//	3: "四审终审制",
	//	4: "五审终审制",
	AuditTypeList = []string{"一审终审制", "二审终审制", "三审终审制", "四审终审制", "五审终审制"}

	// AuditUserList
	//  0: "一审人员",
	//	1: "二审人员",
	//	2: "三审人员",
	//	3: "四审人员",
	//	4: "五审人员",
	//	5: "超级审人员",
	AuditUserList = []string{"一审人员", "二审人员", "三审人员", "四审人员", "五审人员", "超级审人员"}

	// StateText
	// 0: "正常"
	// 1: "冻结"
	// 2: "失效"
	StateText = []string{"正常", "冻结", "失效"}

	// SexText
	// 0: "男"
	// 1: "女"
	SexText = []string{"男", "女"}

	// WalletDealText
	// 0: "未处理"
	// 1: "是"
	// 2: "否"
	WalletDealText = []string{"未处理", "是", "否"}
	// BaseText
	// 0: "否"
	// 1: "是"
	BaseText = []string{"否", "是"}

	// WalletStateText
	// 0: "处理中"
	// 1: "处理完成"
	WalletStateText = []string{"处理中", "处理完成"}

	// WalletResultText
	// 0: ""
	// 1: "未出账"
	// 2: "已出账"
	WalletResultText = []string{"", "未出账", "已出账"}

	// IsTestText
	// 0: "正式账号"
	// 1: "测试账号"
	IsTestText = []string{"正式账号", "测试账号"}

	// BillState
	// 0: "链上地址转入-正在接收"
	// 1: "链上地址转入-接收成功"
	// 2: "链上地址转入-接收失败"
	// 3: "转到链上地址-冻结"
	// 4: "转到链上地址-确认"
	// 5: "转到链上地址-失败回滚"
	BillState = []string{"链上地址转入-正在接收", "链上地址转入-接收成功", "链上地址转入-接收失败", "转到链上地址-冻结", "转到链上地址-确认", "转到链上地址-失败回滚"}

	// TxTypeList
	// 0: "发送"
	// 1: "接收"
	TxTypeList = []string{"发送", "接收"}
	// OrderResult
	// 0:"待审核"
	// 1:"已通过"
	// 2:"冻结"
	// 3:"解冻"
	// 4:"拒绝"
	OrderResult = []string{"待审核", "已通过", "已冻结", "已解冻", "已拒绝"}

	OrderResultColor = map[string]string{}
	// OrderType
	// 0:"提现"
	// 1:"充值"
	OrderType = []string{"提现"}
	Colors    = []string{"#F7B500", "#5B76F9", "#36C3FC", "#50DFB2", "#2B4563", "#F7B910", "#5B1111", "#36C3BB", "#502222", "#2B8888"}

	// UnitNameList
	// 根据名称获取汇率表信息
	UnitNameList = map[string]decimal.Decimal{}

	// UnitIdList
	// 根据Id获取汇率表信息
	UnitIdList = []unitUsdt.UnitUsdt{}

	IconMap = map[string]interface{}{}

	// RouterList
	// 基础路由表
	RouterList = map[string]api.Entity{}

	// TxTypeNameList
	// 0: "链上地址转入-正在接收"
	// 1: "链上地址转入-接收成功"
	// 2: "链上地址转入-接收失败"
	// 3: "转到链上地址-冻结"
	// 4: "转到链上地址-确认"
	// 5: "转到链上地址-失败回滚"
	TxTypeNameList = []string{"链上地址转入-正在接收", "链上地址转入-接收成功", "链上地址转入-接收失败", "转到链上地址-冻结", "转到链上地址-确认", "转到链上地址-失败回滚"}

	// BillStateList
	// 0: "充值确认中"
	// 1: "充值成功"
	// 2: "充值失败"
	// 3: "提现确认中"
	// 4: "提现成功"
	// 5: "提现失败"
	BillStateList = []string{"充值确认中", "充值成功", "充值失败", "提现确认中", "提现成功", "提现失败"}
	// BillStateColors
	// 状态颜色
	BillStateColors = map[string]string{}
)

func init() {
	InitAllData()
	HooFee = InitHooFee()
}

func InitAllData() {
	IconMap = map[string]interface{}{
		"已通过": map[string]string{
			"icon": "iconfont icon-yuanxinggou",
		},
		"已拒绝": map[string]string{
			"icon": "el-icon-close",
		},
		"已冻结": map[string]string{
			"icon": "iconfont icon-suo1",
		},
	}
	BillStateColors = map[string]string{
		"充值确认中": "warning",
		"充值成功":  "success",
		"充值失败":  "danger",
		"提现确认中": "warning",
		"提现成功":  "success",
		"提现失败":  "danger",
	}
	OrderResultColor = map[string]string{
		"待审核": "warning",
		"已通过": "success",
		"已冻结": "danger",
		"已解冻": "success",
		"已拒绝": "danger",
	}
}

func GetHooFee(chain, coin string) decimal.Decimal {
	HooFee = map[string]HooFeeInfo{}
	cache.GetRedisClientConn().Get(global.CustodyHooFee, &HooFee)
	if len(HooFee) == 0 {
		HooFee = InitHooFee()
		cache.GetRedisClientConn().Set(global.CustodyHooFee, &HooFee, time.Hour)
	}
	// 查询数据
	if coinInfo, ok := HooFee[strings.ToUpper(coin)]; ok {

		fee, err := decimal.NewFromString(coinInfo.Fee)
		if err != nil {
			return decimal.Decimal{}
		}
		// 收益：主链币 == 收费类型
		if coinInfo.FeeUnit == chain {
			return fee
		}

		findChain, err := base.FindCoinsByName(chain)
		if err != nil {
			return decimal.Decimal{}
		}
		// 是USDT类型
		if coinInfo.FeeUnit == "USDT" {
			return fee.Div(findChain.PriceUsd)
		}
		// 不是 USDT 是其他类型
		if coinInfo.FeeUnit == coin {
			// 收的是代币的数
			// 取代币
			findCoin, err := base.FindCoinsByName(coin)
			if err != nil {
				return decimal.Decimal{}
			}
			// 代币fee数*币价格 / 主链币单价
			cfee := fee.Mul(findCoin.PriceUsd)
			return cfee.Div(findChain.PriceUsd)
		}
	}
	return decimal.Decimal{}
}
