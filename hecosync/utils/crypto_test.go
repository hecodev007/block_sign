//run:go test
package utils

import (
	"fmt"
	"testing"
)

const DefaulAesKey = "vkkhgF1xVu9DDm/5YJTwZ0L1x8vDdd+K"

const aesKey = "THLWkroDC0BcpSxTK7Ate4PeNx8Dn5uQ"

const aesKey0 = "Cf4T0bYhYm7Z0s0/DLv01Dq2fa6sKlHT"

const aesKeyDB = "THLWkroDC0BcpSxTK7Ate4PeNx8Dn5uQ"

func Test_AesCrypt_de(t *testing.T) {
	//decode
	aes_decode("70bxsqQQRCa9KMYXtVUYn3suNzOh/Z4AvewXK/eJ49f6Ypz+ZsIyjHpyRfL9Z6GNezHXMevAE2/eWAuYvwV+VFIp+AhQo8DORhE5", DefaulAesKey)
	aes_decode("4EXhuL0ER2js", DefaulAesKey)
	aes_decode("5WDdobo8DWy5fIQ6u25YkoPZ", DefaulAesKey)

}

func Test_AesCrypt_block(t *testing.T) {
	//decode
	aes_decode("70bxsqQQRCa9KMYXtVUYn3suNzOh/Z4AvewXK/eJ49f6Ypz+ZsIyjHpyRfL9Z6GNezHXMevAE2/eWAuYvwV+VFIp+AhQo8DORhE5", aesKey)
	aes_decode("http://47.75.171.17:2201", aesKey)
	aes_decode("", aesKey)

	aes_encode("http://47.75.171.17:22022", aesKey)
	aes_encode("", aesKey)

}

//name= "6kDzq6McR2q5Id4Br0Q="
//type= "mysql"
//url= "70XjuL0YW323J9ZKolEHy2dO6URED2pJn/zF51B37pivhogejk9cjkpIh6kRh35PEyPR/hrj1l92AFKFiTzy"
//user= "70XjuJEOTHmoLdAB"
//password= "32in/6k/X2K8fNwO+UIG3Q=="

//name= "8kvjra8OUGW9"
//type= "mysql"
//url= "70bxsqQQRCW9Jd4eqUEJh6+rxQyAm/u6l7UHpyET8UsjYHSthv5OeRx3jFjl7iccw1ucPR/26dIuNf0="
//user= "70XjuJ0YW323J9Y="
//password= "5WDdobo8DWy5fIQ6u25YkoPZ"
func Test_AesCrypt_en(t *testing.T) {
	aes_encode("hecosync", DefaulAesKey)
	aes_encode("addrmanagement", DefaulAesKey)
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
