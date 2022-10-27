package tests

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JFJun/arweave-go/utils"
	"github.com/btcsuite/btcutil/base58"
	"github.com/mendsley/gojwk"
	"sync"
	"testing"
	"time"
	"wallet-sign/util"
)

func createAddressInfo() {
	//var addrInfo util.AddrInfo
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	fmt.Println(priv.D)
	if err != nil {
		panic(err)
	}
	key, err := gojwk.PrivateKey(priv)
	data, err := gojwk.Marshal(key)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	k2, _ := gojwk.Unmarshal(data)
	priv2, err := k2.DecodePrivateKey()
	if err != nil {
		panic(err)
	}
	p := priv2.(*rsa.PrivateKey)
	fmt.Println(p.D)
	fmt.Println(len(p.D.Bytes()))
	fmt.Println(hex.EncodeToString(p.D.Bytes()))
	priv.D.Bytes()
	h := sha256.New()
	h.Write(priv.PublicKey.N.Bytes())
	address := utils.EncodeToBase64(h.Sum(nil))
	fmt.Println(address)
}

func Test_Ar(t *testing.T) {
	////createAddressInfo()
	//c, err := api.Dial("https://arweave.net")
	//if err != nil {
	//	panic(err)
	//}
	////info,err:=c.GetInfo(context.TODO())
	////if err != nil {
	////	panic(err)
	////}
	////fmt.Println(info)
	////resp, err := c.LastTransaction(context.TODO(), "MfXpZ_h_uX-icWAHOuGM_5zXhn60yECfnd-rGQPVG3w")
	////
	////if err != nil {
	////	panic(err)
	////}
	////fmt.Println(resp)
	//resp,err:=c.GetBalance(context.TODO(),"NE_Xl5Rp085n4p0QMpnDR9NfKvbA50wyVQykHsYJYZE")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(resp)

	//now:=time.Now().Unix()
	//fmt.Println(Timestamp(now))
	//fmt.Println(now)
	//
	//newNow:=time.Unix(now,0).Add(time.Minute*time.Duration(3))
	//
	//fmt.Println(Timestamp(newNow.Unix()))
	//s := map[int]int{
	//	1: 1,
	//	2: 2,
	//	3: 3,
	//}
	//
	//for k, v := range s {
	//	if v == 2 {
	//		delete(s, k)
	//	}
	//}
	//fmt.Println(s)
	//a,_:=hex.DecodeString("acecf4fc8c5bc44465bf60d2de4608691264192ea579bc857eab231881c2f7f4")
	//fmt.Println(a)
	//s:=ed25519.NewKeyFromSeed(a)
	//fmt.Println(s)
	//fmt.Println(base58.Encode(s[32:]))
	//d := fmt.Sprintf("0x70a08231%064s", "4cd05c7e7d674550f741289abfd622b76d11fd37")
	//fmt.Println(d)
	//TT()
	isPendingStatus("0xacd39943680106a6613cd55a10e299214ff9882fa363f585aacf2bdc5f8e323b")
}

func Timestamp(seconds int64) string {
	var timelayout = "2006-01-02 T 15:04:05.000" //时间格式

	datatimestr := time.Unix(seconds, 0).Format(timelayout)

	return datatimestr

}

func TT() {
	wif := "5JJfXBM3V5ciPkXwQiL8zN2JPBFbYFywMm3SwhY7LvKLwnekB6L"
	decodeWIF(wif)
}

func decodeWIF(wif string) error {
	pc := base58.Decode(wif)
	checkSum := pc[len(pc)-4:]
	payload := pc[:len(pc)-4]
	if len(payload) != 33 {
		return errors.New("private key len is not correct")
	}
	c := DoubleSha256(payload)
	checkSum2 := c[:4]
	if !bytes.Equal(checkSum, checkSum2) {
		return errors.New("checkSum is not equal")
	}
	private := payload[1:]
	fmt.Println(hex.EncodeToString(private))
	fmt.Println(private)
	fmt.Println(len(private))

	return nil
}
func DoubleSha256(data []byte) []byte {
	d := sha256.Sum256(data)
	dd := sha256.Sum256(d[:])
	return dd[:]

}

func isPendingStatus(txid string) bool {
	client := util.New("http://eth.rylink.io:31545", "", "")
	data, err := client.SendRequest("eth_getTransactionReceipt", []interface{}{txid})
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
	return false
}

var (
	nonceCtl = sync.Map{}
)

func Test_BSCNonce(t *testing.T) {
	nonceCtl.Store("test1", map[string]int64{
		"1": 4,
		"2": 5,
		"3": 6,
	})
	nonce, err := aegisNonce("test1")
	if err != nil {
		panic(err)
	}
	fmt.Println("获取到的nonce为： ", nonce)
	//========================================
	fmt.Println("====================================")
	nm, ok := nonceCtl.Load("test1")
	if ok {
		nonceMap := nm.(map[string]int64)
		nonceMap["4"] = nonce + 1
		nonceCtl.Store("test1", nonceMap)
	} else {
		nonceMap := make(map[string]int64)
		nonceMap["4"] = nonce + 1
		nonceCtl.Store("test1", nonceMap)
	}
	v, _ := nonceCtl.Load("test1")
	fmt.Println(v)
}

func aegisNonce(address string) (int64, error) {
	//1. 先获取该地址链上的nonce值
	nonce := int64(5)

	// 2. 判断内存中是否有这个地址的nonce
	value, ok := nonceCtl.Load(address)
	if !ok {
		// 2.1  如果内存中没有，直接返回链上的nonce
		return nonce, nil
	}
	if value == nil {
		return -1, fmt.Errorf(" %s do not find any value in map", address)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return -1, err
	}
	var bnd map[string]int64
	err = json.Unmarshal(data, &bnd)
	if err != nil {
		return -1, err
	}
	if len(bnd) > 30 {
		return -1, fmt.Errorf("%s address pending tx is big than 30", address)
	}
	for k, v := range bnd {
		//判断是否处于pending状态
		if v <= 5 {
			delete(bnd, k)
			continue
		}
		if v > nonce {
			nonce = v
		}
	}
	nonceCtl.Store(address, bnd)
	return nonce, nil
}
