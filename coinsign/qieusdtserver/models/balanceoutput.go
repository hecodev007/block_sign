package models

type BalanceOutput struct {
	Address    string              `json:"address"`    //地址
	Propertyid int                 `json:"propertyid"` //代币编号
	Result     ClientBalanceResult `json:"result"`     //返回结果
	Error      string              `json:"error"`      //错误结果
	Id         uint64              `json:"id"`
}

type ClientBalanceResult struct {
	Balance  string `json:"balance"`  //余额
	Reserved string `json:"reserved"` //the amount reserved by sell offers and accepts
	Frozen   string `json:"frozen"`   //冻结金额
}
