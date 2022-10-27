package loadkeyservice

import (
	"encoding/csv"
	"github.com/group-coldwallet/nep5server/model/global"
	"github.com/group-coldwallet/nep5server/service"
	"github.com/group-coldwallet/nep5server/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

type LoadService struct {
}

func (srv *LoadService) ReadNewFolder(folderpath string) {
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
			readUsbAFile(filepath)
		} else if list[1] == "b" {
			readUsbBFile(filepath)
		}
	}

	for k, _ := range EncryptWifMap {
		kk, _ := util.Base64Decode([]byte(EncryptWifMap[k]))
		prv, _ := util.AesCrypt(kk, []byte(WifKeyListMap[k]), false)
		logrus.Infof("ReadNewFolder 解密地址：%s", k)
		global.SetValue(k, string(prv))
	}
}

func NewLoadService() service.LoadKeyService {
	return &LoadService{}
}

var (
	EncryptWifMap map[string]string = make(map[string]string)
	WifKeyListMap map[string]string = make(map[string]string)
)

//新文件地址在下标0
func readUsbAFile(usb_a string) {
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

//新文件地址在下标0
func readUsbBFile(usb_b string) {
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
