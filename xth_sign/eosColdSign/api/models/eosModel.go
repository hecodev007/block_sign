package models

//from:https://bos.eosn.io/v1/chain/get_info
//mainet
var CHAINID = []byte("\xd5\xa3\xd1\x8f\xbb\x3c\x08\x4e\x3b\x1f\x3f\xa9\x8c\x21\x01\x4b\x5f\x3d\xb5\x36\xcc\x15\xd0\x8f\x9f\x64\x79\x51\x7c\x6a\x3d\x86")

type EosModel struct{}

//获取私钥
func (m *EosModel) GetPrivate(mchName, pubkey string) (private string, err error) {
	//todo:注释
	if pubkey == "EOS6VeUZo93nzcmhK3HfQaXBsiw9tsd6hPfU2QwS2adpYQqM9G2Rt" {
		return "5JFnwrLsvo6nmCRPQ2636U2zygHZ9nj2YrNHh5WrTyxC4vwJ9q7", nil
	}
	return "", nil
}
