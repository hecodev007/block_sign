package models

type AddressInput struct {
	TagName string `json:"tagName"`
}

type BatchAddressInput struct {
	//TagName string `json:"tagName"`
	Num int64 `json:"num"`
}
