package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
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

func AesCrypt2(str, key, iv []byte, encry bool) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
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

func Md5(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

func Md5Buf(buf []byte) string {
	hashMd5 := md5.New()
	hashMd5.Write(buf)
	//return fmt.Sprintf("%x", hashMd5.Sum(nil))
	return hex.EncodeToString(hashMd5.Sum(nil))
}

func Md5File(reader io.Reader) string {
	var buf = make([]byte, 4096)
	hashMd5 := md5.New()
	for {
		n, err := reader.Read(buf)
		if err == io.EOF && n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			break
		}

		hashMd5.Write(buf[:n])
	}

	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

func AesBase64Str(data, key string, encry bool) (string, error) {
	if encry {
		if result, err := AesCrypt([]byte(data), []byte(key), true); err != nil {
			return "", err
		} else {
			return base64.StdEncoding.EncodeToString(result), nil
		}
	} else {
		if dst, err := base64.StdEncoding.DecodeString(data); err != nil {
			return "", err
		} else {
			if result, err := AesCrypt(dst, []byte(key), false); err != nil {
				return "", err
			} else {
				return string(result), nil
			}
		}
	}
}

// write by jun
/*
对热钱包的参数进行签名，服务端进行验证
*/
func CreateTransferParamsSign(from, to, amount, contract, currentTime string) (string, error) {
	secret := fmt.Sprintf("%sHOO_WALLET_TRANSFER_SERVICE%s", currentTime, currentTime)
	params := fmt.Sprintf("from=%s&to=%s&amount=%s&contract=%s&time=%s", from, to, amount, contract, currentTime)
	key := sha256.Sum256([]byte(secret))
	sig, err := AesBase64Crypt([]byte(params), key[:], true)
	if err != nil {
		return "", fmt.Errorf("加密req data error:%v", err)
	}
	crypt := sha256.Sum256(sig)
	return "0x" + hex.EncodeToString(crypt[:]), err
}
