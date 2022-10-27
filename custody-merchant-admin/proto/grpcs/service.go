package grpcs

type Services interface {
	SendMessage(param *ParamRequest) *ParamReply
}
