package service

import "github.com/group-coldwallet/blockchains-go/model/merge"

//账户模型多个地址出账的时候，单个地址不足以满足出账的时候，手动合并出账
type MergeService interface {
	//msg,执行结果
	MergeCoin(params merge.MergeParams) (msg string, err error)
}
