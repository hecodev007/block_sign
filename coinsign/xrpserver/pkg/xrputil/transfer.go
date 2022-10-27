package xrputil

import (
	"errors"
	"github.com/go-chain/go-xrp/rpc"
	"xrpserver/common/log"
)

var accounts map[string]string

func init(){
	accounts = make(map[string]string,0)
	//测试私钥
	//accounts["rM43k4sCCR9ipy8igUmjtfZk2wNota4y4F"]="a930c133cc32fb247d09ce211b4b0941c8c7fdb1869fdddad29187d294724422"

}

func Transfer(from,amount string)(txid string,err error){
	if _,ok := accounts[from];!ok{
		return "", errors.New("from("+from+")账户不存在")
	}
	client := rpc.NewClient("https://s1.ripple.com:51234", "https://data.ripple.com")

	//检查地址金额
	address := from
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
		log.Info(err.Error())
		return "", err
	}

	//获取节点信息填充last信息
	server, err := client.GetServerInfo()
	if err != nil {
		log.Info(err.Error())
		return "", err
	}

	//随机填tag
	//rand.Seed(time.Now().Unix())
	//tag := uint32(rand.Intn(12345))

	//指定tag
	tag := uint32(0)
	//发送模板
	tpl := &XprSignTpl{
		From:         address,
		FromPrivate:  accounts[address],
		To:           "rEcQvidASJ8Q7MnettfyyAuptGgRsu5VV1",
		AmountFloat:  amount, //改大金额就可以假充值
		FeeFloat:     "0.000012",
		FromSequence: account.AccountData.Sequence,
		LastSequence: server.State.ValidatedLedger.Seq + 100,
		Tag:          tag,
		Currency:     "XRP",
		CoinDecimal:  6,
	}
	rawtx, err := tpl.XrpSignTx()
	if err != nil {
		log.Info(err.Error())
		return "", err
	}
	log.Info("raw tx: ", rawtx)

	//如果需要广播，就使用下面的代码
	//广播 这里没有向上抛出异常，应该自己再封装
	tx, err := client.Submit(rawtx)
	if err != nil {
		log.Info("交易发送错误:"+err.Error())
		return "", err
	}
	log.Info("txid: ", tx.TxJson.Hash)
	return tx.TxJson.Hash,nil
}
