package conf

import (
	"github.com/BurntSushi/toml"
)

const defConfFile = "./conf/app.toml"

func init() {
	if err := LoadConfig(defConfFile); err != nil {
		panic(err)
	}
}

var config Config
var Global *globalConfig

type Config struct {
	RunMode string `toml:"name"`
	Body
}

type Body struct {
	AppName          string `toml:"appname"`
	CoinName         string `toml:"coinName"`
	HttpPort         int64  `toml:"httpport"`
	OutFile          string `toml:"outFile"`
	ErrFile          string `toml:"errFile"`
	Url              string `toml:"url"`
	GlobalConfigPath string `toml:"globalConfigPath"`
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

	//globalFilePath, err := filepath.Abs(config.GlobalConfigPath)
	//if err != nil {
	//	panic(err)
	//}
	//if _, err := toml.DecodeFile(globalFilePath, &Global); err != nil {
	//	panic(fmt.Sprintf("Global config file error,Err=[%v]", err))
	//}

	return nil
}
func GetConfig() *Config {
	return &config
}

//================================详细配置参数===================================

type globalConfig struct {
	Secret SecretConf `toml:"secret"`
	KMS    *KMSConf   `toml:"kms"`
}

type KMSConf struct {
	Url         string `toml:"url"`
	HttpTimeout uint32 `toml:"httpTimeout"`
	RetryCount  uint32 `toml:"retryCount"`
}

type SecretConf struct {
	TransportSecureKey string `toml:"transportSecureKey"`
	PrivateKeyPem      string `toml:"privateKeyPem"`
	PublicKeyPem       string `toml:"publicKeyPem"`
}
