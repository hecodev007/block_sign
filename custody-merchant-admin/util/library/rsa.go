package library

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
)

// GenerateRSAKey 生成RSA私钥和公钥，保存到文件中
// bits 证书大小
func GenerateRSAKey(bits int) {
	//GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	//Reader是一个全局、共享的密码用强随机数生成器
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(err)
	}
	//保存私钥
	//通过x509标准将得到的ras私钥序列化为ASN.1 的 DER编码字符串
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	//使用pem格式对x509输出的内容进行编码
	//创建文件保存私钥
	privateFile, err := os.Create("private.pem")
	if err != nil {
		panic(err)
	}
	defer privateFile.Close()
	//构建一个pem.Block结构体对象
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	//将数据保存到文件
	pem.Encode(privateFile, &privateBlock)

	//保存公钥
	//获取公钥的数据
	publicKey := privateKey.PublicKey
	//X509对公钥编码
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		panic(err)
	}
	//pem格式编码
	//创建用于保存公钥的文件
	publicFile, err := os.Create("public.pem")
	if err != nil {
		panic(err)
	}
	defer publicFile.Close()
	//创建一个pem.Block结构体对象
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	//保存到文件
	pem.Encode(publicFile, &publicBlock)
}

// RSA_Encrypt RSA加密
// plainText 要加密的数据
// path 公钥匙文件地址
func RSA_Encrypt(plainText []byte, path string) []byte {
	//打开文件
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	//读取文件的内容
	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)
	//pem解码
	block, _ := pem.Decode(buf)
	//x509解码

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	//类型断言
	publicKey := publicKeyInterface.(*rsa.PublicKey)
	//对明文进行加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText)
	if err != nil {
		panic(err)
	}
	//返回密文
	return cipherText
}

// RSA_Decrypt RSA解密
// cipherText 需要解密的byte数据
// path 私钥文件路径
func RSA_Decrypt(cipherText []byte, path string) ([]byte, error) {
	//打开文件
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	//获取文件内容
	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)
	//pem解码
	block, _ := pem.Decode(buf)
	//X509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//对密文进行解密
	plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	//返回明文
	return plainText, nil
}

func RSAEncryptOAEP(publicKeypem, labeltext, plaintext string) (ciphertext string, err error) {
	publicBlock, _ := pem.Decode([]byte(publicKeypem))
	if publicBlock == nil {
		//panic("public key error")
		return "", fmt.Errorf("public key error")
	}
	pub, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		//panic("publicKey is not  *rsa.PublicKey")
		return "", err
	}
	publicKey := pub.(*rsa.PublicKey)
	rng := rand.Reader

	secretMessage := []byte(plaintext)
	label := []byte(labeltext)
	cipherbyte, err := rsa.EncryptOAEP(sha256.New(), rng, publicKey, secretMessage, label)
	if err != nil {
		//panic(fmt.Sprintf("Error from encryption: %s\n", err))
		return "", err
	}

	// 由于加密是随机函数，密文将是
	// 每次都不一样。
	//fmt.Printf("Ciphertext: %x\n", cipherbyte)
	ciphertext = fmt.Sprintf("%x\n", cipherbyte)
	return ciphertext, nil
}

func RSADecryptOAEP(privateKeypem, labeltext, ciphertext string) (plaintext string, err error) {
	privateBlock, _ := pem.Decode([]byte(privateKeypem))
	if privateBlock == nil {
		//panic("private key error")
		return "", fmt.Errorf("private key error")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		//panic("privateKey is not  *rsa.PrivateKey")
		return "", fmt.Errorf("privateKey is not  *rsa.PrivateKey")
	}

	/*
	   prkI, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	   if err != nil {
	       panic("privateKey is not  *rsa.PrivateKey")
	   }
	   privateKey := prkI.(*rsa.PrivateKey)
	*/
	rng := rand.Reader
	cipherByte, _ := hex.DecodeString(ciphertext)
	label := []byte(labeltext)

	plainbyte, err := rsa.DecryptOAEP(sha256.New(), rng, privateKey, cipherByte, label)
	if err != nil {
		//panic(fmt.Sprintf("Error decrypting: %s\n", err))
		return "", fmt.Errorf("Error decrypting: %s\n", err)

	}
	// 由于加密是随机函数，密文将是
	// 每次都不一样。
	plaintext = string(plainbyte)
	return plaintext, nil
}

