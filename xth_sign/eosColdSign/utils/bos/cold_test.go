package bos

import (
	"bosSign/common/validator"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/eoscanada/eos-go/ecc"

	"github.com/eoscanada/eos-go"
)

func Test_time(t *testing.T) {
	t.Log(time.Now().Unix())
	tm := time.Unix(1641454454, 0)
	t.Log(tm.String())
	js, _ := tm.MarshalJSON()
	t.Log(string(js))

	//"2022-01-05 20:44:21 +0800 CST"
	//	"2021-05-06T02:09:34.634"
}
func Test_sign(t *testing.T) {
	wifKey := "5JFnwrLsvo6nmCRPQ2636U2zygHZ9nj2YrNHh5WrTyxC4vwJ9q7"
	//paramstr := "{\"account\":\"tethertether\",\"actor\":\"eoshoowallet\",\"chain_id\":\"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906\",\"coinName\":\"eos\",\"data\":\"90558c8653da30551082082188c0a6f1cd2416000000000004555344540000000c79616e673132313131323131\",\"eos_code\":\"tethertether\",\"expiration\":\"2022-01-05 20:44:21 +0800 CST\",\"hash\":\"ed3de7f0568659f1b94a2a77ef873ca2\",\"mchId\":\"hoo\",\"orderId\":\"HsgwHy8+UvfSepHNISPmcVxRVzjJJRa7bWmab9UYNEw+0WbwUQ==_1620237362\",\"public_key\":\"EOS6bJAf3BZUmjR6pgegtjXGJtrQzycysvQemvMftxQQNGP8W8Gis\",\"ref_block_num\":182214239,\"ref_block_prefix\":1370158199}"
	paramstr := "{\"account\":\"eosio.token\",\"actor\":\"xutonghua111\",\"chain_id\":\"d5a3d18fbb3c084e3b1f3fa98c21014b5f3db536cc15d08f9f6479517c6a3d86\",\"coinName\":\"eos\",\"data\":\"a592d5618c8b8dc7ce07000000000100a6823403ea3055000000572d3ccdcd01104230bab149b3ee00000000a8ed323228104230bab149b3ee90558c8653da303d40420f000000000004454f5300000000073437393639383600\",\"eos_code\":\"eosio.token\",\"expiration\":\"2022-01-05 20:44:21 +0800 CST\",\"hash\":\"ed3de7f0568659f1b94a2a77ef873ca2\",\"mchId\":\"hoo\",\"orderId\":\"HsgwHy8+UvfSepHNISPmcVxRVzjJJRa7bWmab9UYNEw+0WbwUQ==_1620237362\",\"public_key\":\"EOS6VeUZo93nzcmhK3HfQaXBsiw9tsd6hPfU2QwS2adpYQqM9G2Rt\",\"ref_block_num\":35724,\"ref_block_prefix\":130992013}"
	params := &validator.ColdSign{}
	err := json.Unmarshal([]byte(paramstr), params)
	if err != nil {
		t.Fatal(err.Error())
	}

	chainid, _ := hex.DecodeString(params.ChainID)
	payload, _ := hex.DecodeString(params.Data)
	//hash, _ := hex.DecodeString(params.Hash)
	//t.Log(string(hash))
	digest := eos.SigDigest(chainid, payload, []byte{})
	t.Log(hex.EncodeToString(digest))
	privKey, err := ecc.NewPrivateKey(wifKey)
	if err != nil {
		t.Fatal(err.Error())
	}
	sigrature, err := privKey.Sign(digest)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(sigrature.String())
}
