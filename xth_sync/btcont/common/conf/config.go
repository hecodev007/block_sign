package conf

import (
	"fmt"
	"io/ioutil"
)

var Cfg = new(Config)

type Config struct {
	EthNode string
	Db      DatabaseConfig
}

func init() {

	if body, err := ioutil.ReadFile("../environment"); err != nil {
		fmt.Println(err.Error())
		fmt.Println("生产")
		Cfg.EthNode = "http://192.169.1.40:20545"
		Cfg.Db = DatabaseConfig{
			Name:     "finance_data",
			Type:     "mysql",
			Url:      "finance.c4hvmlwnwyiv.ap-northeast-1.rds.amazonaws.com:12306",
			User:     "aMO2DlQcHrPPTfOb",
			PassWord: "sMJwbaMNw1^jMiDm",
			Mode:     "release",
		}

	} else {
		fmt.Println(string(body))
		fmt.Println("测试")
		Cfg.EthNode = "http://3.113.0.101:20545"
		Cfg.Db = DatabaseConfig{
			Name:     "finance_data",
			Type:     "mysql",
			Url:      "rm-j6c5ekl1af4dc9k8w499.mysql.rds.aliyuncs.com:3306",
			User:     "hoocustody",
			PassWord: "Eb!ZXrNt!!x5xru0",
			Mode:     "release",
		}
	}
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
	Mode      string `toml:"mode"    json:"mode"`
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
}

type NodeConfig struct {
	Url       string `toml:"url"`
	ScanApi   string `toml:"scan_api"`
	ScanKey   string `toml:"scan_key"`
	Node      string `toml:"node"`
	RPCKey    string `toml:"rpc_key"`
	RPCSecret string `toml:"rpc_secret"`
}
