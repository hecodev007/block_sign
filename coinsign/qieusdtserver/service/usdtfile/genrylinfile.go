package usdtfile

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/config"
	"github.com/group-coldwalle/coinsign/qieusdtserver/db"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/group-coldwalle/coinsign/qieusdtserver/service"
	"github.com/group-coldwalle/coinsign/qieusdtserver/util"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	MaxPerLimit = 500
	MaxWorker   = 100
	ChanLen     = 50
)

const (
	defaultSleepDuration = 2 * time.Second
)

var (
	defaultFile = &FileConfig{}
)

type FileConfig struct {
	genAFile string
	genBFile string
	genCFile string
	genDFile string
}

func SetDefaultFileConfig(cfg *config.GlobalConfig) error {
	const (
		fileA = "usdt_a.csv"
		fileB = "usdt_b.csv"
		fileC = "usdt_c.csv"
		fileD = "usdt_d.csv"
	)

	fi, err := os.Stat(path.Join(cfg.GenFilePath, fileA))
	if err == nil {
		if fi.IsDir() {
			return errors.New("gen file A usdt_a is dir")
		}
	}
	fi, err = os.Stat(path.Join(cfg.GenFilePath, fileB))
	if err == nil {
		if fi.IsDir() {
			return errors.New("gen file B usdt_b is dir")
		}
	}
	fi, err = os.Stat(path.Join(cfg.GenFilePath, fileC))
	if err == nil {
		if fi.IsDir() {
			return errors.New("gen file C usdt_c is dir")
		}
	}
	fi, err = os.Stat(path.Join(cfg.GenFilePath, fileD))
	if err == nil {
		if fi.IsDir() {
			return errors.New("gen file D usdt_d is dir")
		}
	}
	defaultFile.genAFile = path.Join(cfg.GenFilePath, fileA)
	defaultFile.genBFile = path.Join(cfg.GenFilePath, fileB)
	defaultFile.genCFile = path.Join(cfg.GenFilePath, fileC)
	defaultFile.genDFile = path.Join(cfg.GenFilePath, fileD)
	return nil
}

func GenRylinkFile(addressinput *models.BatchAddressInput) (string, string, string, error) {
	return defaultFile.GenRylinkFile(addressinput)
}

//单线程执行
func GenRylinkFileBySingleThread(addressinput *models.BatchAddressInput) (string, string, string, error) {
	return defaultFile.GenRylinkFileBySingleThread(addressinput)
}

func (c *FileConfig) GenRylinkFile(addressinput *models.BatchAddressInput) (string, string, string, error) {

	//创建不存在文件
	_, err := os.Stat(c.genAFile)
	if err != nil {
		os.Create(c.genAFile)
	}
	_, err = os.Stat(c.genBFile)
	if err != nil {
		os.Create(c.genBFile)
	}
	_, err = os.Stat(c.genCFile)
	if err != nil {
		os.Create(c.genCFile)
	}
	_, err = os.Stat(c.genDFile)
	if err != nil {
		os.Create(c.genDFile)
	}

	//以追加的方式的打开文件
	fileA, err := os.OpenFile(c.genAFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_a file error: ", err)
		return "", "", "", err
	}
	defer fileA.Close()
	fileB, err := os.OpenFile(c.genBFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_b error: ", err)
		return "", "", "", err
	}
	defer fileB.Close()
	fileC, err := os.OpenFile(c.genCFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_c error: ", err)
		return "", "", "", err
	}
	defer fileC.Close()
	fileD, err := os.OpenFile(c.genDFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_d error: ", err)
		return "", "", "", err
	}
	defer fileD.Close()

	wa := csv.NewWriter(fileA) //创建一个新的写入文件流
	wb := csv.NewWriter(fileB)
	wc := csv.NewWriter(fileC)
	wd := csv.NewWriter(fileD)
	ch := make(chan *GenKey, ChanLen)
	km := NewGenKeyManager(addressinput.Num, ch)
	go km.Start()
	i := 0
	for k := range ch {
		i++
		wa.Write([]string{string(k.Address), string(k.AESPrivateKey)})
		wb.Write([]string{string(k.Address), string(k.AESKey)})
		wc.Write([]string{string(k.Address), string(k.PrivateKey)})
		wd.Write([]string{string(k.Address), ""})
	}
	fmt.Println("i:", i)
	wa.Flush()
	wb.Flush()
	wc.Flush()
	wd.Flush()

	return c.genAFile, c.genBFile, c.genDFile, nil
}

