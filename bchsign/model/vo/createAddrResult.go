package vo

import "github.com/group-coldwallet/bchsign/model"

type CreateAddrResult struct {
	model.MchInfo
	Num   int               `json:"num"` //数量
	Addrs map[string]string `json:"address"`
}
