package fil

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/big"
	"testing"
)

func Test_address(t *testing.T) {
	a := big.NewInt(131106)
	//c, _ := a.Bytes()
	t.Log(fmt.Sprintf("0x%x", a))
	//addr, pri, err := CreateAddress2()
	//if err != nil {
	//	panic(err.Error())
	//}
	//fmt.Println(addr, pri)
	//fmt.Println("123")
	//pri := "6d4802f3d1f93833c8e739e960695ca3c95b5991730f9a692694b6728df8b1a1"
	//private, _ := hex.DecodeString(pri)
	//pub, err := new(SecpSigner).ToPublic(private)
	//if err != nil {
	//	panic(err.Error())
	//}
	//addr, err := address.NewSecp256k1Address(pub)
	//if err != nil {
	//	panic(err.Error())
	//}
	//fmt.Println(addr.String())

}
