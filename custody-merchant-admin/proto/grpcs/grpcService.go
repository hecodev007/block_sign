package grpcs

import (
	"custody-merchant-admin/util"
)

type DealService struct{}

func (receiver *DealService) SendMessage(param *ParamRequest) *ParamReply {
	name := ""
	reply := &ParamReply{
		Code: 200,
		Msg:  "已经收到消息," + name,
	}
	err := util.Deserialize(param.Params["name"], &name)
	if err != nil {
		reply.Code = 400
		reply.Msg = err.Error()
		return reply
	}
	send := map[string][]byte{}
	data, err := util.Serialize("123")
	if err != nil {
		reply.Code = 500
		reply.Msg = err.Error()
		return reply
	}
	send["data"] = data
	reply.Code = 200
	reply.Msg = "已经收到消息," + name
	reply.RpcReply = send
	return reply
}
