package library

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestRSA(t *testing.T) {
	//bytes, err := base64.StdEncoding.DecodeString("admin@123")
	//if err != nil {
	//	return
	//}
	pu := RSA_Encrypt([]byte("admin@123"), "public.pem")
	fmt.Println(string(pu))

	decodeString, err := base64.StdEncoding.DecodeString("8738590fb64af8fb27f75f993b3a7edb96042ef967e64ac4b1a2cd12360b0323376172e68b69e36f70576bce2048ef59cc9e9776ad2153d733d754a8aa9fd506ce7ce346e680c8fb395589483d4676024ed2d913dcc544449f3b52549b6970e788f081c458d3a6f6221b20c02cba48736384f3127b8c96e291d958ffa271818f711f5ea071e15592bc9ff30d81854308f4cf2e6475a0279e44188e7b05ab112496c1334e5d6f1de3b98bb1763088d821ec853c91203cdb788e8544b023f3261426b2791971ffdc19e379a6f8631ab3ac3f30340fd99041b4cf280f6475d27ae6878dd36ed384ac31ec897c345d07a0acb923ef82a9bf04071291ae9e02c0a324")
	if err != nil {
		return
	}
	str, _ := RSA_Decrypt(decodeString, "private.pem")
	fmt.Println(string(str))
}

//测试加密函数
func TestEncryptPwd(t *testing.T) {
	fmt.Println(EncryptSha256Password("123456", "FCUXIVPXUY"))
}

func TestRSA2(t *testing.T) {

	privatekey, _ := ioutil.ReadFile("private.pem")
	pirvatekey := string(privatekey)
	public, _ := ioutil.ReadFile("public.pem")
	pubKey := string(public)
	//pirvatekey := "-----BEGIN RSA PRIVATE KEY-----\nMIICXgIBAAKBgQDXmqwFmi9TPVL1NgYZgGaHdPz7ZsskAOukaYg/hZkAn4rj8JJcaE/DzRy707bq6YBUpp+1ssiaadJRKR6+XMSzdyJz44Vl5rrArExrzGiNry+txRVo00U4xbIrmMX+aRKvPQ/xJ7xkxRhdxsY4F/wMcWDyJS+eyxJkHiBrHJ3ciQIDAQABAoGBAJV7eZUQx4sQ03mLkUMREQUNiXDMXj+CG96MBJj2CZSzCNrsqq1C7Tq19RwMt5+7cOw/8i9J22ejwtvehKA7NWxpsUBC+lDqXk3FCvtbL3d2fcARdh/1zWZN9WRvafkVPNPAeRC6ARp63DOe8FkT0C22DTOd0Xyvo0Zp7pF/GjXhAkEA/BkAFlV/4jHEglyXHNGrReMjClw2ClqKK5VXIk6UCJfVaGNDGbfw0ueYFnnOeIo8GPhgVjSC4wU2rX89pSFxTQJBANrxDqKc6wFw1jGpmxI25inxYTvA3SuSk36b4CSrRL7w3g9r+6QQfAlpBRZ9NBCL9WHeWHtgauxeDGJB2kmXui0CQQDhadlSHxFCSA3WIsRb2H609uwWD22ixGJXpilLW8eyB1GjDV6qWHbVno+3SSL9VV13Vl+NtVZzd+30JJoSVVzhAkB8sISxP8TnUSfrqLhUK0fx4zKJIVHUmum9VXDV8WR5ihwtlEYALhM2GMV5BV09fzgEwOiLe2Hps7ZBz1dOSkcRAkEA+D3kzvNpEYtqpjGHfUCxwmu/BwathDo09vj+gCcjhoJh/ADpa8+a0RQA6vcVMges0UcmIiIyQPNzCGlLBXtl9A==\n-----END RSA PRIVATE KEY-----"
	//pubKey := "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDXmqwFmi9TPVL1NgYZgGaHdPz7ZsskAOukaYg/hZkAn4rj8JJcaE/DzRy707bq6YBUpp+1ssiaadJRKR6+XMSzdyJz44Vl5rrArExrzGiNry+txRVo00U4xbIrmMX+aRKvPQ/xJ7xkxRhdxsY4F/wMcWDyJS+eyxJkHiBrHJ3ciQIDAQAB\n-----END PUBLIC KEY-----"
	label := ""
	plaintext := "admin@123"
	// OAEP加密
	ciphertext, _ := RSAEncryptOAEP(pubKey, label, plaintext)
	fmt.Println("OAEP加密")
	fmt.Println(ciphertext)
	// OAEP解密
	plaintext, _ = RSADecryptOAEP(pirvatekey, label, ciphertext)
	fmt.Println("OAEP解密")
	fmt.Println(plaintext)

	// PKCS1v15加密
	ciphertext, _ = RSAEncryptPKCS1v15(pubKey, plaintext)
	fmt.Println(ciphertext)
	// PKCS1v15解密
	plaintext, _ = RSADecryptPKCS1v15(pirvatekey, "la2HFvwi+eL0H4T4LnRbpBcn09oHp1lagw6+8zsx6lO1+PAy2fAX2i7xJXhe9kFQatJJNn4twtGQbUAR7wYP6FBcwqisTr0IXd6pUiiCVkM/rCmwyWbnNXbckGL984qEKSFj8V2l8a38bveiQXDOF7eA3TbNvkdn5wcfi2J1wtYswNcHkeVA6zo4K/pZrunLJUfEBR1KIKruapw+Pp4ZqhW5EMGRJ/gfwIyAcuij0s1z2eOGSflNx4QQi3Lz7CA6b/5DiJzeuqoLWnNzfLm3wBnZE8r8owSjoMWzjXNj9czBnqtTsXUSVCZajc0hJ8twcERaPRsi4QuwA8/XZVTQrw==")
	fmt.Printf(plaintext)

	//签名 和验证
	signBytes, err := SignPKCS1v15(pirvatekey, []byte(plaintext), crypto.SHA256)
	if err != nil {
		fmt.Println("SignPKCS1v15 error")
	}
	fmt.Println("SignPKCS1v15 base64:" + base64.StdEncoding.EncodeToString(signBytes))
	err = VerifyPKCS1v15Verify(pubKey, []byte(plaintext), signBytes, crypto.SHA256)
	if err == nil {
		fmt.Println("VerifyPKCS1v15Verify oaky")
	}
}

// cipherText:加密后的文本 path：存放私钥的地址
func RSADecrypt(cipherText, path string) string {
	code, _ := base64.StdEncoding.DecodeString(cipherText)
	//打开文件
	file, err := os.Open(path)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	//对密文进行解密
	plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, privateKey, code)
	//返回明文
	return string(plainText)
}
