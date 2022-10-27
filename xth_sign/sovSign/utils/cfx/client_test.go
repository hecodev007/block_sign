package cfx

import (
	"github.com/shopspring/decimal"
	"math/big"
	"testing"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"

)

func Test_client(t *testing.T){
	client,err :=sdk.NewClient("http://main.confluxrpc.org/v2")
	if err != nil {
		panic(err.Error())
	}
	v,err:=client.GetClientVersion()

	if err != nil {
		panic(err.Error())
	}
	t.Log(v)
	return
	from :=cfxaddress.MustNewFromBase32("CFX:TYPE.USER:AAKGVZ75B8DBVTF23DHSZGY0A8WT08MMCUSDTZBJFT")
	balance ,err :=client.GetBalance(from)
	t.Log(from.String())

	if err != nil {
		panic(err.Error())
	}
	t.Log(balance.ToInt().String())
	t.Log(decimal.NewFromBigInt(balance.ToInt(),-18).String())
}
func Test_token(t *testing.T){
	client,err :=sdk.NewClient("http://main.confluxrpc.org/v2")
	if err != nil {
		panic(err.Error())
	}
	abijson := "[{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"
	tokenaddr := "cfx:achc8nxj7r451c223m18w2dwjnmhkd6rxawrvkvsy2"
	tokenaddress,err := cfxaddress.NewFromBase32(tokenaddr)
	if err != nil{
		panic(err.Error())
	}
	balance := &struct{ Balance *big.Int }{}

	erc20,err := sdk.NewContract([]byte(abijson),client,&tokenaddress)
	addr,_ :=cfxaddress.NewFromBase32("CFX:TYPE.USER:AAKGVZ75B8DBVTF23DHSZGY0A8WT08MMCUSDTZBJFT")
	//addr2,_:=addr.ToHex()
	//addr.MustGetCommonAddress()
	//t.Log(addr2)
	if err = erc20.Call(nil, balance, "balanceOf", addr.MustGetCommonAddress());err != nil{
		panic(err.Error())
	}
	t.Log(balance.Balance.String())
}