package dingding

import (
	model "github.com/group-coldwallet/blockchains-go/model/dingding"
)

// key = 钉钉ID
// value 对应角色
var DingUsers map[string]model.DingUser

// 角色权限
var DingRoles map[model.DingRole]*model.DingRoleAuth

// 由于该部分功能,操作权限较高，暂时不读取配置
var coins = []string{"btc", "klay", "eth", "gxc", "eos", "seek",
	"usdt", "etc", "kava", "luna", "lunc", "bnb", "rub", "wbc", "cds", "ar", "hx",
	"zvc", "cocos", "mdu", "fo", "ont", "cocos", "ar", "ksm", "bnc", "hnt",
	"crab", "vet", "bsv", "uca", "celo", "mtr", "fio", "qtum", "mtr",
	"ltc", "sol", "tlos", "pcx", "ghost", "dot", "azero", "sgb-sgb", "kar", "bch", "zec", "ckb",
	"btm", "hc", "stx", "dcr", "nas", "doge", "avax", "bsc", "dash", "fil", "wd",
	"near", "oneo", "atom", "biw", "near", "yta", "cfx", "star", "fis", "atp", "cph-cph",
	"xlm", "bcha", "xec", "waves", "ada", "trx", "zen", "mw", "dip", "algo", "ori", "bos", "okt", "glmr", "avaxcchain", "heco", "stx", "nyzo", "xdag", "iost", "hsc", "dhx", "dom", "wtc", "moac", "satcoin", "eac", "iota", "rbtc", "movr", "sep20", "ccn", "optim", "brise-brise", "ftm", "welups", "rose", "one", "rev", "tkm", "ron", "kai", "neo", "icp", "flow", "uenc", "cspr", "matic-matic",
	"crust", "waxp", "iotx", "rei", "evmos", "aur", "dscc", "mob", "dscc1", "deso", "lat", "nodle", "hbar", "steem",
}

var guestCommands = []model.DingCommand{
	model.DING_ETH_LIST_AMOUNT,
	model.DING_CHECK_ORDER,

	//model.DING_ORDER_TX_LINK,
	//model.DING_REPLACE_FAILURE_TXS,
	//model.DING_CANCEL_TXS,
	//model.DING_FORCE_CANCEL_TXS,
	//model.DING_COLLECT,
	//model.DING_ORDER_COLLECT,
	//model.DING_SAME_WAY_BACK,
	//model.DING_OUT_COLLECT,
	//model.DING_FAIL_CHAIN,
	//model.DING_VERIFY,
	//model.DING_REPUSH_ORDER,
	//model.DING_DISCARD_REPUSH_ORDER,
	//model.DING_DISCARD_ROLLBACK_ORDER,
	//model.DING_CANCEL_ORDER,
	//model.DING_ROLLBACK_ORDER,
	////model.DING_ABANDONED_ORDER,
	//model.DING_MERGE_ORDER,
	//model.DING_SHOW_ADDR,
	//model.DING_RECYCLE_COIN,
	//model.DING_RESET_ETH,
	//model.DING_ETH_GAS,
	//model.DING_ETH_CLOSE_COLLECT,
	//model.DING_ETH_OPEN_COLLECT,
	//model.DING_ETH_CLOSE_ALLCOLLECT,
	//model.DING_DOT_RECYCLE,
	//model.DING_DHX_RECYCLE,
	//model.DING_BTM_RECYCLE,
	//model.DING_CKB_RECYCLE,
	//model.DING_ETH_FEE,
	//model.DING_ETH_COLLECT_TOKEN,
	//model.DING_ETH_RESET_NONCE,
	//model.DING_CHAIN_REPUSH,
	//model.DING_CHAIN_FORCE_REPUSH,
	//model.DING_ETH_INTERNAL,
	//model.DING_FIX_BALANCE,
	//model.DING_DEL_KEY,
	//model.DING_REFRESH_KEY,
	//model.DING_XRP_SUPPLEMENTAL,
	//model.DING_BTC_MERGE,
	//model.DING_BTC_RECYCLE,
	//model.DING_MAIN_CHAIN_REFRESH,
	//model.DING_FIX_ADDR_AMOUNT_ALL0C,
}

