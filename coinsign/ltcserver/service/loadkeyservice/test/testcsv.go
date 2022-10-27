package main

import (
	"encoding/csv"
	"github.com/group-coldwallet/ltcserver/model/global"
	"github.com/group-coldwallet/ltcserver/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

func main() {
	ReadNewFolder("csv")
}

//读取指定文件夹的地址
func ReadNewFolder(folderpath string) {
	files, err := util.GetAllFile(folderpath)
	if err != nil {
		logrus.Infof("ReadNewFolder 加载文件异常,path:%s,err:%v", folderpath, err)
	}
	for _, fileName := range files {
		list := strings.Split(fileName, "_")
		//len(list) < 3 判断为新文件，过滤
		if len(list) < 3 || (list[1] != "a" && list[1] != "b") {
			continue
		}
		filepath := fileName

		logrus.Infof("ReadNewFolder 加载新版本地址文件：%s", filepath)
		// 读取配置文件
		if list[1] == "a" {
			readUsbAConfigNew(filepath)
		} else if list[1] == "b" {
			readUsbBConfigNew(filepath)
		}
	}

	for k, _ := range EncryptWifMap {
		kk, _ := util.Base64Decode([]byte(EncryptWifMap[k]))
		prv, _ := util.AesCrypt(kk, []byte(WifKeyListMap[k]), false)
		logrus.Infof("ReadNewFolder 解密地址：%s，私钥：%s", k, string(prv))
		global.SetValue(k, string(prv))
	}
}

//读取指定文件
func ReadFile(fileAPath, fileBPath string) {
	if fileAPath == "" || fileBPath == "" {
		logrus.Error("ReadFile加载指定地址文件,empty path")
		return
	}
	logrus.Infof("ReadFile 加载指定地址文件，A:%s,B:%s", fileAPath, fileBPath)
	encryptWifMap := readUsbFileA(fileAPath)
	wifKeyListMap := readUsbFileB(fileBPath)
	for k, _ := range encryptWifMap {
		kk, err := util.Base64Decode([]byte(encryptWifMap[k]))
		if err != nil {
			logrus.Infof("decode address error：%v,address：%s", err, k)
			continue
		}
		prv, err := util.AesCrypt(kk, []byte(wifKeyListMap[k]), false)
		if err != nil {
			logrus.Infof("decode address error：%v,address：%s", err, k)
			continue
		}
		//logrus.Infof("ReadFile 解密地址：%s,私钥：%s", k, prv)
		logrus.Infof("ReadFile 解密地址：%s", k)
		global.SetValue(k, string(prv))
	}

}

//新文件地址在下标0 ,旺旺旧文件地址下标1
func readUsbAConfigNew(usb_a string) {
	if usb_a == "" {
		return
	}

	cntb, err := ioutil.ReadFile(usb_a)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	encryptList, _ := r2.ReadAll()

	for i := 0; i < len(encryptList); i++ {
		EncryptWifMap[encryptList[i][0]] = encryptList[i][1]
	}
}

//新文件地址在下标0 ,旺旺旧文件地址下标1
func readUsbFileA(usb_a string) map[string]string {
	encryptWifMap := make(map[string]string)
	if usb_a == "" {
		return nil
	}

	cntb, err := ioutil.ReadFile(usb_a)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	encryptList, _ := r2.ReadAll()

	for i := 0; i < len(encryptList); i++ {
		encryptWifMap[encryptList[i][1]] = encryptList[i][0]
	}
	return encryptWifMap
}

//新文件地址在下标0 ,旺旺旧文件地址下标1
func readUsbFileB(usb_b string) map[string]string {
	wifKeyListMap := make(map[string]string)
	if usb_b == "" {
		return nil
	}

	cntb, err := ioutil.ReadFile(usb_b)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	keyList, _ := r2.ReadAll()

	for i := 0; i < len(keyList); i++ {
		//fmt.Println(keyList[i][0], keyList[i][1])
		wifKeyListMap[keyList[i][1]] = keyList[i][0]
	}
	return wifKeyListMap
}

//新文件地址在下标0 ,旺旺旧文件地址下标1
func readUsbBConfigNew(usb_b string) {
	if usb_b == "" {
		return
	}

	cntb, err := ioutil.ReadFile(usb_b)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	keyList, _ := r2.ReadAll()

	for i := 0; i < len(keyList); i++ {
		//fmt.Println(keyList[i][0], keyList[i][1])
		WifKeyListMap[keyList[i][0]] = keyList[i][1]
	}
}

//自动读取以前遗留的旧文件，注意旧文件address在下标1位置
//旺旺项目读取方法迁移到这里
var (
	EncryptWifMap map[string]string = make(map[string]string)
	WifKeyListMap map[string]string = make(map[string]string)
)

//读取所有
func ReadOleFolder(folderpath string) {
	// 读取csv目录
	files, _ := ioutil.ReadDir(folderpath)
	for _, file := range files {
		if !file.IsDir() {
			// 解析文件名
			list := strings.Split(file.Name(), "_")
			//len(list) < 3 判断为新文件，过滤
			if len(list) < 3 || (list[1] != "a" && list[1] != "b") {
				continue
			}
			filepath := folderpath + "/" + file.Name()
			// 读取配置文件
			logrus.Infof("ReadOleFolder 加载旧版本地址文件：%s", filepath)
			if list[1] == "a" {
				//controllers.ReadUsbAConfig(filepath)
				readUsbAConfig(filepath)
			} else if list[1] == "b" {
				readUsbBConfig(filepath)
			}
		}
	}
	for k, _ := range EncryptWifMap {
		kk, _ := util.Base64Decode([]byte(EncryptWifMap[k]))
		prv, _ := util.AesCrypt(kk, []byte(WifKeyListMap[k]), false)
		logrus.Infof("解密地址：%s", k)
		global.SetValue(k, string(prv))
	}

}

func readUsbAConfig(usb_a string) {
	if usb_a == "" {
		return
	}

	cntb, err := ioutil.ReadFile(usb_a)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	encryptList, _ := r2.ReadAll()

	for i := 0; i < len(encryptList); i++ {
		//fmt.Println(encryptList[i][0], encryptList[i][1])
		EncryptWifMap[encryptList[i][1]] = encryptList[i][0]
	}
}

func readUsbBConfig(usb_b string) {
	if usb_b == "" {
		return
	}

	cntb, err := ioutil.ReadFile(usb_b)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	keyList, _ := r2.ReadAll()

	for i := 0; i < len(keyList); i++ {
		//fmt.Println(keyList[i][0], keyList[i][1])
		WifKeyListMap[keyList[i][1]] = keyList[i][0]
	}
}
