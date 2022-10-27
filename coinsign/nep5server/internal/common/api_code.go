package common

const (
	SUCCESS         = 0
	ERROR           = 1
	PAGE_NOT_FOUND  = 404
	ERROR_JSON_BODY = 10000
	ERROR_COIN      = 10001
	ERROR_AMOUNT    = 10002
)

var MsgFlags = map[int]string{
	SUCCESS:         "ok",
	ERROR:           "fail",
	PAGE_NOT_FOUND:  "page not found",
	ERROR_JSON_BODY: "error json body",
	ERROR_COIN:      "error coin",
	ERROR_AMOUNT:    "error amount",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[ERROR]
}
