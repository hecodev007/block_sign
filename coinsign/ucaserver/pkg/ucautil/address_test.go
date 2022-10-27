package ucautil

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"testing"
)

func TestAddress(t *testing.T) {
	wif, _ := CreatePrivateKey()
	//wif, _ := ImportWIF("your compressed privateKey Wif")
	//wif, _ := ImportWIF("L3sy1eqQJKkz2kjShUukLpS1zGd2ZoGrBz9uupiP7sUYWhu44KVg")

	address, _ := GetAddress(wif)
	t.Log("Common Address:", address.EncodeAddress())
	t.Log("PrivateKeyWifCompressed:", wif.String())

	addressStaking, _ := GetStakingAddress(wif)
	t.Log("addressStaking Address:", addressStaking.EncodeAddress())
	t.Log("PrivateKeyWifCompressed:", wif.String())
}

func TestAddrByPrivkey(t *testing.T) {
	//UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19
	//VT52WZEzwbfUZUrGu114PtZP3fKxDydb4wzhAQ61x1JwyFAqT1ik
	//76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac
	//02d53547cefcd44f108ceaca86225145f4c28e8c8f5394564a155f90f140d7f32e
	//{"result":{"isvalid":true,"address":"UZ5i1otQ3bxsAmeogTciRsXtB9Nf4KpE6k","scriptPubKey":"76a91479bce72d49660fba53484b97ef4fa1adb12730ba88ac","ismine":true,"isstaking":false,"iswatchonly":false,"isscript":false,"pubkey":"0239dfd45c88a3a25850f56d3c485538b4fd6c94420f6e28b0312dcbce394423ee","iscompressed":true,"account":"test"},"error":null,"id":11}

	//VXEMBieFtW8mP1ce6opJyaasNoTiGnZB29NinkF3UZSu4rgWzSKn
	//UZ5i1otQ3bxsAmeogTciRsXtB9Nf4KpE6k
	//0239dfd45c88a3a25850f56d3c485538b4fd6c94420f6e28b0312dcbce394423ee
	//0239dfd45c88a3a25850f56d3c485538b4fd6c94420f6e28b0312dcbce394423ee

	//解析
	wif2, _ := ImportWIF("VT52WZEzwbfUZUrGu114PtZP3fKxDydb4wzhAQ61x1JwyFAqT1ik")
	address2, _ := GetAddress(wif2)
	t.Log("解析后地址:", address2.EncodeAddress())
	t.Log("解析后地址公钥:", hex.EncodeToString(address2.PubKey().SerializeCompressed()))
	addr, _ := btcutil.DecodeAddress(address2.EncodeAddress(), params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	t.Log("解析后脚本公钥:", hex.EncodeToString(pkScript))

	//CWLYiNzUA41ZWmvi1GmmtxEaGft7q3nWUd
	//VQapdf2LZ58UxzgPg7s9efgRQegYZgKA2nW7s4qEjmyNvZJqH91Z
	//{"result":{"isvalid":true,"address":"CWLYiNzUA41ZWmvi1GmmtxEaGft7q3nWUd","scriptPubKey":"76a914982e462b1593d526a915be9aa98ba3fd1f4bae5d88ac","ismine":true,"isstaking":true,"iswatchonly":false,"isscript":false,"pubkey":"03f1fcc97737830234514863af48532a90726a5f2f6cdf6603e2ae7d7f25292db9","iscompressed":true,"account":""},"error":null,"id":11}

}
