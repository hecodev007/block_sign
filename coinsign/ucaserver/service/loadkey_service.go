package service

import (
	"bufio"
	"fmt"
	"github.com/group-coldwallet/ucaserver/model/global"
	"github.com/group-coldwallet/ucaserver/pkg/util"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type LoadService struct {
}

var (
	EncryptWifMap map[string]string = make(map[string]string)
	WifKeyListMap map[string]string = make(map[string]string)
)

//todo：方法有点乱，需要进行精简

//读取指定文件夹的地址
func (s *LoadService) ReadNewFolder(folderpath string) {
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
		//logrus.Infof("ReadNewFolder 解密地址：%s", k)
		global.SetValue(k, string(prv))
	}
	logrus.Infof("加载地址数量：%d", len(global.KeyStore))
}

//读取指定文件
func (s *LoadService) ReadFile(fileAPath, fileBPath string) {
	if fileAPath == "" || fileBPath == "" {
		logrus.Error("ReadFile加载指定地址文件,empty path")
		return
	}
	logrus.Infof("ReadFile 加载指定地址文件，A:%s,B:%s", fileAPath, fileBPath)
	readUsbAConfigNew(fileAPath)
	readUsbBConfigNew(fileBPath)
	for k, _ := range EncryptWifMap {
		kk, _ := util.Base64Decode([]byte(EncryptWifMap[k]))
		prv, err := util.AesCrypt(kk, []byte(WifKeyListMap[k]), false)
		if err != nil {
			logrus.Infof("decode address error：%v,address：%s", err, k)
			continue
		}
		//logrus.Infof("ReadFile 解密地址：%s,私钥：%s", k, prv)
		logrus.Infof("ReadFile 解密地址：%s", k)
		global.SetValue(k, string(prv))
	}
}

//新文件地址在下标0
func readUsbAConfigNew(usb_a string) {
	if usb_a == "" {
		return
	}

	f, err := os.Open(usb_a)
	defer f.Close()
	if err != nil {
		panic(err.Error())
	}
	buf := bufio.NewReader(f)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err.Error())
		}
		lineArr := strings.Split(string(line), ",")
		if len(lineArr) < 2 {
			panic(fmt.Sprintf("文件长度不对:%s", string(line)))
		}
		EncryptWifMap[lineArr[0]] = lineArr[1]
	}

}

func readUsbBConfigNew(usb_b string) {
	if usb_b == "" {
		return
	}

	f, err := os.Open(usb_b)
	defer f.Close()
	if err != nil {
		panic(err.Error())
	}
	buf := bufio.NewReader(f)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err.Error())
		}
		lineArr := strings.Split(string(line), ",")
		if len(lineArr) < 2 {
			panic(fmt.Sprintf("文件长度不对:%s", string(line)))
		}
		WifKeyListMap[lineArr[0]] = lineArr[1]
	}
}
