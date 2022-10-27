package cring

import (
	"encoding/json"
	"fmt"
	stafirpc "github.com/JFJun/stafi-substrate-go/client"
	"github.com/coldwallet-group/bifrost-go/client"
	"github.com/coldwallet-group/substrate-go/rpc"

	"math/big"
	"testing"
)
func Test_dot0(t *testing.T){
	cli,err := rpc.New("http://13.114.44.225:30993","","")
	if err != nil {
		t.Fatal(err.Error())
	}
	b,err :=cli.GetBlockByNumber(5202217)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(b))
}
func Test_dot(t *testing.T){
	cli,err := stafirpc.New("http://13.114.44.225:30993")
	if err != nil {
		t.Fatal(err.Error())
	}
	b,err :=cli.GetBlockByNumber(5202217)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(b))
}
func Test_dot2(t *testing.T){
	cli,err := client.New("http://13.114.44.225:30993")
	if err != nil {
		t.Fatal(err.Error())
	}
	b,err :=cli.GetBlockByNumber(5202217)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(b))
}
func Test_info3(t *testing.T){
	cli,err := client.New("http://13.114.44.225:31933")
	if err != nil {
		t.Fatal(err.Error())
	}
	info,err :=cli.GetAccountInfo("5Hn8xoicWvpJv49yjwm1X7Z5UcHxj1Znj817W76q9WtfGko7")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(cli.ChainName)
	t.Log(String(info))
	b,err :=cli.GetBlockByNumber(5317088)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(b))
}

func Test_info2(t *testing.T){
	cli,err := rpc.New("http://13.114.44.225:31933","","")
	if err != nil {
		t.Fatal(err.Error())
	}
	info,err :=cli.GetAccountInfo("5Hn8xoicWvpJv49yjwm1X7Z5UcHxj1Znj817W76q9WtfGko7")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(cli.CoinType)
	t.Log(string(info))
	b,err :=cli.GetBlockByNumber(5317088)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(b))
}

func Test_info(t *testing.T){
	cli,err := New("http://13.114.44.225:31933","","")
	if err != nil {
		t.Fatal(err.Error())
	}
	info,err :=cli.GetAccountInfo("5Hn8xoicWvpJv49yjwm1X7Z5UcHxj1Znj817W76q9WtfGko7")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(cli.CoinType)
	t.Log(String(info))
	//130812796,016279090
	//57000439187762514493440000001
	//241307016967867274577816276061650945
	//161267610891257680
	//297486249345682649016519743116410881
}
func Test_byte(t *testing.T){
//[143 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 5 211 220 191 90 121 46 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	a :=big.NewInt(13081279601627909)
	fmt.Println(a.Bytes())
	fmt.Println([]byte("13081279601627909000000000"))
	//49514856495055754484954505557485748
	//241307016967867274577816276061650945
}
func String(d interface{}) string{
	str,_:= json.Marshal(d)
	return string(str)
}
type AccountInfo struct {
	Nonce    uint64 `json:"nonce"`
	RefCount uint8  `json:"ref_count"`
	Data     struct {
		Free       big.Int `json:"free"`
		Reserved   uint64 `json:"reserved"`
		MiscFrozen uint64 `json:"misc_frozen"`
		FreeFrozen uint64 `json:"free_frozen"`
	} `json:"data"`
}