func (c *FileConfig) GenRylinkFileBySingleThread(addressinput *models.BatchAddressInput) (string, string, string, error) {

	//创建不存在文件
	_, err := os.Stat(c.genAFile)
	if err != nil {
		os.Create(c.genAFile)
	}
	_, err = os.Stat(c.genBFile)
	if err != nil {
		os.Create(c.genBFile)
	}
	_, err = os.Stat(c.genCFile)
	if err != nil {
		os.Create(c.genCFile)
	}
	_, err = os.Stat(c.genDFile)
	if err != nil {
		os.Create(c.genDFile)
	}
	//以追加的方式的打开文件
	fileA, err := os.OpenFile(c.genAFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_a file error: ", err)
		return "", "", "", err
	}
	defer fileA.Close()

	fileB, err := os.OpenFile(c.genBFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_b error: ", err)
		return "", "", "", err
	}
	defer fileB.Close()

	fileC, err := os.OpenFile(c.genCFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_c error: ", err)
		return "", "", "", err
	}
	defer fileC.Close()

	fileD, err := os.OpenFile(c.genDFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("Open usdt_d error: ", err)
		return "", "", "", err
	}
	defer fileD.Close()

	wa := csv.NewWriter(fileA) //创建一个新的写入文件流
	wb := csv.NewWriter(fileB)
	wc := csv.NewWriter(fileC)
	wd := csv.NewWriter(fileD)
	for i := int64(0); i < addressinput.Num; i++ {
		//name := addressinput.TagName + strconv.FormatInt(i, 10)
		info, err := service.CreateNewAddress()
		if err != nil {
			//跳过错误
			logrus.Errorf("createn new address error,numbers: %d, error:", i, err)
			continue
		}
		aesKey := util.RandBase64Key()
		aesPrivateKey, err := util.AesBase64Crypt([]byte(info.PrivateKey), aesKey, true)
		if err != nil {
			logrus.Infoln("util.AesBase64Crypt error: ", err)
			continue
		}
		wa.Write([]string{info.Address, string(aesPrivateKey)})
		wb.Write([]string{info.Address, string(aesKey)})
		wc.Write([]string{string(info.Address), string(info.PrivateKey)})
		wd.Write([]string{info.Address, ""})
	}
	wa.Flush()
	wb.Flush()
	wc.Flush()
	wd.Flush()
	return c.genAFile, c.genBFile, c.genDFile, nil
}

type GenKey struct {
	WorkerId      int
	KeyNum        int64
	PrivateKey    []byte
	PublicKey     []byte
	Address       []byte
	AESKey        []byte
	AESPrivateKey []byte
}

func NewGenKeyManager(num int64, outCh chan<- *GenKey) *GenKeyManager {
	var (
		nums      []int64
		remain    int64
		workerNum int64
	)
	workerNum = num / MaxPerLimit
	remain = num % MaxPerLimit
	if workerNum == 0 {
		workerNum = 1
		nums = []int64{remain}
	} else {
		if remain > 0 {
			workerNum++
		}
		if workerNum > MaxWorker {
			nums = make([]int64, MaxWorker, MaxWorker)
			perLimit := num / MaxWorker
			remain = num % MaxWorker
			var i int64
			for i = 0; i < MaxWorker; i++ {
				if remain > 0 {
					nums[i] = perLimit + 1
					remain--
				} else {
					nums[i] = perLimit
				}
			}
		} else if remain > 0 {
			var i int64
			nums = make([]int64, workerNum, workerNum)
			for i = 0; i < workerNum; i++ {
				if i+1 == workerNum {
					nums[i] = remain
				} else {
					nums[i] = MaxPerLimit
				}
			}
		} else {
			var i int64
			nums = make([]int64, workerNum, workerNum)
			for i = 0; i < workerNum; i++ {
				nums[i] = MaxPerLimit
			}
		}
	}
	return &GenKeyManager{
		MaxWokerNum:   workerNum,
		keyNums:       nums,
		OutCh:         outCh,
		WorkerFunc:    genKey,
		jobWg:         &sync.WaitGroup{},
		sleepDuration: defaultSleepDuration,
	}
}

