package models

type AddrCheck struct {
	MchInfo
	Addresses []string `json:"addresses"` //地址数组
	Hash      string   `json:"hash"`      //hash值
}
