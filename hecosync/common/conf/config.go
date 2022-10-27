package conf

import (
	"fmt"
	"hecosync/utils"
	"os"

	"github.com/BurntSushi/toml"
)

var Cfg = new(Config)

var defConfFiles = []string{"./app_dev.toml", "./app.toml"}

const DefaulAesKey = "vkkhgF1xVu9DDm/5YJTwZ0L1x8vDdd+K"

func init() {
	for _, confFile := range defConfFiles {
		if _, err := os.Lstat(confFile); !os.IsNotExist(err) {
			fmt.Println("读取配置文件:", confFile)
			if err := LoadConfig(confFile, Cfg); err != nil {
				panic(err.Error())
			}
			break
		}
	}
}

type Rabbit struct {
	HostPort string
	Username string
	Password string
}
type Config struct {
	Mode           string `toml:"mode"`
	DatabaseCrypto bool   `toml:"database_crypto"`
	Server         ServerConfig
	Log            LogConfig
	Push           PushConfig
	Sync           SyncConfig
	DataBases      map[string]DatabaseConfig
	Node           NodeConfig `toml:"node"`
	Mq             Rabbit
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
	IP   string
	Port string
}

type DatabaseConfig struct {
	Name     string `toml:"name"`
	Type     string `toml:"type"`
	Url      string `toml:"url"`
	User     string `toml:"user"`
	PassWord string `toml:"password"`
}

type LogConfig struct {
	Console   bool   `toml:"console"`
	Level     string `toml:"level"`
	Formatter string `toml:"formatter"`
	OutFile   string `toml:"outfile"`
	ErrFile   string `toml:"errfile"`
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

type SyncConfig struct {
	CoinName        string `toml:"coin_name"`
	FullBackup      bool   `toml:"fullbackup"`
	EnableSync      bool   `toml:"enablesync"`
	EnableMempool   bool   `toml:"enablemempool"`
	MultiScanNum    int64  `toml:"enablemultiscan"`
	InitHeight      int64  `toml:"initheight"`
	EnableRollback  bool   `toml:"enablerollback"`
	EnableInternal  bool   `toml:"enableInternal"`
	RollHeight      int64  `toml:"rollheight"`
	Confirmations   int64  `toml:"confirmations"`
	AddressRecover  int64  `toml:"address_discover"`
	ContractRecover int64  `toml:"contract_discover"`
}

type NodeConfig struct {
	Url       string `toml:"url"`
	RPCKey    string `toml:"rpc_key"`
	RPCSecret string `toml:"rpc_secret"`
}

//从相对路径Load conf
//请传入指针类型
func LoadConfig(cfgPath string, cfg *Config) error {

	if cfgPath == "" {
		panic("err file path:" + cfgPath)
	}

	if _, err := toml.DecodeFile(cfgPath, cfg); err != nil {
		return err
	}
	if cfg.DatabaseCrypto {
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
