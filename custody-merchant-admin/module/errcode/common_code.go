package errcode

// 公共错误代码
var (
	Success                   = NewError(0, "成功")
	ServerError               = NewError(100001, "服务内部错误")
	InvalidParams             = NewError(100002, "入参错误")
	NotFound                  = NewError(100003, "找不到")
	ParamSignError            = NewError(100009, "参数签名错误")
	UnauthorizedAuthNotExist  = NewError(100004, "鉴权失败,找不到对应的AppKey和AppSecret")
	UnauthorizedTokenError    = NewError(100005, "鉴权失败,Token错误")
	UnauthorizedTokenTimeout  = NewError(100006, "鉴权失败,Token超时")
	UnauthorizedTokenGenerate = NewError(100007, "鉴权失败,Token生成失败")
	UnauthorizedAuthFail      = NewError(100009, "权限验证不通过")
	TooManyRequests           = NewError(100008, "请求太多")

	// ErrorCasbinCreateFail 权限管理错误码
	ErrorCasbinCreateFail = NewError(601001, "创建权限规则失败")
	ErrorCasbinUpdateFail = NewError(601002, "更新权限规则失败")
	ErrorCasbinListFail   = NewError(601003, "获取权限规则列表失败")
)
