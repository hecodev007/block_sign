package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"os"
)

var (
	Conf              config // holds the global app config.
	DefaultConfigFile = "config.toml"
	EnvPro            = false
)

type config struct {
	Mod               string              `json:"mod"`
	ReleaseMode       bool                `toml:"release_mode" `
	CdnDisable        bool                `toml:"cdn_disable"`
	JwtOut            bool                `toml:"jwt_out"`
	LogLevel          string              `toml:"log_level"`
	LogFile           string              `toml:"log_file"`
	SessionStore      string              `toml:"session_store"`
	CacheStore        string              `toml:"cache_store" description:"cache缓存信息"`
	Sms               map[string]sms      `toml:"sms" description:"手机短信配置"`
	Price             map[string]price    `toml:"price" `
	DB                map[string]database `toml:"database" description:"MySQL 配置"`
	Grpc              map[string]grpc     `toml:"grpc" description:"grpc 配置"`
	Email             email               `description:"邮箱信息配置"`
	Sns               sns                 `description:"权限模型列表"`
	App               app                 `description:"应用配置"`
	Server            server              `description:"权限模型列表"`
	Redis             redis               `description:"Redis配置"`
	Memcached         memcached           `description:"Memcached内存配置"`
	Opentracing       opentracing         `description:"链路追踪配置"`
	RabbitMQ          Rabbitmq            `description:"mq配置"`
	Casbin            casbin              `description:"权限模型列表"`
	Finance           Finance             `description:"财务配置"`
	Blockchain        blockchain          `description:"blockchainsgo的Api配置"`
	BlockchainCustody blockchaincustody   `description:"blockchainsgo的路由配置"`
	Fee               Fee                 `toml:"fee"`
	Wlwx              wlwx                `description:"wlwx短信"`
	StaticFile        string              `toml:"static_file"`
}

type sms struct {
	AppKey    string `toml:"app_key"`
	AppSecret string `toml:"app_secret"`
	AppCode   string `toml:"app_code"`
	Batch     string `toml:"batch"`
	Distinct  string `toml:"distinct"`
	Balance   string `toml:"balance"`
}

type sns struct {
	Region      string `toml:"region"`
	AccessKeyId string `toml:"access_key_id"`
	SecretKey   string `toml:"secret_key"`
}

type price struct {
	Url string `toml:"url"`
}

type Fee struct {
	Open  bool    `toml:"open"`
	Url   string  `toml:"url"`
	Rate  float64 `toml:"rate"`
	Limit int     `toml:"limit"`
}

type email struct {
	SmtpPassword string   `toml:"smtp_password"`
	SmtpUsername string   `toml:"smtp_username"`
	IamUserName  string   `toml:"iam_user_name"`
	Host         string   `toml:"host"`
	Port         int      `toml:"port"`
	Title        string   `toml:"title"`
	Recipient    []string `toml:"recipient"`
}

type wlwx struct {
	CustomName   string `toml:"custom_name"`
	CustomPwd    string `toml:"custom_pwd"`
	SmsClientUrl string `toml:"sms_client_url"`
	Uid          string `toml:"uid"`
	Content      string `toml:"content"`
	CestMobiles  string `toml:"cest_mobiles"`
	NeedReport   bool   `toml:"need_report"`
	SpCode       string `toml:"sp_code"`
}

type app struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

type server struct {
	Graceful     bool   `toml:"graceful"`
	Addr         string `toml:"addr"`
	DomainApi    string `toml:"domain_api"`
	DomainWeb    string `toml:"domain_web"`
	DomainSocket string `toml:"domain_socket"`
}

type database struct {
	Name     string `toml:"name"`
	UserName string `toml:"user_name"`
	Pwd      string `toml:"pwd"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
}

type blockchain struct {
	Url       string `toml:"url"`
	ClientId  string `toml:"client_id"`
	ApiSecret string `toml:"api_secret"`
}

type blockchaincustody struct {
	ClientId             string `toml:"client_id"`
	ApiSecret            string `toml:"api_secret"`
	CallBackBaseUrl      string `toml:"call_back_base_url"`
	BaseUrl              string `toml:"base_url"`
	CoinList             string `toml:"coin_list"`
	CreateMch            string `toml:"create_mch"`
	ResetMch             string `toml:"reset_mch"`
	GetMch               string `toml:"get_mch"`
	VerifyParam          string `toml:"verify_param"`
	CreateAddress        string `toml:"create_address"`
	CreateLotCoinAddress string `toml:"create_lot_coin_address"`
	BindAddress          string `toml:"bind_address"`
	Withdraw             string `toml:"withdraw"`
	Balance              string `toml:"balance"`
	ChainStatus          string `toml:"chain_status"`
	WhiteIp              string `toml:"white_ip"`
}

type grpc struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

type redis struct {
	Model          string   `toml:"model"`
	AloneAddress   string   `toml:"alone_address"`
	AlonePwd       string   `toml:"alone_pwd"`
	ClusterPwd     string   `toml:"cluster_pwd"`
	ClusterAddress []string `toml:"cluster_address"`
}

type memcached struct {
	Server string `toml:"server"`
}

type opentracing struct {
	Disable     bool   `toml:"disable"`
	Type        string `toml:"type"`
	ServiceName string `toml:"service_name"`
	Address     string `toml:"address"`
}

type Rabbitmq struct {
	Prefix     string `toml:"prefix"`
	MQUrl      string `toml:"mq_url"`
	MQUser     string `toml:"mq_user"`
	MQPassword string `toml:"mq_password"`
	Reconns    int    `toml:"reconns"`
}

type Finance struct {
	Url string `toml:"url"`
}

type casbin struct {
	ModelPath string `toml:"model_path"`
}

func init() {
	err := InitConfig("")
	if err != nil {
		panic(err)
	}
}

// InitConfig initConfig initializes the app configuration by first setting defaults,
// then overriding settings from the app config files, then overriding
// It returns an error if any.
func InitConfig(configFile string) error {
	if configFile == "" {
		configFile = DefaultConfigFile
	}

	// Set defaults.
	Conf = config{
		ReleaseMode: false,
		LogLevel:    "DEBUG",
	}
	// @TODO 读取配置
	if _, err := os.Stat(configFile); err != nil {
		return errors.New("config files err:" + err.Error())
	} else {
		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return errors.New("config load err:" + err.Error())
		}
		_, err = toml.Decode(string(configBytes), &Conf)
		if err != nil {
			return errors.New("config decode err:" + err.Error())
		}
	}
	// @TODO 配置检查
	log.Debugf("config data:%v ", Conf)
	if Conf.Mod != "pro" {
		EnvPro = true
	}
	return nil
}

func GetLogLvl() log.Lvl {
	// DEBUG INFO WARN ERROR OFF
	switch Conf.LogLevel {
	case "DEBUG":
		return log.DEBUG
	case "INFO":
		return log.INFO
	case "WARN":
		return log.WARN
	case "ERROR":
		return log.ERROR
	case "OF":
		return log.OFF
	}

	return log.DEBUG
}

const (
	// Template Type
	PONGO2   = "PONGO2"
	TEMPLATE = "TEMPLATE"

	// Bindata
	BINDATA = "BINDATA"

	// File
	FILE = "FILE"

	// Redis
	REDIS = "REDIS"

	// Memcached
	MEMCACHED = "MEMCACHED"

	// Cookie
	COOKIE = "COOKIE"

	// In Memory
	IN_MEMORY = "IN_MEMARY"

	// RabbitMQ
	RabbitMQ = "RABBITMQ"
)
