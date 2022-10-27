package conf

/*
使用toml进行配置
*/

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego"
	"path/filepath"
	"sync"
)

var (
	Global *globalConfig
	once   sync.Once
)

func Init() {
	once.Do(func() {
		globalFilePath, err := filepath.Abs(beego.AppConfig.String("globalConfigPath"))
		if err != nil {
			panic(err)
		}
		if _, err := toml.DecodeFile(globalFilePath, &Global); err != nil {
			panic(fmt.Sprintf("Global config file error,Err=[%v]", err))
		}

	})
}

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
