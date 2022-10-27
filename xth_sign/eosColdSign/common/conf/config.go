package conf

import (
	"github.com/BurntSushi/toml"
)

const defConfFile = "app.toml"

func init() {
	if err := LoadConfig(defConfFile); err != nil {
		panic(err)
	}
}

var config Config

type Config struct {
	RunMode string `toml:"name"`
	Body
}

type Body struct {
	AppName  string `toml:"appname"`
	CoinName string `toml:"coinName"`
	HttpPort int64  `toml:"httpport"`
	OutFile  string `toml:"outFile"`
	ErrFile  string `toml:"errFile"`
	Url      string `toml:"url"`
}

//从相对路径Load conf
//请传入指针类型
func LoadConfig(cfgPath string) error {
	if cfgPath == "" {
		cfgPath = defConfFile
	}

	if _, err := toml.DecodeFile(cfgPath, &config); err != nil {
		return err
	}

	return nil
}
func GetConfig() *Config {
	return &config
}
