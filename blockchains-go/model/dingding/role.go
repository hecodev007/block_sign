package dingding

import "strings"

type DingRole string

const (
	DingAdmin     DingRole = "admin"     // 管理员
	DingDeveloper DingRole = "developer" // 开发人员
	DingCustomer  DingRole = "customer"  // 客服
	DingTmp       DingRole = "tmp"       // 临时人员
	DingOperation DingRole = "operation" // 运营
	Guest         DingRole = "guest"     // 访客
)

type DingCommand string

const (
	DING_SET_PRIORITY_ORDER     DingCommand = "设置优先订单"
	DING_CANCEL_PRIORITY_ORDER  DingCommand = "取消优先订单"
	DING_ORDER_TX_LINK          DingCommand = "订单交易关联"
	DING_REPLACE_FAILURE_TXS    DingCommand = "替换失败交易" // 替换失败交易
	DING_CANCEL_TXS             DingCommand = "取消交易"   // 取消交易
	DING_FORCE_CANCEL_TXS       DingCommand = "强制取消交易" // 取消交易
	DING_OUT_COLLECT            DingCommand = "出账归集"   // 出账归集
	DING_VERIFY                 DingCommand = "审核"     // 审核订单
	DING_REPUSH_ORDER           DingCommand = "重推"     // 重推订单
	DING_DISCARD_REPUSH_ORDER   DingCommand = "废弃重推"   // 废弃重推
	DING_DISCARD_ROLLBACK_ORDER DingCommand = "废弃回滚"   // 废弃回滚
	DING_CANCEL_ORDER           DingCommand = "取消"     // 取消审核状态订单
	DING_CHECK_ORDER            DingCommand = "检查"     // 取消订单
	DING_ROLLBACK_ORDER         DingCommand = "回滚"     // 回滚订单 暂时无用
	//DING_ABANDONED_ORDER         DingCommand = "废弃"      // 废弃交易 状态30 广播4 txid 存在 重置为 状态49  广播10
	DING_MERGE_ORDER           DingCommand = "合并"      // 冷地址合并，针对账户模型
	DING_FAIL_CHAIN            DingCommand = "链上失败回滚"  // 冷地址合并，针对账户模型
	DING_SHOW_ADDR             DingCommand = "列举地址"    // 列举币种金额以及地址类型，目前是查询这个商户币种的前十个有钱地址金额
	DING_RECYCLE_COIN          DingCommand = "零散回收"    // 列举币种金额以及地址类型，目前是查询这个商户币种的前十个有钱地址金额
	DING_RESET_ETH             DingCommand = "重置ETH"   // 废弃链上交易
	DING_ETH_GAS               DingCommand = "ETH-GAS" // 重新设置gas
	DING_ETH_CLOSE_ALLCOLLECT  DingCommand = "关闭所有归集"  // 关闭归集
	DING_ETH_CLOSE_COLLECT     DingCommand = "关闭归集"    // 关闭归集
	DING_ETH_OPEN_COLLECT      DingCommand = "开启归集"    // 开启归集 json
	DING_DOT_RECYCLE           DingCommand = "DOT回收"
	DING_DHX_RECYCLE           DingCommand = "DHX回收"
	DING_BTM_RECYCLE           DingCommand = "BTM回收"
	DING_CKB_RECYCLE           DingCommand = "CKB回收"
	DING_ETH_FEE               DingCommand = "ETH手续费"
	DING_ORDER_COLLECT         DingCommand = "订单归集"
	DING_COLLECT               DingCommand = "归集"
	DING_ETH_LIST_AMOUNT       DingCommand = "列举金额"
	DING_ETH_COLLECT_TOKEN     DingCommand = "ETH代币归集"
	DING_ETH_INTERNAL          DingCommand = "ETH内部转账"
	DING_ETH_RESET_NONCE       DingCommand = "ETH重置NONCE"
	DING_CHAIN_REPUSH          DingCommand = "补数据"
	DING_CHAIN_FORCE_REPUSH    DingCommand = "强制补数据"
	DING_FIX_BALANCE           DingCommand = "纠正余额"
	DING_DEL_KEY               DingCommand = "清除key"
	DING_REFRESH_KEY           DingCommand = "加币刷新"
	DING_MAIN_CHAIN_REFRESH    DingCommand = "主链刷新"
	DING_XRP_SUPPLEMENTAL      DingCommand = "XRP补充"
	DING_BTC_RECYCLE           DingCommand = "BTC零散回收"
	DING_BTC_MERGE             DingCommand = "BTC合并"
	DING_FIX_ADDR_AMOUNT_ALL0C DingCommand = "分配固定地址金额"
	DING_SAME_WAY_BACK         DingCommand = "设置交易回退"

	// DING_COIN_FEE   	      DingCommand = "打手续费"
	// DING_COIN_COLLECT_TOKEN	  DingCommand =  "代币归集"
	// DING_FIND_ADDRESS_FEE	  DingCommand = "查看地址手续费"

)

func (d DingCommand) ToString() string {
	return string(d)
}

// 人员关系绑定
type DingUser struct {
	RoleName DingRole // 角色
	Desc     string   // 人员描述
}

type DingRoleAuth struct {
	Coins    []string      // 币种权限，一级判断
	Commands []DingCommand // 命令权限 二级判断
}

// 查询角色是否拥有权限
func (auth *DingRoleAuth) HaveCommand(commandName string) (bool, DingCommand) {
	commandName = strings.TrimSpace(commandName)
	for _, cm := range auth.Commands {
		// 首先解析命令
		if strings.HasPrefix(commandName, cm.ToString()) {
			return true, cm
		}
	}
	return false, ""
}

func (auth *DingRoleAuth) HaveCoin(coinName string) bool {
	for _, coin := range auth.Coins {
		if strings.ToLower(coin) == strings.ToLower(coinName) {
			return true
		}
	}
	return false
}
