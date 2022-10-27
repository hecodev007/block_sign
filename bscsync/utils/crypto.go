package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
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

// 加解密字符串
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

func AesCrtCrypt(str, key, iv []byte) ([]byte, error) {
	// 指定加密、解密算法为AES，返回一个AES的Block接口对象
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("iv len isn't %d", aes.BlockSize)
	}
	// 指定计数器,长度必须等于block的块尺寸
	// var iv = key[:aes.BlockSize]
	// 指定分组模式
	blockMode := cipher.NewCTR(aesBlock, iv)
	// 执行加密、解密操作
	message := make([]byte, len(str))
	blockMode.XORKeyStream(message, str)
	return message, nil
}

// 加解密出base64可读字符串
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
	// return fmt.Sprintf("%x", hashMd5.Sum(nil))
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