var adminCommands = []model.DingCommand{
	model.DING_ORDER_TX_LINK,
	model.DING_REPLACE_FAILURE_TXS,
	model.DING_CANCEL_TXS,
	model.DING_FORCE_CANCEL_TXS,
	model.DING_COLLECT,
	model.DING_ORDER_COLLECT,
	model.DING_SAME_WAY_BACK,
	model.DING_OUT_COLLECT,
	model.DING_FAIL_CHAIN,
	model.DING_VERIFY,
	model.DING_REPUSH_ORDER,
	model.DING_DISCARD_REPUSH_ORDER,
	model.DING_DISCARD_ROLLBACK_ORDER,
	model.DING_CANCEL_ORDER,
	model.DING_CHECK_ORDER,
	model.DING_ROLLBACK_ORDER,
	//model.DING_ABANDONED_ORDER,
	model.DING_MERGE_ORDER,
	model.DING_SHOW_ADDR,
	model.DING_RECYCLE_COIN,
	model.DING_RESET_ETH,
	model.DING_ETH_GAS,
	model.DING_ETH_CLOSE_COLLECT,
	model.DING_ETH_OPEN_COLLECT,
	model.DING_ETH_CLOSE_ALLCOLLECT,
	model.DING_DOT_RECYCLE,
	model.DING_DHX_RECYCLE,
	model.DING_BTM_RECYCLE,
	model.DING_CKB_RECYCLE,
	model.DING_ETH_FEE,
	model.DING_ETH_LIST_AMOUNT,
	model.DING_ETH_COLLECT_TOKEN,
	model.DING_ETH_RESET_NONCE,
	model.DING_CHAIN_REPUSH,
	model.DING_CHAIN_FORCE_REPUSH,
	model.DING_ETH_INTERNAL,
	model.DING_FIX_BALANCE,
	model.DING_DEL_KEY,
	model.DING_REFRESH_KEY,
	model.DING_XRP_SUPPLEMENTAL,
	model.DING_BTC_MERGE,
	model.DING_BTC_RECYCLE,
	model.DING_MAIN_CHAIN_REFRESH,
	model.DING_FIX_ADDR_AMOUNT_ALL0C,
	// model.DING_COIN_FEE,
	// model.DING_COIN_COLLECT_TOKEN,
	// model.DING_FIND_ADDRESS_FEE,
}

var developerCommands = []model.DingCommand{
	model.DING_ORDER_TX_LINK,
	model.DING_COLLECT,
	model.DING_ORDER_COLLECT,
	model.DING_SAME_WAY_BACK,
	model.DING_CANCEL_TXS,
	model.DING_VERIFY,
	model.DING_REPUSH_ORDER,
	model.DING_CANCEL_ORDER,
	model.DING_CHECK_ORDER,
	model.DING_SHOW_ADDR,
	model.DING_DOT_RECYCLE,
	model.DING_MERGE_ORDER,
	model.DING_BTM_RECYCLE,
	model.DING_CKB_RECYCLE,
	model.DING_RECYCLE_COIN,
	// 2020-09-28 write by flynn
	model.DING_ETH_FEE,
	model.DING_ETH_LIST_AMOUNT,
	model.DING_ETH_COLLECT_TOKEN,
	model.DING_ETH_RESET_NONCE,
	model.DING_CHAIN_REPUSH,
	model.DING_ETH_CLOSE_COLLECT,
	model.DING_ETH_OPEN_COLLECT,
	model.DING_ETH_CLOSE_ALLCOLLECT,
	model.DING_FIX_BALANCE,
	model.DING_DEL_KEY,
	model.DING_DHX_RECYCLE,
	model.DING_BTC_MERGE,
	model.DING_BTC_RECYCLE,
	model.DING_REFRESH_KEY,
	model.DING_MAIN_CHAIN_REFRESH,
	model.DING_XRP_SUPPLEMENTAL,
	// model.DING_COIN_FEE,
	// model.DING_COIN_COLLECT_TOKEN,
	// model.DING_FIND_ADDRESS_FEE,
}

var customerCommands = []model.DingCommand{
	model.DING_VERIFY,
	model.DING_RECYCLE_COIN,
	model.DING_CHAIN_REPUSH,
}

// 假期临时权限
var tmpCommands = []model.DingCommand{
	model.DING_REPUSH_ORDER,
	model.DING_CHECK_ORDER,
	model.DING_CHAIN_REPUSH,
	model.DING_ETH_OPEN_COLLECT,
}

