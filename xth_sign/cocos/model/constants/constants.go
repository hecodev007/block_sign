package constants

const (
	Http200              = `{"status":200,"message":"OK"}`
	HttpRequestError     = `{"status":1001,"message":"Http request failed"}`
	HttpRequestErrorCode = 1001
)

const (
	JsonUnmarshalErrorCode = 2001
	JsonUnmarshalError     = `{"code":2001,"body":"","error":"Json Unmarshal failure"}`
	JsonMarshalErrorCode   = 2002
	JsonMarshalError       = `{"code":2002,"body":"","error":"Json Marshal failure"}`
)
