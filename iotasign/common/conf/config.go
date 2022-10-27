package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

const defConfFile = "./conf/app.toml"

func init() {
	if err := LoadConfig(defConfFile); err != nil {
		panic(err)
	}
}

var Cfg *Config

type Config struct {
	Name   string    `toml:"name"`
	Mode   string    `json:"mode"`
	Csv    CsvConfig `toml:"csv"`
	Server ServerConfig
	Log    LogConfig
	Node   NodeConfig `toml:"node"`
}

type NodeConfig struct {
	Url       string `toml:"url"`
	RPCKey    string `toml:"rpc_key"`
	RPCSecret string `toml:"rpc_secret"`
	AssetID   string `toml:"assetID"`
	NetworkID uint32 `toml:"networkID"`
	ChainID   string `toml:"chainID"`
}

type CsvConfig struct {
	Dir       string `toml:"dir"`
	ReadFile  string `toml:"read_file"`
	WriteFile string `toml:"write_file"`
}

type ServerConfig struct {
	IP           string
	Port         string
	ReadTimeout  int    `toml:"read_timeout"`
	WriteTimeout int    `toml:"write_timeout"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
}

type DatabaseConfig struct {
	Name     string `toml:"name"`
	Type     string `toml:"type"`
	Url      string `toml:"url"`
	User     string `toml:"user"`
	PassWord string `toml:"password"`
	Mode     string `toml:"mode"`
}

type LogConfig struct {
	Level     string `toml:"level"       json:"level"`
	Formatter string `toml:"formatter"   json:"formatter"`
	OutFile   string `toml:"outfile"    json:"outfile"`
	ErrFile   string `toml:"errfile"    json:"errfile"`
}

//从相对路径Load conf
//请传入指针类型
func LoadConfig(cfgPath string) error {
	if cfgPath == "" {
		cfgPath = defConfFile
	}

	if _, err := toml.DecodeFile(cfgPath, &Cfg); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
func GetConfig() *Config {
	return Cfg
}
func GetLogConfig() *LogConfig {
	return &Cfg.Log
}
func GetServerConfig() *ServerConfig {
	return &Cfg.Server
}