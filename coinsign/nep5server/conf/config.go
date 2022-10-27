package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"time"
)

type Config struct {
	Dev     string     `toml:"dev"`   //开发模式
	HttpCfg HttpConfig `toml:"http"`  //http服务器配置
	LogCfg  LogConfig  `toml:"log"`   //日志配置
	Model   string     `toml:"model"` //冷热模式
	Nep5Cfg []Nep5     `toml:"nep5"`
	PemPath string     `toml:"pem_path"`
}
type HttpConfig struct {
	Port         string        `toml:"port"`         //服务器端口
	ReadTimeout  time.Duration `toml:"readtimeout"`  //读取超时,秒
	WriteTimeout time.Duration `toml:"writetimeout"` //写入超时,秒
}

type LogConfig struct {
	LogName  string `toml:"name"`
	LogPath  string `toml:"path"`
	LogSPath string `toml:"spath"` //软连接
	LogLevel string `toml:"level"`
}

type Nep5 struct {
	Name     string `toml:"name"`
	Decimal  int32  `toml:"decimal"`
	AssetsId string `toml:"assetsId"`
}

//初始化配置文件
func InitConfig(confName string) *Config {
	cfg := new(Config)
	if confName == "" {
		confName = "application.toml"
	}
	configFile := fmt.Sprintf("./conf/%s", confName)
	if _, err := toml.DecodeFile(configFile, cfg); err != nil {
		panic(err)
	}
	return cfg
}
