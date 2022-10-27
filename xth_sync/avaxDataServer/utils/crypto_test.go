//run:go test
package utils

import (
	"fmt"
	"testing"
)

const DefaulAesKey = "vkkhgF1xVu9DDm/5YJTwZ0L1x8vDdd+K"

func Test_AesCrypt_de(t *testing.T) {
	//decode
	aes_decode("6kDzq6McR2q5Id4Br0Q=", DefaulAesKey)
	aes_decode("6lL2ob0ER2g=", DefaulAesKey)
	aes_decode("6knmqfRSBnmnKNoKqgoC284lriwMppBmnYedYdOK9LFRdab8+F9Fo7uT1PsVaW/XnGhtL5xTqCA=", DefaulAesKey)

}
func Test_AesCrypt_en(t *testing.T) {
	aes_encode("avaxsync", DefaulAesKey)
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
