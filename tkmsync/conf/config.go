package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"rsksync/utils"
)

const defConfFile = "conf/app.toml"
const DefaulAesKey = "vkkhgF1xVu9DDm/5YJTwZ0L1x8vDdd+K"

type Rabbit struct {
	HostPort string
	Username string
	Password string
}
type Config struct {
	Name       string `toml:"name"`
	Mode       string `toml:"mode"`
	Server     ServerConfig
	Log        LogConfig
	Push       PushConfig
	Sync       SyncConfig
	DataBases  map[string]DatabaseConfig
	Nodes      map[string]NodeConfig
	EthScanCfg EthScanConfig `toml:"ethscan"`
	Mq         Rabbit
}
type FixConfig struct {
	Name  string `toml:"name"`
	Csv   CsvConfig
	Log   LogConfig
	Nodes map[string]NodeConfig
}

type CsvConfig struct {
	Dir       string `toml:"dir"`
	ReadFile  string `toml:"read_file"`
	WriteFile string `toml:"write_file"`
}

type ServerConfig struct {
	IP           string
	Port         string
	ReadTimeout  int `toml:"read_timeout"`
	WriteTimeout int `toml:"write_timeout"`
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

type PushConfig struct {
	Enable   bool   `toml:"enable"`
	Type     string `toml:"type"`
	Agent    bool   `toml:"agent"`
	Url      string `toml:"url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	//MqUrl      string   `toml:"mqurl"`
	Reconns    int      `toml:"reconns"`
	Publishers []string `toml:"publishers"`
}

type SyncConfig struct {
	Name            string `toml:"name"`
	FullBackup      bool   `toml:"fullbackup"`
	EnableSync      bool   `toml:"enablesync"`
	EnableMempool   bool   `toml:"enablemempool"`
	MultiScanNum    int64  `toml:"enablemultiscan"`
	EnableGoroutine bool   `toml:"enablegoroutine"`
	EpochCount      int64  `toml:"epochcount"`
	EpochTime       int64  `toml:"epochtime"`
	InitHeight      int64  `toml:"initheight"`
	EnableRollback  bool   `toml:"enablerollback"`
	RollHeight      int64  `toml:"rollheight"`
	Confirmations   int64  `toml:"confirmations"`
	IntervalTime    int64  `toml:"intervaltime"`
	AddressRecover  int64  `toml:"address_discover"`
	ContractRecover int64  `toml:"contract_discover"`
}

type NodeConfig struct {
	Url       string `toml:"url"`
	RPCKey    string `toml:"rpc_key"`
	RPCSecret string `toml:"rpc_secret"`
}

type EthScanConfig struct {
	Host string `toml:"host"`
	Key  string `toml:"key"`
}

//???????????????Load conf
//?????????????????????
func LoadConfig(cfgPath string, cfg *Config) error {

	if cfgPath == "" {
		cfgPath = defConfFile
	}

	if _, err := toml.DecodeFile(cfgPath, cfg); err != nil {
		fmt.Println(err)
		return err
	}
	if cfg.Mode == "prod" {
		for k, v := range cfg.DataBases {
			v.Name, _ = utils.AesBase64Str(v.Name, DefaulAesKey, false)
			v.Url, _ = utils.AesBase64Str(v.Url, DefaulAesKey, false)
			v.User, _ = utils.AesBase64Str(v.User, DefaulAesKey, false)
			v.PassWord, _ = utils.AesBase64Str(v.PassWord, DefaulAesKey, false)
			cfg.DataBases[k] = v
		}
		//cfg.Push.MqUrl, _ = utils.AesBase64Str(cfg.Push.MqUrl, DefaulAesKey, false)
	}
	return nil
}