// 运营人员权限
var operationCommands = []model.DingCommand{
	model.DING_MAIN_CHAIN_REFRESH,
}

func InitDingRols(modelType string) {
	// 权限初始化
	DingRoles = map[model.DingRole]*model.DingRoleAuth{
		model.DingAdmin: &model.DingRoleAuth{
			Coins:    coins,
			Commands: adminCommands,
		},
		model.DingDeveloper: &model.DingRoleAuth{
			Coins:    coins,
			Commands: developerCommands,
		},
		model.DingCustomer: &model.DingRoleAuth{
			Coins:    coins,
			Commands: customerCommands,
		},
		model.DingTmp: &model.DingRoleAuth{
			Coins:    coins,
			Commands: tmpCommands,
		},
		model.DingOperation: &model.DingRoleAuth{
			Coins:    coins,
			Commands: operationCommands,
		},
		model.Guest: {
			Coins:    coins,
			Commands: guestCommands,
		},
	}

	// 角色初始化
	DingUsers = make(map[string]model.DingUser, 0)
	//DingUsers["$:LWCP_v1:$r1zfb9PtSpuH/4dYeiHu1A=="] = model.DingUser{
	//	RoleName: model.DingAdmin,
	//	Desc:     "zhuwenjian",
	//}

	DingUsers["guest"] = model.DingUser{
		RoleName: model.Guest,
		Desc:     "guest",
	}
	DingUsers["$:LWCP_v1:$7T+wAtaQ0scPArIXzp39Ng=="] = model.DingUser{
		RoleName: model.DingAdmin,
		Desc:     "lijiayi",
	}
	DingUsers["$:LWCP_v1:$QQJ1og8I3pNk0fDTiIHRPA=="] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "xutonghua",
	}
	DingUsers["$:LWCP_v1:$/sYvbj+f8dM4l3ALrh28Gw=="] = model.DingUser{
		RoleName: model.DingAdmin,
		Desc:     "chenyuanjian",
	}
	DingUsers["$:LWCP_v1:$4OCj2eOh4UE0z14IWHcMqIm4kE92GOvh"] = model.DingUser{
		RoleName: model.DingAdmin,
		Desc:     "huangjunheng",
	}
	DingUsers["$:LWCP_v1:$3k27hVBEfVcWDt+YhyZ/nZirSGhKNPFn"] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "yanyaoqiang",
	}
	DingUsers["$:LWCP_v1:$W52mWoHHxAEtchOMkhPkCw=="] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "yaoxinrui",
	}
	DingUsers["$:LWCP_v1:$W1Q5Nabj/1HsBWN3ITG6HGd4yr+SkvYR"] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "sunyinchong",
	}
	DingUsers["$:LWCP_v1:$oT5lV7sQE0xhN+DT5t5+8w=="] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "qiuxingxing",
	}
	DingUsers["$:LWCP_v1:$TnxdiGFMqIxvFAtyZ/Q2dryq+wnguT//"] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "ouyangfan",
	}
	DingUsers["$:LWCP_v1:$m4bsmSG4ZazIz8q33XqmH0+LtczdRGLg"] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "wanggang",
	}
	DingUsers["$:LWCP_v1:$JFobZb1T/f6h1JKdrW52Yz38o3jHS2/t"] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "zhangzhaoyi",
	}
	DingUsers["$:LWCP_v1:$CdkiRa1VnkZQ3pk+7zLjoA=="] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "liaorongwen",
	}
	DingUsers["$:LWCP_v1:$5etHtZy6z/vfbdMnLfpScg=="] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "sunjiaqi",
	}
	DingUsers["$:LWCP_v1:$8s2CD21PlxutXqNHIDz4IpmnV7Svx22T"] = model.DingUser{
		RoleName: model.DingDeveloper,
		Desc:     "kevin",
	}
	DingUsers["$:LWCP_v1:$0RWCo5ooqnVBFOb6nmKdFQ=="] = model.DingUser{
		RoleName: model.DingOperation,
		Desc:     "Macin",
	}
	DingUsers["$:LWCP_v1:$rf3iruKz4L3qXkcnx1dmWOX6NXBFQXV7"] = model.DingUser{
		RoleName: model.DingOperation,
		Desc:     "Vincent",
	}

	if modelType != "prod" {
		customerCommands = append(customerCommands, model.DING_REPUSH_ORDER)
	}
}
