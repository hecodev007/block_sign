package xrputil

import (
	"encoding/json"
	"fmt"
	"github.com/go-chain/go-xrp/rpc"
	"github.com/go-chain/go-xrp/tools/http"
	"testing"
)

func TestXrpSignTx(t *testing.T) {
	//节点配置
	client := rpc.NewClient("https://s1.ripple.com:51234", "https://data.ripple.com")

	//检查地址金额
	address := "r3f7Tc7zPiaLKsmdRrAWRRQKdQu5oVCzPP"
	//res, err := client.GetAccountBalances(address, map[string]string{})
	//if err != nil {
	//	t.Error("get err: ", err)
	//}
	//for _, v := range res.Balances {
	//	fmt.Printf("balance: %+v\n", v)
	//}

	//获取账号信息，填充Sequence
	account, err := client.GetAccountInfo(address)
	if err != nil {
		t.Fatal(err.Error())
	}

	//获取节点信息填充last信息
	server, err := client.GetServerInfo()
	if err != nil {
		t.Fatal(err.Error())
	}

	//随机填tag
	//rand.Seed(time.Now().Unix())
	//tag := uint32(rand.Intn(12345))

	//指定tag
	tag := uint32(0)
	//发送模板
	tpl := &XprSignTpl{
		From:         address,
		FromPrivate:  "",
		To:           "rEcQvidASJ8Q7MnettfyyAuptGgRsu5VV1",
		AmountFloat:  "13719", //改大金额就可以假充值
		FeeFloat:     "0.000012",
		FromSequence: account.AccountData.Sequence,
		LastSequence: server.State.ValidatedLedger.Seq + 100,
		Tag:          tag,
		Currency:     "XRP",
		CoinDecimal:  6,
	}
	rawtx, err := tpl.XrpSignTx()
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println("raw tx: ", rawtx)

	//如果需要广播，就使用下面的代码
	//广播 这里没有向上抛出异常，应该自己再封装
	tx, err := client.Submit(rawtx)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println("txid: ", tx.TxJson.Hash)
}

func TestSubmit(t *testing.T) {

	rawtx := "120000228000000023000012FF24036AACE82E000012FF201B036AC13161400000E8D4A5100068400000000000000C73210268904BF3F8490CD517B450E3599F0614C0C1723A396EE79A432D1A04A02812C174463044022020E4552B7FFC19E1CE8A1CC3753D854C237C92D7152F8FAF8014D34E47DC5041022038E09EF74B72EEF9E807EB42541F2004991FD3DC5E7AB9D9476DCEEE1B76B58381148C4B960927D41767F83706065D839B7193D25B5F83146D705DEB997DEDB49DAD5857EF45A7A5FC43C905"
	res := &rpc.SubmitResp{}
	params := `{"method": "submit", "params": [{"tx_blob": "` + rawtx + `"}]}`
	resp, err := http.HttpPost("https://s1.ripple.com:51234", []byte(params))
	//resp, err := http.HttpPost("http://xrp.rylink.io:35005", []byte(params))
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("resp:%s", string(resp))
	err = json.Unmarshal(resp, res)
	if err != nil {
		t.Fatal(err.Error())
	}
	if res.Result.TxJson.Hash == "" {
		t.Fatal("广播失败")
	}
	t.Logf("txid: %s", res.Result.TxJson.Hash)

	//client.Submit()
}
