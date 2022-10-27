package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/exp/rand"
	"hash"
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
func AesCryptCfb(str, key []byte, encry bool) ([]byte, error) {
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
func AesBase64CryptCfb(data, key []byte, encry bool) ([]byte, error) {
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

func Hash160String(data []byte) string {
	return hex.EncodeToString(Hash160(data))
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}

func Md5HashString(data []byte) string {
	signByte := []byte(data)
	hash := md5.New()
	hash.Write(signByte)
	return hex.EncodeToString(hash.Sum(nil))
}
