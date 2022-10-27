package httpresp

const FAIL = -1
const SUCCESS = 0
const NO_PERMISSION = 1
const PARAM_ERROR = 2
const INSERT_ADDRESS_ERROR = 3
const INSERT_CONTRACT_ERROR = 4
const DELETE_ADDRESS_ERROR = 5
const DELETE_CONTRACT_ERROR = 6
const UPDATA_ADDRESS_ERROR = 7

type HttpCode int

var MsgFlags = map[int]string{
	SUCCESS:               "ok",
	FAIL:                  "系统错误",
	NO_PERMISSION:         "permission error",
	PARAM_ERROR:           "参数错误",
	INSERT_ADDRESS_ERROR:  "插入地址错误",
	INSERT_CONTRACT_ERROR: "插入合约地址错误",
	DELETE_ADDRESS_ERROR:  "删除地址错误",
	DELETE_CONTRACT_ERROR: "删除合约地址错误",
	UPDATA_ADDRESS_ERROR:  "更新地址错误",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[PARAM_ERROR]
}
