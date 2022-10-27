package common

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Unconfirmed = 0

	Confirmed = 1
	Spent     = 2
	//Vote = 4
	Claimed = 8
	//Locked = 16
	Frozen = 32
	//WatchOnly = 54
)

// string to int64
func StrToInt64(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}

// string to int64, base 进制(2,8,16)等
func StrBaseToInt64(str string, base int) int64 {
	if base == 16 {
		if strings.HasPrefix(str, "0x") {
			str = strings.TrimPrefix(str, "0x")
		}
	}
	val, _ := strconv.ParseInt(str, base, 64)
	return val
}

// string to int64, base 进制(2,8,16)等
func StrBaseToInt(str string, base int) int {
	if base == 16 {
		if strings.HasPrefix(str, "0x") {
			str = strings.TrimPrefix(str, "0x")
		}
	}
	val, _ := strconv.ParseInt(str, base, 32)
	return int(val)
}

// string to int64, base 进制(2,8,16)等
func StrBaseToBigInt(str string, base int) (*big.Int, bool) {
	if base == 16 {
		if strings.HasPrefix(str, "0x") {
			str = strings.TrimPrefix(str, "0x")
		}
	}
	return big.NewInt(0).SetString(str, base)
}

// string to int
func StrToInt(str string) int {
	val, _ := strconv.Atoi(str)
	return val
}

// string to float
func StrToFloat32(str string) float32 {
	tmp, _ := strconv.ParseFloat(str, 32)
	return float32(tmp)
}

// string to float64
func StrToFloat64(str string) float64 {
	tmp, _ := strconv.ParseFloat(str, 64)
	return tmp
}

// float64 to string
func Float64ToString(val float64) string {
	return strconv.FormatFloat(val, 'E', -1, 64)
}

// int to string
func IntToString(val int) string {
	return strconv.Itoa(val)
}

// int64 to string
func Int64ToString(val int64, base int) string {
	return strconv.FormatInt(val, base)
}

// int64 to string
func UInt64ToString(val uint64, base int) string {
	return strconv.FormatUint(val, base)
}

// utc时间转换成zh字符串时间
func TimeToStr(val int64) string {
	if val == 0 {
		return "2006-01-02 15:04:05"
	}
	tm := time.Unix(val, 0)
	return tm.Format("2006-01-02 15:04:05")
}

// utc时间转换成zh字符串时间
func StrToTime(val string) int64 {
	if val == "" {
		return 0
	}
	p, _ := time.Parse("2006-01-02 15:04:05", val)
	return p.Unix()
}

// 获取毫秒UTC时间
func GetMillTime() int64 {
	timestamp := time.Now().UnixNano() / 1000000
	return timestamp
}

// 返回生成的RSA私钥和公钥
func GenerateRSAKey(bits int) (string, string) {
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

	return "private.pem", "public.pem"
}

func GetRSAKey(path string) string {
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
	return string(buf)
}

//RSA公钥加密
func RSAEncrypt(path string, data []byte) []byte {

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

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	//类型断言
	publicKey := publicKeyInterface.(*rsa.PublicKey)
	//对明文进行加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		panic(err)
	}
	//返回密文
	return cipherText
}

func ParsePrivateKey(der []byte) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key.(*rsa.PrivateKey), nil
		default:
			return nil, errors.New("crypto/tls: found unknown private key type in PKCS#8 wrapping")
		}
	}
	return nil, errors.New("crypto/tls: failed to parse private key")
}

//RSA私钥解密
func RSADecrypt(path string, data []byte) []byte {
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
	//privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	//if err!=nil{
	//	panic(err)
	//}
	privateKey, err := ParsePrivateKey(block.Bytes)

	//对密文进行解密
	plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)

	//返回明文
	return plainText
}

//使用私钥签名，path是私钥路径，msg是要签名的信息
func RSASign(path string, data []byte) []byte {
	//签名函数中需要的数据散列值
	//首先从文件中提取公钥
	fp, _ := os.Open(path)
	defer fp.Close()
	fileinfo, _ := fp.Stat()
	buf := make([]byte, fileinfo.Size())
	fp.Read(buf)
	block, _ := pem.Decode(buf)
	PrivateKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	//加密操作,需要将接口类型的pub进行类型断言得到公钥类型

	hash := sha256.Sum256(data)
	//调用签名函数,填入所需四个参数，得到签名
	sign, _ := rsa.SignPKCS1v15(rand.Reader, PrivateKey, crypto.SHA256, hash[:])
	//fmt.Printf("sign:%x\n", sign)
	return sign

}

// 公钥验证签名
func RSAVerifySign(path string, signText []byte, data []byte) bool {
	//首先从文件中提取公钥
	fp, _ := os.Open(path)
	defer fp.Close()
	//测量文件长度以便于保存
	fileinfo, _ := fp.Stat()
	buf := make([]byte, fileinfo.Size())
	fp.Read(buf)
	//下面的操作是与创建秘钥保存时相反的
	//pem解码
	block, _ := pem.Decode(buf)
	//x509解码,得到一个interface类型的pub
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	//签名函数中需要的数据散列值
	hash := sha256.Sum256(data)
	//验证签名
	err := rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, hash[:], signText)
	if err != nil {
		return false //"认证失败"
	} else {
		return true //"认证成功"
	}
}
