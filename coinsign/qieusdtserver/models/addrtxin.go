package models

type AddrTxin struct {
	Gas           int64    `json:"gas"`
	Addrs         []string `json:"addrs"`         //地址序列
	Toaddr        string   `json:"toaddr"`        //目标地址
	Changeaddress string   `json:"changeaddress"` //找零地址
}

type AddrTxinUseFee struct {
	Gas           int64  `json:"gas"`
	Addr          string `json:"addr"`          //地址序列
	Toaddr        string `json:"toaddr"`        //目标地址
	Changeaddress string `json:"changeaddress"` //找零地址
	Feeaddress    string `json:"feeaddress"`    //手续费地址
}