func RSAEncryptPKCS1v15(publicKeypem, plaintext string) (ciphertext string, err error) {
	publicBlock, _ := pem.Decode([]byte(publicKeypem))
	if publicBlock == nil {
		//panic("public key error")
		return "", fmt.Errorf("public key error")

	}
	pub, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		//panic("publicKey is not  *rsa.PublicKey")
		return "", fmt.Errorf("publicKey is not  *rsa.PublicKey")
	}
	publicKey := pub.(*rsa.PublicKey)
	plaintbyte := []byte(plaintext)
	blocks := pkcs1Padding(plaintbyte, publicKey.N.BitLen()/8)

	buffer := bytes.Buffer{}
	for _, block := range blocks {
		ciphertextPart, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, block)
		if err != nil {
			//panic(fmt.Sprintf("Error EncryptPKCS1v15: %s\n", err))
			return "", fmt.Errorf("Error EncryptPKCS1v15: %s\n", err)

		}
		buffer.Write(ciphertextPart)
	}

	ciphertext = fmt.Sprintf("%x\n", buffer.Bytes())
	return ciphertext, nil
}

func RSADecryptPKCS1v15(privateKeypem, ciphertext string) (plaintext string, err error) {
	privateBlock, _ := pem.Decode([]byte(privateKeypem))
	if privateBlock == nil {
		return "", err
		//panic("private key error")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return "", err
		//panic("privateKey is not  *rsa.PrivateKey")
	}

	/*
	   prkI, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	   if err != nil {
	       panic("privateKey is not  *rsa.PrivateKey")
	   }
	   privateKey := prkI.(*rsa.PrivateKey)
	*/
	///
	cipherByte, _ := hex.DecodeString(ciphertext)

	ciphertextBlocks := unPadding(cipherByte, privateKey.N.BitLen()/8)

	buffer := bytes.Buffer{}
	for _, ciphertextBlock := range ciphertextBlocks {
		plaintextBlock, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertextBlock)
		if err != nil {
			return "", err
			//panic(fmt.Sprintf("Error DecryptPKCS1v15: %s\n", err))
		}
		buffer.Write(plaintextBlock)
	}
	plaintext = string(buffer.Bytes())
	return plaintext, err
}

func SignPKCS1v15(privateKeypem string, src []byte, hash crypto.Hash) ([]byte, error) {
	privateBlock, _ := pem.Decode([]byte(privateKeypem))
	if privateBlock == nil {
		//panic("private key error")
		return nil, fmt.Errorf("private key error")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("privateKey is not  *rsa.PrivateKey")

		//panic("privateKey is not  *rsa.PrivateKey")
	}
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
}

func VerifyPKCS1v15Verify(publicKeypem string, src []byte, sign []byte, hash crypto.Hash) error {
	publicBlock, _ := pem.Decode([]byte(publicKeypem))
	if publicBlock == nil {
		//panic("public key error")
		return fmt.Errorf("public key error")

	}
	pub, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		//panic("publicKey is not  *rsa.PublicKey")
		return err

	}
	publicKey := pub.(*rsa.PublicKey)
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.VerifyPKCS1v15(publicKey, hash, hashed, sign)
}

func pkcs1Padding(src []byte, keySize int) [][]byte {
	srcSize := len(src)
	blockSize := keySize - 11
	var v [][]byte
	if srcSize <= blockSize {
		v = append(v, src)
	} else {
		groups := len(src) / blockSize
		for i := 0; i < groups; i++ {
			block := src[:blockSize]
			v = append(v, block)
			src = src[blockSize:]
			if len(src) < blockSize {
				v = append(v, src)
			}
		}
	}
	return v
}

func unPadding(src []byte, keySize int) [][]byte {
	srcSize := len(src)
	blockSize := keySize
	var v [][]byte
	if srcSize == blockSize {
		v = append(v, src)
	} else {
		groups := len(src) / blockSize
		for i := 0; i < groups; i++ {
			block := src[:blockSize]

			v = append(v, block)
			src = src[blockSize:]
		}
	}
	return v
}

//RSA公钥私钥产生 GenRsaKey(1024)
func GenRsaKey(bits int) error {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	file, err := os.Create("private.pem")
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	// 生成公钥文件
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	file, err = os.Create("public.pem")
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil

}
