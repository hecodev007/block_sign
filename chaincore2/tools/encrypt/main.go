package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/models"
	"log"
)

var (
	help     bool
	act      string
	genkey   bool
	key      string
	src      string
	confpath string
)

var StdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
var AsciiChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+,.?/:;{}[]`~")

func init() {
	flag.BoolVar(&help, "h", false, "this help")
	flag.BoolVar(&help, "help", false, "this help")

	flag.StringVar(&act, "act", "", "act:[genkey,enc,dec,encconf]")
	flag.StringVar(&key, "key", "", "encrypt key")
	flag.StringVar(&src, "src", "", "String to be encrypted")
	flag.StringVar(&confpath, "confpath", "", "config path")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	switch act {
	case "genkey":
		log.Println(NewLenChars(32, AsciiChars))
	case "enc":
		if key != "" && src != "" {
			dsc, err := common.AesEncrypt(src, []byte(key))
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(dsc)
		}
	case "dec":
		if key != "" && src != "" {
			dsc, err := common.AesDecrypt(src, []byte(key))
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(dsc)
		}
	case "encconf":
		if confpath != "" {
			if err := beego.LoadAppConfig("ini", confpath); err != nil {
				log.Println(err)
				return
			}
		}
		conf := &models.EtcdConfig{
			Enableauth: beego.AppConfig.DefaultBool("enableauth", true),
			Username:   beego.AppConfig.String("username"),
			Password:   beego.AppConfig.String("password"),

			// db
			Userdsn: beego.AppConfig.String("userdsn"),
			Syncdsn: beego.AppConfig.String("syncdsn"),

			// node
			Nodeurl:   beego.AppConfig.String("nodeurl"),
			Walleturl: beego.AppConfig.String("walleturl"),
			Rpcuser:   beego.AppConfig.String("rpcuser"),
			Rpcpass:   beego.AppConfig.String("rpcpass"),

			// agent
			Agenturl:  beego.AppConfig.String("agenturl"),
			Agentuser: beego.AppConfig.String("agentuser"),
			Agentpass: beego.AppConfig.String("agentpass"),
		}
		data, err := json.Marshal(conf)
		if err == nil {
			log.Println(string(data))
			dsc, err := common.AesEncrypt(string(data), []byte(key))
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(dsc)
		}
	default:
		flag.Usage()
	}
}

func usage() {
	flag.PrintDefaults()
}

// NewLenChars returns a new random string of the provided length, consisting of the provided byte slice of allowed characters(maximum 256).
func NewLenChars(length int, chars []byte) string {
	if length == 0 {
		return ""
	}
	clen := len(chars)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("Error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue // Skip this number to avoid modulo bias.
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
