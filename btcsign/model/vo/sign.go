package vo

import "github.com/group-coldwallet/btcsign/model"

//签名结果hex
type SignResult struct {
	Hex string `json:"hex"`
	model.MchInfo
}
