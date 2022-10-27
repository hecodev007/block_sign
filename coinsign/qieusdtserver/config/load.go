package config

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/db"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	CONFIG_FILE_NAME = "config/application.yml"
)

var (
	appCfgFile string = CONFIG_FILE_NAME
)

//从相对路径Load config
func LoadConfig(cfgPath string) (*GlobalConfig, error) {

	var cfg = &GlobalConfig{}
	if cfgPath != "" {
		SetCfgPath(cfgPath)
	}
	absoluteConfigFile := GetConfigAbsolutePath(GetCfgPath())
	if err := parseConfig(cfg, absoluteConfigFile); err != nil {
		return nil, err
	}
	return cfg, nil
}

//从全路径Load config
func LoadConfigFullPath(fullPath string) (*GlobalConfig, error) {
	var cfg = &GlobalConfig{}
	if fullPath == "" {
		return nil, errors.New("empty full path")
	}
	if err := parseConfig(cfg, fullPath); err != nil {
		return nil, err
	}
	return cfg, nil
}

func SetCfgPath(cfgPath string) {
	appCfgFile = cfgPath
}

func GetCfgPath() string {
	return appCfgFile
}

func parseConfig(cfg interface{}, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	in, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(in, cfg)
}

//GetConfigAbsolutePath find config absolute path, for go test
func GetConfigAbsolutePath(file string) string {
	app := path.Dir(os.Args[0])
	if strings.HasPrefix(app, os.TempDir()) {
		return getConfigAbsolutePathForTest(file)
	}

	return getConfigAbsolutePathForBase(file)
}

func getConfigAbsolutePathForBase(file string) string {
	app := path.Base(os.Args[0])
	for _, dir := range []string{
		"",
		"config",
		"/etc/" + app,
		path.Join(os.Getenv("HOME"), "."+app),
	} {
		cf := path.Join(dir, file)
		if fileExists(cf) {
			return cf
		}
	}

	return ""
}

func getConfigAbsolutePathForTest(file string) string {
	_, filename, _, _ := runtime.Caller(2)
	dir := path.Dir(filename)
	for {
		for _, d := range []string{"", "config"} {
			cf := path.Join(dir, d, file)
			if fileExists(cf) {
				return cf
			}
		}
		dir = path.Dir(strings.TrimRight(dir, "/"))
		if dir == "/" {
			break
		}
	}
	return file
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

//读取旧项目遗留的地址和私钥得配对方式
//文件规则为序号,地址,私钥
func ReadOldeCsv(execlFileName string) error {
	logrus.Info("加载文件。。。")
	//file, err := os.Open("./config/old.csv")
	file, err := os.Open(execlFileName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	// 这个方法体执行完成后，关闭文件
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		// Read返回的是一个数组，它已经帮我们分割了，
		record, err := reader.Read()
		// 如果读到文件的结尾，EOF的优先级居然比nil还高！
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("记录集错误:", err)
			return err
		}
		keys := models.ImportKey2{}
		for i := 0; i < len(record); i++ {
			//fmt.Print(record[i] + " ")
			switch i {
			case 1:
				keys.Address = record[i]
			case 2:
				keys.Privkey = record[i]
			}
		}
		//fmt.Println("keys:", keys)
		db.KeyStore.Store(keys.Address, []byte(keys.Privkey))
		//log.Info(keys.Address + "----" + keys.Privkey)
		logrus.Info(keys.Address)
	}
	//os.Remove(execlFileName) //删除文件
	return nil

}
