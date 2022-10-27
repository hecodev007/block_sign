package crust

// import (
// 	"crustsign/common/log"
// 	"encoding/json"
// 	"testing"

// 	"github.com/JFJun/bifrost-go/client"
// )

// func Test_RPC(t *testing.T) {
// 	c, err := NewClient("http://13.114.44.225:30943")
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	}
// 	// info, err := c.GetAccountInfo("rB2xctiQMNP1isAWcqiMqbVeF9TAmQWeGd7mu5erqFgJAff")
// 	// t.Log(info, info.Data.Free.String())
// }
// func Test_submit(t *testing.T) {
// 	cli, err := client.New("wss://karura-rpc-1.aca-api.network")
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	}
// 	block, err := cli.GetBlockByNumber(330000)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	}

// 	t.Log(String(block))
// 	var result interface{}
// 	rawtx := "0x41028400583ebc2ba0ff987dea9d214991fc42af73f613b141d364265917a91d7b0f6a1e016a92b1c0cffd41d967fec26a884702198b2b333ccb1b0028cedc065996a37746f93c0f081be75b26c76559884c34f3ab100d0df7fac6c892013649896ff4ef810008000a000051bfc148fc6c0028fc58389368857d89bc2325f1966e62f36c3fb7df8d0e24800b200c45bd2745"
// 	err = cli.C.Client.Call(&result, "author_submitExtrinsic", rawtx)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	t.Log(String(result))
// }

// func String(d interface{}) string {
// 	str, _ := json.Marshal(d)
// 	return string(str)
// }
