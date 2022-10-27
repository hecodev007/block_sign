package repush

//{"uid":5,"txid":"0x46606386830383faff5253b332c6af73444e0009b9678f7761d4f49564de5cf3","coin":"eth"}
type DingRepush struct {
	Uid        uint   `json:"uid"`
	Txid       string `json:"txid"`
	Coin       string `json:"coin"`
	Height     uint   `json:"height"`
	IsInternal bool   `json:"isInternal"`
}
