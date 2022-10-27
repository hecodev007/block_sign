package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func Base64Encode(data []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return dst
}

func Base64Decode(data []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(dst, data)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}

//加解密字符串
func AesCrypt(str, key []byte, encry bool) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var iv = key[:aes.BlockSize]
	crypted := make([]byte, len(str))
	if encry {
		cipher.NewCFBEncrypter(aesBlock, iv).XORKeyStream(crypted, str)
		return crypted, nil
	}
	cipher.NewCFBDecrypter(aesBlock, iv).XORKeyStream(crypted, str)
	return crypted, nil
}

//加解密出base64可读字符串
func AesBase64Crypt(data, key []byte, encry bool) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var iv = key[:aes.BlockSize]
	crypted := make([]byte, len(data))
	if encry {
		cipher.NewCFBEncrypter(aesBlock, iv).XORKeyStream(crypted, data)
		return Base64Encode(crypted), nil
	}
	cipher.NewCFBDecrypter(aesBlock, iv).XORKeyStream(crypted, data)
	return Base64Encode(crypted), nil
}

func RandBase64Key() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	enKey := Base64Encode(key)
	return enKey[0:32]
}

func Demo() {
	aesKey := RandBase64Key()
	aesKey = []byte("9YGiaxgBjLVZXbCLnl3UKOHgCBQDZoes")
	//apk, _ := AesBase64Crypt([]byte("KyAevMVhccueJq9c4FiUUsXjQJudXWHHCLDeXaCMbx4JAnTY3VgK"), aesKey, true)
	apk, _ := AesBase64Crypt([]byte("KzBQakrhMAch8SLontwNxc43GUPvL9PHtotb56BdbMhsQjb29jhh"), aesKey, true)
	fmt.Println(string(apk))
	fmt.Println(string(aesKey))

}
