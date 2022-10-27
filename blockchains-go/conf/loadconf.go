package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
)

var Cfg *Config

const (
	//defaultConfPath = "/app/services/blockchains-go/conf/application.conf"
	defaultConfPath = ""
)

//初始化配置文件
func InitConfig() {
	Cfg = LoadConfig()
}

func LoadConfig() *Config {
	var err error

	cfgPath := defaultConfPath
	cfg := &Config{}
	if cfgPath == "" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("laod config getpwd error:%s", err.Error())
		}
		cfgPath = fmt.Sprintf("%s/conf/%s.conf", dir, "application")
	}

	bs, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Fatalf("load config decode error:%s", err.Error())
		return nil
	}

	//bss, err := util.AesBase64Str(string(bs), aesKey0, false)
	//if err != nil {
	//	log.Fatalf("read conf  error:%s", err.Error())
	//	return nil
	//}

	if _, err = toml.Decode(string(bs), cfg); err != nil {
		log.Fatalf("load config decode error:%s", err.Error())
	}

	log.Infof("string(bs): %v", string(bs))
	log.Infof("cfg: %v", cfg)

	//if _, err := toml.DecodeFile(configFile, cfg); err != nil {
	//	log.Fatalf("load config decode error:%s", err.Error())
	//}
	if err = checkCfg(cfg); err != nil {
		log.Fatalf("load config check error:%s", err.Error())
	}
	//解密数据库
	if cfg.Encryption {
		//log.Infof("decryptCfg")
		decryptCfg(cfg)
	}
	//else {
	//	log.Infof("loadCfg")
	//	loadCfg(cfg)
	//}

	return cfg
}
func LoadConfig3(configFile string, cfg interface{}) error {
	if configFile == "" {
		dir, err := os.Getwd()
		if err != nil {
			log.Errorf("laod config getpwd error:%s", err.Error())
			return err
		}
		configFile = fmt.Sprintf("%s/conf/%s.toml", dir, "app")
	}
	if _, err := toml.DecodeFile(configFile, cfg); err != nil {
		log.Errorf("load config decode error:%s", err.Error())
		return err
	}
	return nil
}

//业务那边是读写分离
func checkCfg(cfg *Config) error {
	if cfg.DB.Master == "" {
		return errors.New("miss master config")
	}
	if len(cfg.DB.Slaves) == 0 {
		return errors.New("miss slaves config")
	}
	return nil
}
func DecryptCfg(cfg *CollectConfig) {
	if cfg.Mode == "prod" {
		master, _ := util.AesBase64Str(cfg.DB.Master, aesKeyDB, false)
		slaves := make([]string, 0)
		for _, v := range cfg.DB.Slaves {
			slave, _ := util.AesBase64Str(v, aesKeyDB, false)
			if slave != "" {
				slaves = append(slaves, slave)
			}
		}
		cfg.DB.Master = master
		cfg.DB.Slaves = slaves

		for k, v := range cfg.Collectors {
			wUrl, _ := util.AesBase64Str(v.Url, aesKey, false)
			wUser, _ := util.AesBase64Str(v.User, aesKey, false)
			wPwd, _ := util.AesBase64Str(v.Password, aesKey, false)
			cfg.Collectors[k].Url = wUrl
			cfg.Collectors[k].User = wUser
			cfg.Collectors[k].Password = wPwd
		}
	}
}

