package okt

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/exchain-go-sdk/utils"
	"github.com/shopspring/decimal"
)

func Test_addr(t *testing.T) {
	addr, pri, err := GentAccount2()
	if err != nil {
		panic(err.Error())
	}
	t.Log(addr, pri)

}
func Test_key(t *testing.T) {
	addr, err := utils.ToHexAddress("ex1fv6rnjyy3mj0gfvgu9kuv7vpsaqpjnz28u7ln6")
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log(addr.String())
	addrbytes, err := types.GetFromBech32("okexchain1y93qykqa489ftwym3gvx4pqwacmzua6tefmkht", "okexchain")
	if err != nil {
		t.Fatal(err.Error())
	}
	acc := types.AccAddress(addrbytes)
	acc.String()
	t.Log(acc.String())
}
func Test_pri(t *testing.T) {
	//0x37570d947c5e66c784b7763424a0f13524396106
	info, err := utils.CreateAccountWithPrivateKey("02E437EBB4163F4F080209E6CEA9B348F1F351386DB6F78B5493642C2F321B24", "test", "")
	if err != nil {
		panic(err.Error())
	}
	t.Log(info.GetAddress().String())

}

func Test_dec(t *testing.T) {
	price := decimal.NewFromInt(123456)
	d := decimal.NewFromBigInt(price.Mul(decimal.NewFromFloat(1.2)).BigInt(), 0)
	t.Log(d.String())
}
