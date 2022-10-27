package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/group-coldwallet/flynn/register-service/util"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"sync"
)

const (
	aesKey = "vkkhgF1xVu9DDm/5YJTwZ0L1x8vDdd+K"
)

type tomlConfig struct {
	Debug   bool   `toml:"debug"`
	Port    string `toml:"port"`
	Version string `toml:"version"`
	AuthCfg struct {
		Encrypt  bool   `toml:"encrypt"`
		Enable   bool   `toml:"enable"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"auth"`

	DataBases    map[string]DatabaseConfig
	ScanServices map[string]ScanServiceCfg `toml:"services"`
}

type DatabaseConfig struct {
	Name     string `toml:"name"`
	Type     string `toml:"type"`
	Url      string `toml:"url"`
	User     string `toml:"user"`
	PassWord string `toml:"password"`
	Mode     string `toml:"mode"`
}

var (
	Config *tomlConfig
	once   sync.Once
)

func InitConfig(cfgPath string) {
	log.Infof("加载配置文件路径： [%s]", cfgPath)
	once.Do(func() {
		filePath, err := filepath.Abs(cfgPath)
		if err != nil {
			panic("do not find any config file")
		}
		if _, err := toml.DecodeFile(filePath, &Config); err != nil {
			//log.Errorf("Read config file error,Err=[%v]",err)
			panic(fmt.Sprintf("Read config file error,Err=[%v]", err))
		}
		//解密 auth
		if Config.AuthCfg.Encrypt {
			err = decryptAuthConfig(Config)
			if err != nil {
				panic(fmt.Sprintf("decode auth info error,Err=%v", err))
			}
		}
		if !Config.Debug {
			for k, v := range Config.DataBases {
				v.Name, _ = util.AesBase64Str(v.Name, aesKey, false)
				v.Url, _ = util.AesBase64Str(v.Url, aesKey, false)
				v.User, _ = util.AesBase64Str(v.User, aesKey, false)
				v.PassWord, _ = util.AesBase64Str(v.PassWord, aesKey, false)
				Config.DataBases[k] = v
			}
		}
	})
}
func decryptAuthConfig(cfg *tomlConfig) error {
	user, err := util.AesBase64Crypt([]byte(cfg.AuthCfg.User), []byte(aesKey), false)
	if err != nil {
		return err
	}
	cfg.AuthCfg.User = string(user)
	password, err := util.AesBase64Crypt([]byte(cfg.AuthCfg.Password), []byte(aesKey), false)
	if err != nil {
		return err
	}
	cfg.AuthCfg.Password = string(password)
	return nil
}

type ScanServiceCfg struct {
	Name     string `toml:"name"`
	Url      string `toml:"url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

func IsSupportThisCoin(name string) bool {
	if len(Config.ScanServices) == 0 {
		return false
	}
	for _, v := range Config.ScanServices {
		if strings.ToLower(v.Name) == strings.ToLower(name) {
			return true
		}
	}
	return false
}
