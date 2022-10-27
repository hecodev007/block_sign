package conf

import (
	"github.com/BurntSushi/toml"
)

var Cfg = new(Config)

var defConfFile = "app.toml"
type Config struct {
	Name      string `toml:"appname"`
	Cointype string `toml:"coin"`
	Mode      string `toml:"mode"`
	Server    ServerConfig
	Log       LogConfig
}


type ServerConfig struct {
	Port         string
}


type LogConfig struct {
	Level     string `toml:"level"       json:"level"`
	Formatter string `toml:"formatter"   json:"formatter"`
	OutFile   string `toml:"outfile"    json:"outfile"`
	ErrFile   string `toml:"errfile"    json:"errfile"`
}

type PushConfig struct {
	Enable     bool     `toml:"enable"`
	Type       string   `toml:"type"`
	Agent      bool     `toml:"agent"`
	Url        string   `toml:"url"`
	User       string   `toml:"user"`
	Password   string   `toml:"password"`
	MqUrl      string   `toml:"mqurl"`
	Reconns    int      `toml:"reconns"`
	Publishers []string `toml:"publishers"`
}

//从相对路径Load conf
//请传入指针类型
func LoadConfig(cfgPath string, cfg *Config) error {

	if cfgPath == "" {
		cfgPath = defConfFile
	}


	if _, err := toml.DecodeFile(cfgPath, cfg); err != nil {
		return err
	}

	return nil
}
