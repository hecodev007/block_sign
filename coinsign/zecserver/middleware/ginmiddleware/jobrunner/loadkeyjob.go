package jobrunner

import (
	"github.com/bamzi/jobrunner"
	"github.com/group-coldwallet/zecserver/util"
	"github.com/sirupsen/logrus"
	"strings"
)

var fileArr []string

//cronSpec cron表达式
//path 文件路径
//加载方法 loadMethod
func LoadKeyJob(cronSpec, path string, addressIndex int, loadMethod func(fileA, fileB string, addrIndex int)) {
	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	jobrunner.Schedule(cronSpec, LoadKey{Path: path, AddressIndex: addressIndex, LoadMethod: loadMethod})
}

// Job Specific Functions
type LoadKey struct {
	Path         string
	AddressIndex int
	LoadMethod   func(string, string, int)
}

//监听指定文件夹，把新文件加载进入内存
func (load LoadKey) Run() {
	//差异性文件集合
	diffFiles := make([]string, 0)
	logrus.Info("Listening folder")
	//
	files, err := util.GetAllFile(load.Path)
	if err != nil {
		logrus.Error(err)
		return
	}

	if len(fileArr) == len(files) {
		logrus.Info("File directory has not changed")
		return
	} else {
		logrus.Info("Discover new files")
		for _, file := range files {
			has := false
			for _, v := range fileArr {
				if file == v {
					//相同的剔除
					has = true
					break
				}
			}
			if !has {
				//把不相同的文件提取出来
				diffFiles = append(diffFiles, file)
			}
		}
		//校验a b文件是否存在
		for _, file := range diffFiles {
			// 解析文件名
			fileName := strings.Split(file, "_")
			//len(list) < 3 判断为新文件，过滤
			if len(fileName) < 3 || (fileName[1] != "a") {
				continue
			}
			//得到a文件，校验b文件是否存在
			fileBPath := strings.Replace(file, "_a_", "_b_", 1)
			has := util.IsFileExist(fileBPath)
			if !has {
				logrus.Info("FileB not exists")
			}
			//读取文件，加载进入内存
			load.LoadMethod(file, fileBPath, load.AddressIndex)
		}
		fileArr = make([]string, len(files))
		copy(fileArr, files)

	}
	return
}
