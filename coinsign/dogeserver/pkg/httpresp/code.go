package httpresp

type HttpCode int

// 定义错误信息
const FAIL = -1
const SUCCESS = 0

var MsgFlags = map[int]string{
	SUCCESS: "ok",
	FAIL:    "fail",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return "error"
}
