//run:go test
package utils

import (
	"fmt"
	"testing"
)

const DefaulAesKey = "vkkhgF1xVu9DDm/5YJTwZ0L1x8vDdd+K"

func Test_AesCrypt_de(t *testing.T) {
	//decode
	aes_decode("70bxsqQQRCa9KMYXtVUYn3suNzOh/Z4AvewXK/eJ49f6Ypz+ZsIyjHpyRfL9Z6GNezHXMevAE2/eWAuYvwV+VFIp+AhQo8DORhE5", DefaulAesKey)
	aes_decode("4EXhuL0ER2js", DefaulAesKey)
	aes_decode("5WDdobo8DWy5fIQ6u25YkoPZ", DefaulAesKey)

}

//name= "6kDzq6McR2q5Id4Br0Q="
//type= "mysql"
//url= "70XjuL0YW323J9ZKolEHy2dO6URED2pJn/zF51B37pivhogejk9cjkpIh6kRh35PEyPR/hrj1l92AFKFiTzy"
//user= "70XjuJEOTHmoLdAB"
//password= "32in/6k/X2K8fNwO+UIG3Q=="

//"data_service:TL0&gBvib8oj8rll@tcp(dataservice-cluster.cluster-camzhqc6mnkb.ap-northeast-1.rds.amazonaws.com:12306)/dotsync
func Test_AesCrypt_en(t *testing.T) {
	aes_encode("dataService", DefaulAesKey)
	aes_encode("nDJxtA$gg87^z^2#QS", DefaulAesKey)
	aes_encode("dbfkjmm-cluster.cluster-camzhqc6mnkb.ap-northeast-1.rds.amazonaws.com:12306", DefaulAesKey)
	aes_encode("cringsync", DefaulAesKey)
	aes_encode("TL0&gBvib8oj8rll", DefaulAesKey)
}

//加密
func aes_encode(data, key string) (endata string, err error) {
	endata, err = AesBase64Str(data, key, true)
	fmt.Println(endata, data)
	return
}

//解密
func aes_decode(data, key string) (dedata string, err error) {
	dedata, err = AesBase64Str(data, key, false)
	fmt.Println(data, dedata)
	return
}
