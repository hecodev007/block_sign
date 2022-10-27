package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/group-coldwallet/trxsync/utils"
)

const defConfFile = "conf/app.toml"
const DefaulAesKey = ""

type Rabbit struct {
	HostPort string
	Username string
	Password string
}
type Config struct {
	Mode      string `toml:"mode"`
	Server    ServerConfig
	Push      PushConfig
	Sync      SyncConfig
	DataBases map[string]DatabaseConfig
	Nodes     map[string]NodeConfig
	Mq        Rabbit
}
type FixConfig struct {
	Name  string `toml:"name"`
	Csv   CsvConfig
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

//type LogConfig struct {
//	Level     string `toml:"level"       json:"level"`
//	Formatter string `toml:"formatter"   json:"formatter"`
//	OutFile   string `toml:"outfile"    json:"outfile"`
//	ErrFile   string `toml:"errfile"    json:"errfile"`
//}

type PushConfig struct {
	Enable   bool   `toml:"enable"`
	Type     string `toml:"type"`
	Agent    bool   `toml:"agent"`
	Url      string `toml:"url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	//MqUrl      string   `toml:"mqurl"`
	//Reconns    int      `toml:"reconns"`
	//Publishers []string `toml:"publishers"`
}

type SyncConfig struct {
	Name             string `toml:"name"`
	EnableSync       bool   `toml:"enablesync"`
	MultiScanTaskNum int64  `toml:"multiscantasknum"`
	MultiScanNum     int64  `toml:"multiscannum"`
	InitHeight       int64  `toml:"initheight"`
	EnableRollback   bool   `toml:"enablerollback"`
	RollHeight       int64  `toml:"rollheight"`
	DelayHeight      int64  `toml:"delayheight"`
	Confirmations    int64  `toml:"confirmations"`
	SleepTime        int64  `toml:"sleeptime"`
	EnableStop       bool   `toml:"enablestop"`
	StopHeight       int64  `toml:"stopheight"`
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
	if cfg.Sync.SleepTime == 0 {
		cfg.Sync.SleepTime = 1
	}
	return nil
}
