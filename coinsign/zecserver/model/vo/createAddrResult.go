package vo

import "github.com/group-coldwallet/zecserver/model"

type CreateAddrResult struct {
	model.MchInfo
	Num   int      `json:"num"` //数量
	Addrs []string `json:"address"`
}
