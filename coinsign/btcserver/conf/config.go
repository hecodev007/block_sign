package conf

import (
	"fmt"
	"github.com/group-coldwallet/btcserver/util/rylinkbtcutil"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

const (
	DefaultFilePath = "./conf/"
	DefaultFile     = "application.yaml"
)

var GlobalConf = &GlobalConfig{}

//加载配置文件
func init() {
	yamlFile, err := ioutil.ReadFile(DefaultFilePath + DefaultFile)
	if err != nil {
		panic(fmt.Sprintf("yamlFile.Get err   #%v ", err))
	}
	err = yaml.Unmarshal(yamlFile, GlobalConf)
	if err != nil {
		panic(fmt.Sprintf("Unmarshal: %v", err))
	}

	rylinkbtcutil.NewCoinNet(GlobalConf.Coinnet)

}

//==================================配置参数====================================
//详细配置参数汇总
type GlobalConfig struct {
	RunModel    string     `yaml:"run_model"`    //运行模式: debug(开发) test(测试) release(生产)             ////基本配置参数
	Coinnet     string     `yaml:"coinnet"`      //main 主网 test测试网
	SystemModel string     `yaml:"system_model"` //冷热系统 hot热系统。cold冷系统
	HttpCfg     HttpConfig `yaml:"http"`         //http配置
	LogCfg      LogConfig  `yaml:"log"`          //log配置
	BtcCfg      BtcConfig  `yaml:"btc"`
	CronCfg     CronConfig `yaml:"cron"`
}

//==================================配置参数====================================

//================================详细配置参数===================================

//http配置
type HttpConfig struct {
	Port         int           `yaml:"port"`          //运行端口
	ReadTimeout  time.Duration `yaml:"read_timeout"`  //超时设置，单位秒
	WriteTimeout time.Duration `yaml:"write_timeout"` //超时设置，单位秒
}

//日志配置
type LogConfig struct {
	LogPath  string `yaml:"log_path"` //日志存储路径
	LogName  string `yaml:"log_name"` //日志名称
	LogLevel string `yaml:"log_level"`
}

type BtcConfig struct {
	Servers        []string `yaml:"servers,flow"`      //节点服务地址列表,冷签名暂时不需要
	PushServers    []string `yaml:"push_servers,flow"` //广播节点
	DefaultFee     int64    `yaml:"default_fee"`       //默认手续费
	MaxFee         int64    `yaml:"max_fee"`           //手续费
	MinFee         int64    `yaml:"min_fee"`           //手续费
	FilePath       string   `yaml:"file_path"`         //新规则的地址目录
	FilePathOld    string   `yaml:"file_path_old"`     //历史遗留旧地址的目录
	CreateAddrPath string   `yaml:"create_addr_path"`  //创建地址目录
}

//定时任务表达式配置
type CronConfig struct {
	LoadKeyJob string `json:"loadkeyjob"` //私钥加载表达式
}

//================================详细配置参数===================================
