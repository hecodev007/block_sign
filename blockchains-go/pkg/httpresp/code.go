package httpresp

type HttpCode int

//常规错误：10000 - 19999
//HTTP API接口错误：20000 - 29999
// 定义错误信息
const FAIL = -1
const SUCCESS = 0
const NO_PERMISSION = 1
const PARAM_ERROR = 2 // 1-被旧程序占用1
const UN_LOGIN = 3
const INSERT_DB_FAILED = 5
const COIN_NOT_EXISTS = 10
const DATA_NOT_EXISTS = 15
const IP_NOT_ALLOWED = 332
const DATA_MPTY = 35
const PARAM_FORMAT_ERROR = 40
const REQUEST_DATA_INCOMPLETE = 45
const OUT_ORDER_ID = 46
const NOT_AMOUNT = 47
const ERROR_TOADDRESS = 48

const COIN_EMPTY = 330
const NOT_POWER = 331
const TYPE_EMPTY = 60
const TYPE_ERROR = 62
const TYPE_UNKNOWN = 65
const TO_ADDRESS_ERROR = 70
const FROM_ADDRESS_ERROR = 72
const ORDERID_ERROR = 75
const ORDERID_LENGTH = 76
const CALLBACK_LENGTH = 77
const COINTO = 78

const COINT_OUT = 301
const COIN_ERROR = 100
const NUM_ERROR = 101
const NUM_EOS_ERROR = 102
const IP_NOT_NULL = 201
const UNKNOWN_ERROR = 205
const NUM_AMOUNT_ERROR = 206

// 出库申请相关
const MERCHANT_ERROR = 103
const APPLY_ID_ERROR = 105
const STATUS_EMPTY = 110
const STATUS_ERROR = 112
const STATUS_UNKNOWN = 115

// 自动生成地址相关
const TASK_HAS_FINISH = 120
const TASK_HAS_TRADE = 125
const TASK_HAS_DELETE = 130
const TASK_HAS_NO_DEAL = 135
const TASK_ADDRESS_NUM_ERROR = 140
const TASK_ADDRESS_LEN_ERROR = 145
const HASH_MD5_ERROR = 150
const MD5_NOT_MATCH = 155
const SFROM_ERROR = 156
const MCHSTATUS_ERROR = 157
const FROZEN_ERROR = 158

const ADR_NONE = 100001
const DATA_EMPTY = 100002

// 对外提供接口
const SOURCE_FROM_ERROR = 305
const SIGNATURE_ERROR = 310

const BTC_MIN = 311
const USDT_MIN = 312
const EOS_MIN = 313
const DWC_MIN = 314
const ETH_MIN = 315
const HX_MIN = 316
const IS_NUMBER = 317
const Permission_denied = 318
const APILIST_ERROR = 319
const EMPTY_NUMBER = 320
const HC_MIN = 321
const LTC_MIN = 322

const TokenAddressEmpty = 4001
const TokenNameEmpty = 4002
const OutOrderIdEmpty = 4003
const OutOrderIdLengthError = 4004
const OutOrderIdRepeat = 4005
const CoinNameEmpty = 4006
const AmountEmpty = 4007
const FromAddressEmpty = 4008
const MemoLengthTooLong = 4009
const FeeInt = 4010
const ToAddressEmpty = 4011
const InvalidCoinName = 4012
const TransferQuantityError = 4013
const InvalidReceiveAddress = 4014
const InvalidTransferAmount = 4015
const UnsupportedToken = 4016
const LessMinQuantity = 4017
const AmountDecimalError = 4018
const MchAmountNotEnough = 4019
const UnknowError = 4020
const FrozenCoin = 4021
const SingleQuotaLimit = 4022
const ExceedingHourLimit = 4023
const ExceedingDayLimit = 4024
const DataError = 4025

//2021-03-10 write by flynn
const ValidAddressError = 4026

const ContractAddrNotAllow = 4027

