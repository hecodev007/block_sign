package library

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestAES(t *testing.T) {

	c := "{\"code\":\"" + "1111" + "\"}"
	fmt.Println(c)
	key := []byte("REGLGZWHBHJCHHGN")
	decodeString := []byte("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6ZmFsc2UsImlkIjoyLCJtZXJjaGFudF9pZCI6MCwicm9sZSI6MiwibmFtZSI6InRvYnkiLCJhY2NvdW50IjoiMTc3NzE3Mjc3MzIiLCJub25jZSI6IlJFR0xHWldIQkhKQ0hIR04iLCJleHAiOjE2NDI0Mzk2MjJ9.ueqKzN2xJQ8pMv5WZR-_BIz8kui3j6THnbWDt7STtT8bde3c2a4-36ec-4f6f-864f-968fe738a1ac1642410880")

	encrypted := AesEncryptECB(decodeString, key)

	//bytes, err := base64.StdEncoding.DecodeString("kDAYxfUqUFSbFrBryBeEplYI8KYTVb2PmSFgxX+mT+MNZuUaCMHGQ1uKPKcdtDK2pxScNLxORyzTKoJ8LzxDw4e78JNmRrh1IRL/cIUlUYsc9OXrohtX2XIh2+pd3nDL9JXs/aoLyxQ+bB5E2jcOKQOdhnWmRJZZ5GEePywNTH98D990Sb6DQAGe77p7QVBK422++6HDmvXvnmCFHp6JQ+IYn0sHyrs19XDbKwxWMBIcnIM+gidGmk8fRffmwae1/cRH6xyU1kyondk/R12H5bpuHCm9aRrqKmDh8MnTB+dicIMwXz7/n/wA16YPxOGq2jRDpmAe13d8oEDV9KRI8u2vkzpgSwJce5IxIiy+s6dDNHgz3e0Y02yg7PTX1/ulV2K2mZyjln7hUrF5WyyjshRy3MEL4+Bd6wfCQp+5T5o=")
	//if err != nil {
	//	return
	//}
	//aesDecrypt, err := AesDecrypt(bytes, key)
	//if err != nil {
	//	return
	//}

	fmt.Println(base64.StdEncoding.EncodeToString(encrypted))
	//fmt.Println(base64.StdEncoding.EncodeToString(aesDecrypt))
}

// =================== CBC ======================
func AesEncryptCBC(origData []byte, key []byte) (encrypted []byte) {
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	origData = pkcs5Padding(origData, blockSize)                // 补全码
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) // 加密模式
	encrypted = make([]byte, len(origData))                     // 创建数组
	blockMode.CryptBlocks(encrypted, origData)                  // 加密
	return encrypted
}
func AesDecryptCBC(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)                              // 分组秘钥
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) // 加密模式
	decrypted = make([]byte, len(encrypted))                    // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)                 // 解密
	decrypted = pkcs5UnPadding(decrypted)                       // 去除补全码
	return decrypted
}
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