type GenKeyManager struct {
	MaxWokerNum   int64
	keyNums       []int64
	OutCh         chan<- *GenKey
	WorkerFunc    func(id int, num int64, outCh chan<- *GenKey, wg *sync.WaitGroup, ctx context.Context)
	jobWg         *sync.WaitGroup
	sleepDuration time.Duration
	stopChan      chan struct{}
}

func (km *GenKeyManager) Start() error {
	if km.stopChan != nil {
		logrus.Info("GenKeyManager already started")
		return errors.New("GenKeyManager already started")
	}
	kns := len(km.keyNums)
	ctx, cancel := context.WithCancel(context.Background())
	km.stopChan = make(chan struct{})
	km.jobWg.Add(kns)

	for i := 0; i < kns; i++ {
		go km.WorkerFunc(i, km.keyNums[i], km.OutCh, km.jobWg, ctx)
	}
	stopChan := km.stopChan
	go func() {
		for {
			select {
			case <-stopChan:
				cancel()
				return
			default:
				time.Sleep(km.sleepDuration)
			}
		}
	}()
	km.AllDone()
	return nil
}

func (km *GenKeyManager) Stop() {
	if km.stopChan != nil {
		km.stopChan <- struct{}{}
	}
	km.AllDone()
}

func (km *GenKeyManager) AllDone() {
	km.jobWg.Wait()
	close(km.OutCh)
}

func genKey(id int, num int64, outCh chan<- *GenKey, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	if num <= 0 {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var (
				i           int64 = 1
				err         error
				addressinfo *models.AddressOutPut
				gk          *GenKey
			)
			for i = 1; i <= num; i++ {
				gk = &GenKey{WorkerId: id, KeyNum: i}

				addressinfo, err = service.CreateNewAddress()
				if err != nil {
					logrus.Infoln("CreateKey error: ", err)
					return
				}
				gk.PrivateKey = []byte(addressinfo.PrivateKey)
				gk.Address = []byte(addressinfo.Address)
				gk.PublicKey = []byte(addressinfo.PublicKey)
				gk.AESKey = util.RandBase64Key()
				if gk.AESPrivateKey, err = util.AesBase64Crypt(gk.PrivateKey, gk.AESKey, true); err != nil {
					logrus.Infoln("util.AesBase64Crypt error: ", err)
					return
				}
				outCh <- gk
			}
			return
		}
	}
}

//
var (
	EncryptWifMap map[string]string = make(map[string]string)
	WifKeyListMap map[string]string = make(map[string]string)
)

//读取指定
func ReadNewFolder(folderpath string) {

	files, err := util.GetAllFile(folderpath)
	if err != nil {
		logrus.Infof("加载文件异常：%s", err.Error())
	}
	for _, fileName := range files {
		list := strings.Split(fileName, "_")
		//len(list) < 3 判断为新文件，过滤
		if len(list) < 3 || (list[1] != "a" && list[1] != "b") {
			continue
		}
		filepath := fileName

		logrus.Infof("加载新版本地址文件：%s", filepath)
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
		//fmt.Println(k, ":::", string(prv))
		//logrus.Printf("解密地址：%s，密钥:%s", k, string(prv))
		logrus.Infof("解密地址：%s", k)
		//db.SetKeys(k, string(prv))
		db.SetKeys(k, string(prv))
	}
}

//新文件地址在下标0 ,旺旺旧文件地址下标1
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

//新文件地址在下标0 ,旺旺旧文件地址下标1
//func readUsbBConfigNew(usb_b string) {
//	if usb_b == "" {
//		return
//	}
//
//	cntb, err := ioutil.ReadFile(usb_b)
//	if err != nil {
//		panic(err.Error())
//	}
//	r2 := csv.NewReader(strings.NewReader(string(cntb)))
//	keyList, _ := r2.ReadAll()
//
//	for i := 0; i < len(keyList); i++ {
//		//fmt.Println(keyList[i][0], keyList[i][1])
//		WifKeyListMap[keyList[i][0]] = keyList[i][1]
//	}
//}

//新文件地址在下标0 ,旺旺旧文件地址下标1
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

//大文件读取
func ReadFileCsv(filePath string, handle func(string)) error {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)

	for {
		line, _, err := buf.ReadLine()
		lineStr := strings.TrimSpace(string(line))
		handle(lineStr)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		return nil
	}
}