func decryptCfg(cfg *Config) {
	//数据库
	master, _ := util.AesBase64Str(cfg.DB.Master, aesKeyDB, false)
	master2, _ := util.AesBase64Str(cfg.DB2.Master, aesKeyDB, false)
	cfg.DB2.Master = master2
	slaves := make([]string, 0)
	for _, v := range cfg.DB.Slaves {
		slave, _ := util.AesBase64Str(v, aesKeyDB, false)
		if slave != "" {
			slaves = append(slaves, slave)
		}
	}

	cfg.DB.Master = master
	cfg.DB.Slaves = slaves

	//热钱包节点信息解密
	mapHotserver := make(map[string]*HotServers, 0)
	for k, v := range cfg.HotServers {
		hurl := v.Url
		huser := v.User
		hpwd := v.Password
		if !strings.HasSuffix(hurl, "http") {
			hurl, _ = util.AesBase64Str(v.Url, aesKey, false)
			huser, _ = util.AesBase64Str(v.User, aesKey, false)
			hpwd, _ = util.AesBase64Str(v.Password, aesKey, false)
		}
		mapHotserver[k] = &HotServers{
			Url:      hurl,
			User:     huser,
			Password: hpwd,
		}
	}
	cfg.HotServers = mapHotserver

	//walletserver信息解密
	wUrl := cfg.Walletserver.Url
	wUser := cfg.Walletserver.User
	wPwd := cfg.Walletserver.Password
	if !strings.HasSuffix(wUrl, "http") {
		wUrl, _ = util.AesBase64Str(cfg.Walletserver.Url, aesKey, false)
		wUser, _ = util.AesBase64Str(cfg.Walletserver.User, aesKey, false)
		wPwd, _ = util.AesBase64Str(cfg.Walletserver.Password, aesKey, false)
	}
	cfg.Walletserver = Walletserver{
		Url:      wUrl,
		User:     wUser,
		Password: wPwd,
	}

	//redis解密
	rUrl := cfg.Redis.Url
	rUser := cfg.Redis.User
	rPwd := cfg.Redis.Password
	if !strings.HasSuffix(rUrl, "http") {
		rUrl, _ = util.AesBase64Str(cfg.Redis.Url, aesKey, false)
		rUser, _ = util.AesBase64Str(cfg.Redis.User, aesKey, false)
		rPwd, _ = util.AesBase64Str(cfg.Redis.Password, aesKey, false)
	}
	cfg.Redis = RedisConfig{
		Url:      rUrl,
		User:     rUser,
		Password: rPwd,
	}

	mapCoinServer := make(map[string]*CoinServers, 0)
	for k, v := range cfg.CoinServers {
		curl := v.Url
		cuser := v.User
		cpwd := v.Password
		if !strings.HasSuffix(curl, "http") {
			curl, _ = util.AesBase64Str(v.Url, aesKey, false)
			cuser, _ = util.AesBase64Str(v.User, aesKey, false)
			cpwd, _ = util.AesBase64Str(v.Password, aesKey, false)
		}
		mapCoinServer[k] = &CoinServers{
			Url:      curl,
			User:     cuser,
			Password: cpwd,
		}
	}
	cfg.WeChat.Url, _ = util.AesBase64Str(cfg.WeChat.Url, aesKey, false)
	cfg.CoinServers = mapCoinServer

}

func json2str(v interface{}) string {
	marshal, _ := json.Marshal(v)
	return string(marshal)
}

//func loadCfg(cfg *Config) {
//	//数据库
//	master := cfg.DB.Master
//	master2 := cfg.DB2.Master
//	cfg.DB2.Master = master2
//	slaves := make([]string, 0)
//	for _, v := range cfg.DB.Slaves {
//		slave := v
//		if slave != "" {
//			slaves = append(slaves, slave)
//		}
//	}
//
//	cfg.DB.Master = master
//	cfg.DB.Slaves = slaves
//
//	//热钱包节点信息解密
//	mapHotserver := make(map[string]*HotServers, 0)
//	log.Infof("cfg.HotServers: %s", json2str(cfg.HotServers))
//	for k, v := range cfg.HotServers {
//		hurl := v.Url
//		huser := v.User
//		hpwd := v.Password
//		mapHotserver[k] = &HotServers{
//			Url:      hurl,
//			User:     huser,
//			Password: hpwd,
//		}
//	}
//	cfg.HotServers = mapHotserver
//	log.Infof("mapHotserver: %s", json2str(mapHotserver))
//
//	//walletserver信息解密
//	wUrl := cfg.Walletserver.Url
//	wUser := cfg.Walletserver.User
//	wPwd := cfg.Walletserver.Password
//	cfg.Walletserver = Walletserver{
//		Url:      wUrl,
//		User:     wUser,
//		Password: wPwd,
//	}
//
//	//redis解密
//	rUrl := cfg.Redis.Url
//	rUser := cfg.Redis.User
//	rPwd := cfg.Redis.Password
//	cfg.Redis = RedisConfig{
//		Url:      rUrl,
//		User:     rUser,
//		Password: rPwd,
//	}
//
//	mapCoinServer := make(map[string]*CoinServers, 0)
//	for k, v := range cfg.CoinServers {
//		curl := v.Url
//		cuser := v.User
//		cpwd := v.Password
//		mapCoinServer[k] = &CoinServers{
//			Url:      curl,
//			User:     cuser,
//			Password: cpwd,
//		}
//	}
//	cfg.WeChat.Url = cfg.WeChat.Url
//	cfg.CoinServers = mapCoinServer
//
//}