var MsgFlags = map[int]string{
	SUCCESS:                 "ok",
	NO_PERMISSION:           "permission error",
	FAIL:                    "系统错误",
	PARAM_ERROR:             "参数错误",
	UN_LOGIN:                "未登录",
	INSERT_DB_FAILED:        "入库操作失败",
	COIN_NOT_EXISTS:         "该币种不存在",
	DATA_NOT_EXISTS:         "没有数据",
	IP_NOT_ALLOWED:          "IP受限",
	UNKNOWN_ERROR:           "未知错误",
	DATA_EMPTY:              "数据为空",
	PARAM_FORMAT_ERROR:      "请求参数格式错误",
	REQUEST_DATA_INCOMPLETE: "数据不全",
	NUM_AMOUNT_ERROR:        "转出金额 和 转出地址数量不一致",
	NOT_AMOUNT:              "金额不足",
	ERROR_TOADDRESS:         "接收地址不合法",
	MCHSTATUS_ERROR:         "账户未审核",
	FROZEN_ERROR:            "账户冻结",

	COIN_EMPTY:         "币种为空",
	NOT_POWER:          "没有权限",
	TYPE_EMPTY:         "类型为空",
	TYPE_ERROR:         "类型错误",
	TYPE_UNKNOWN:       "未知类型",
	TO_ADDRESS_ERROR:   "to_address错误，当前只允许一条且不能是all",
	FROM_ADDRESS_ERROR: "from_address错误，all和列表选其一",
	ORDERID_ERROR:      "订单ID错误",
	COINTO:             "账户不能为空",

	COINT_OUT:     "此币种不可申请地址",
	COIN_ERROR:    "币种不存在",
	NUM_ERROR:     "数量错误",
	NUM_EOS_ERROR: "数量错误,请输入1024倍数",
	IP_NOT_NULL:   "批次ID未提供",
	ADR_NONE:      "该商户未分配冷地址",

	MERCHANT_ERROR: "商户错误",
	APPLY_ID_ERROR: "申请ID错误",
	STATUS_EMPTY:   "状态为空",
	STATUS_ERROR:   "状态错误",
	STATUS_UNKNOWN: "状态未知",

	TASK_HAS_FINISH:        "任务已经完成",
	TASK_HAS_TRADE:         "任务已经处理过",
	TASK_HAS_DELETE:        "任务已删除",
	TASK_HAS_NO_DEAL:       "任务还未开始处理",
	TASK_ADDRESS_NUM_ERROR: "单次申请不能大于100",
	TASK_ADDRESS_LEN_ERROR: "地址长度不统一",
	HASH_MD5_ERROR:         "md5-hash值错误",
	MD5_NOT_MATCH:          "md5值不匹配",
	SFROM_ERROR:            "商户不存在",

	SOURCE_FROM_ERROR: "来源错误",
	SIGNATURE_ERROR:   "签名错误",
	OUT_ORDER_ID:      "订单重复",

	BTC_MIN:      "btc 最小单位为 0.00000546",
	DWC_MIN:      "dwc 最小单位为 0.00000546",
	USDT_MIN:     "usdt 最小单位为 0.00000001",
	EOS_MIN:      "eos 最小单位为 0.0001",
	ETH_MIN:      "eth 最小单位为 0.00000000000000000001",
	HX_MIN:       "hx 最小单位为0.00001",
	HC_MIN:       "hc 最小单位为0.001",
	LTC_MIN:      "ltc 最小单位为0.00000546",
	IS_NUMBER:    "请输入金额格式数据",
	EMPTY_NUMBER: "请输入金额",

	ORDERID_LENGTH:  "订单长度不符合 最小为11个字符,最大为64个字符",
	CALLBACK_LENGTH: "url 长度限制为200个字符",

	Permission_denied: "没有此接口权限",
	APILIST_ERROR:     "无法查到申请信息",

	TokenAddressEmpty:     "代币地址不能为空",
	TokenNameEmpty:        "代币名称不能为空",
	OutOrderIdEmpty:       "outOrderId不能为空",
	OutOrderIdLengthError: "outOrderId长度必须在11-64位",
	OutOrderIdRepeat:      "outOrderId已存在",
	CoinNameEmpty:         "coinName不能为空",
	AmountEmpty:           "amount不能为空",
	FromAddressEmpty:      "fromAddress不能为空",
	MemoLengthTooLong:     "memo不能超过200位",
	FeeInt:                "fee必须是数值",
	ToAddressEmpty:        "toAddress不能为空",
	InvalidCoinName:       "不支持的coinName",
	TransferQuantityError: "转出金额和转出地址数量不一致",
	InvalidReceiveAddress: "接收地址不合法",
	InvalidTransferAmount: "转出金额不合法",
	UnsupportedToken:      "不支持的token",
	LessMinQuantity:       "转出金额小于最小值",
	AmountDecimalError:    "转出金额最大精度错误",
	MchAmountNotEnough:    "商户余额不足",
	UnknowError:           "未知错误,请联系管理员",
	FrozenCoin:            "该币种触发风控阀值已被冻结",
	SingleQuotaLimit:      "超出单笔限额",
	ExceedingHourLimit:    "超出每小时限额",
	ExceedingDayLimit:     "超出单日限额",
	DataError:             "数据错误,请联系管理员",

	ValidAddressError:    "验证地址错误",
	ContractAddrNotAllow: "不支持合约地址",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[PARAM_ERROR]
}
