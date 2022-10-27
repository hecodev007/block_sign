package api

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/db"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/group-coldwalle/coinsign/qieusdtserver/util"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	chanLen   = 50
	workerNum = 4
)

type fileType int

const (
	file_A fileType = iota
	file_B
)

type keyData struct {
	fType   fileType
	address string
	enKey   []byte
}

var (
	checkSignal int32 = 0
	keyDataChan       = make(chan *keyData, chanLen)
	processG          = sync.WaitGroup{}
	isWorkerRun       = false
)

func workerRun(ch <-chan *keyData) {
	for i := 0; i < workerNum; i++ {
		go procKeyData(ch)
	}
	isWorkerRun = true
}

//get
//获取私钥
func GetPrivateKey(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		address string
		privKey string
		ok      bool
	)
	r.ParseForm()
	//获取get 的参数------------------------
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		HttpError(w, []byte("GetPrivateKey: parse form error"), nil)
		return
	} else {
		address = queryForm["address"][0]
		if len(address) == 0 {
			HttpError(w, []byte("GetPrivateKey: please input address"), nil)
			return
		}
	}
	//---------------------------------
	privKey, ok = db.GetPrivKey(address)
	if !ok {
		HttpError(w, []byte("GetPrivateKey: no private key"), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"private_key": privKey})
}

//post
//导入私钥，同时导入客户端
func ImportPrivKey(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		importKey  *models.ImportKey
		importKeyB []byte
		privKey    []byte
	)
	if importKeyB, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(importKeyB) <= 0 {
		HttpError(w, []byte("no import key"), nil)
		return
	}
	importKey, err = util.DecodeImportKey(importKeyB)
	if err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	pp, _ := util.Base64Decode([]byte(importKey.AesPrivkey))
	privKey, err = util.AesCrypt(pp, []byte(importKey.AesKey), false)

	db.SetKeys(importKey.Address, string(privKey))
	//HttpOK(w, map[string]interface{}{"private_key": string(privKey), "address": importKey.Address})
	HttpOK(w, map[string]interface{}{"address": importKey.Address})
}

//post
//导入私钥,不导入客户端
func ImportPrivKey2(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		importKey  *models.ImportKey2
		importKeyB []byte
	)
	if importKeyB, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(importKeyB) <= 0 {
		HttpError(w, []byte("no import key"), nil)
		return
	}
	importKey, err = util.DecodeImportKey2(importKeyB)
	if err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	//db.KeyStore.Store(importKey.Address, []byte(importKey.Privkey))
	db.SetKeys(importKey.Address, importKey.Privkey)
	//HttpOK(w, map[string]interface{}{"private_key": importKey.Privkey, "address": importKey.Address})
	HttpOK(w, map[string]interface{}{"address": importKey.Address})
}

//post
//导入私钥,不导入客户端
func ImportPrivKey3(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		importKey  *models.ImportKey2
		importKeyB []byte
	)
	if importKeyB, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(importKeyB) <= 0 {
		HttpError(w, []byte("no import key"), nil)
		return
	}
	importKey, err = util.DecodeImportKey2(importKeyB)
	if err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	//db.KeyStore.Store(importKey.Address, []byte(importKey.Privkey))
	db.SetKeys(importKey.Address, importKey.Privkey)
	//HttpOK(w, map[string]interface{}{"private_key": importKey.Privkey, "address": importKey.Address})
	HttpOK(w, map[string]interface{}{"address": importKey.Address})
}

//post
//删除私钥
func RemovePrivKey(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		data  []byte
		input *models.RemoveKeyInput
	)
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(data) <= 0 {
		HttpError(w, []byte("no remove key input"), nil)
		return
	}

	if input, err = util.DecodeRemoveKeyInput(data); err != nil {
		HttpError(w, []byte("decode RemoveKey error: "+err.Error()), nil)
		return
	}
	//db.DelKeys(input.Address)
	db.DelKeys(input.Address)
	HttpOK(w, nil)
}

//post
//创建私钥
func CreateKey(w http.ResponseWriter, r *http.Request) {
	//var (
	//	privKey []byte
	//	pubKey  []byte
	//	err     error
	//)
	//if privKey, pubKey, err = command.CreateKey(); err != nil {
	//	HttpError(w, []byte("wallet create key error: "+err.Error()), nil)
	//	return
	//}
	//HttpOK(w, map[string]interface{}{"private_key": string(privKey), "public_key": string(pubKey)})
}

//post
//批量创建私钥并保存到文件
func BatchCreateKey(w http.ResponseWriter, r *http.Request) {
	//var (
	//	data     []byte
	//	err      error
	//	eosAFile string
	//	eosBFile string
	//	eosCFile string
	//	num      int64
	//)
	//if data, err = ioutil.ReadAll(r.Body); err != nil {
	//	HttpError(w, []byte(err.Error()), nil)
	//	return
	//}
	//if len(data) <= 0 {
	//	HttpError(w, []byte("no gen num"), nil)
	//	return
	//}
	//if num, err = util.DecodeGenNum(data); err != nil {
	//	HttpError(w, []byte("decode num error: "+err.Error()), nil)
	//	return
	//}
	//if eosAFile, eosBFile, eosCFile, err = command.GenRylinkFile(num); err != nil {
	//	HttpError(w, []byte("GenRylinkFile error: "+err.Error()), nil)
	//	return
	//}
	//HttpOK(w, map[string]interface{}{"eos_a": eosAFile, "eos_b": eosBFile, "eos_c": eosCFile})
}

//获取文件
func DownLoad(w http.ResponseWriter, r *http.Request) {
	//这里写死,以后要改回配置
	had := http.StripPrefix("/download/", http.FileServer(http.Dir(cfg.GenFilePath)))
	had.ServeHTTP(w, r)
}

//post
func Upload(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		data []byte
	)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if !atomic.CompareAndSwapInt32(&checkSignal, 0, 1) {
		logrus.Error("the last upload file is processing")
		data = []byte("the last upload file is processing")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Write(data)
		return
	}
	defer atomic.StoreInt32(&checkSignal, 0)
	if err = r.ParseMultipartForm(51200000); err != nil {
		logrus.Errorf("upload file size error, message: %s", err.Error())
		data = []byte(err.Error())
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Write(data)
		return
	}

	if !isWorkerRun {
		workerRun(keyDataChan)
	}

	processG.Add(2)
	go func() {
		readFile(r, file_A)
		processG.Done()
	}()
	go func() {
		readFile(r, file_B)
		processG.Done()
	}()
	processG.Wait()

	data = []byte("upload file is done.")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func readFile(r *http.Request, fType fileType) {
	var (
		name    string
		f       multipart.File
		idx     int
		rowData []byte
		err     error
	)
	if fType == file_A {
		name = "usdt_a"
	} else {
		name = "usdt_b"
	}
	f, _, err = r.FormFile(name)
	if err != nil {
		logrus.Error("pase ", name, " file error, message: ", err)
		return
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		rowData, _, err = reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				//is end
				return
			}
			logrus.Error("read ", name, " file error, message: ", err)
			return
		}
		if len(rowData) == 0 {
			continue
		}
		//猜想因为bufio的reader内部用的缓存会重复利用,导致有时候在reader.ReadLine时
		//数据会互相覆盖,导致在并发的时候数据错乱,所以出来后copy一份防止并发
		idx = bytes.Index(rowData, []byte(","))
		if idx <= 0 {
			continue
		}
		kd := &keyData{
			fType:   fType,
			address: string(rowData[:idx]),
			enKey:   make([]byte, len(rowData)-idx-1, len(rowData)-idx-1),
		}
		copy(kd.enKey, rowData[idx+1:])
		keyDataChan <- kd
		//logrus.Infoln(string(rowData))
	}
}

func procKeyData1(ch <-chan *keyData) {
	var (
		//storeValue interface{}
		value   string
		privKey []byte
		isLoad  bool
		err     error
	)
	for kd := range ch {
		///do 解密工作
		_, isLoad = db.KeyStore.LoadOrStore(kd.address, kd.enKey)
		if isLoad {
			value, _ = db.GetPrivKey(kd.address)
			if kd.fType == file_A {
				//那么存的key是加密密钥
				if privKey, err = util.Base64Decode(kd.enKey); err != nil {
					logrus.Error("Base64Decode error,fileType: ", kd.fType, "  pubKey: ", kd.address, " enKey: ", string(kd.enKey), " value: ", string(value), " message: ", err)
					return
				}
				privKey, err = util.AesCrypt(privKey, []byte(value), false)
			} else {
				//flie_B
				//那么存的key是加密后的私钥
				if privKey, err = util.Base64Decode([]byte(value)); err != nil {
					logrus.Error("Base64Decode error,fileType: ", kd.fType, "  pubKey: ", kd.address, " enKey: ", string(kd.enKey), " value: ", string(value), " message: ", err)
					return
				}
				privKey, err = util.AesCrypt(privKey, kd.enKey, false)
			}
			if err != nil {
				logrus.Error("AesDecrypt error,fileType: ", kd.fType, "  pubKey: ", kd.address, " enKey: ", string(kd.enKey), " value: ", string(value), " message: ", err)
				return
			}
			//fmt.Println("kd.pubKey:", kd.pubKey)
			fmt.Println("privKey:", string(privKey))
			db.SetKeys(kd.address, string(privKey))
			//db.KeyStore.Store(kd.address, privKey)

		}
	}
}

var AMap sync.Map
var BMap sync.Map

//var AMap = make(map[string]string, 0)
//var BMap = make(map[string]string, 0)

func procKeyData(ch <-chan *keyData) {
	var (
		privKey []byte
		aesStr  []byte
		err     error
	)
	for kd := range ch {
		if kd.fType == file_A {
			AMap.Store(kd.address, kd.enKey)
			//AMap[kd.address] = string(kd.enKey)
		} else {
			BMap.Store(kd.address, kd.enKey)
			//BMap[kd.address] = string(kd.enKey)
		}

		_, has := db.GetPrivKey(kd.address)
		if has {
			//已经存在无需加载
			//logrus.Printf("已经存在,address:%s,prv:%s", kd.address, string(prv))
			logrus.Infof("已经存在,address:%s", kd.address)
			continue
		}
		//解密
		aesVal, _ := AMap.Load(kd.address)
		aseKey, _ := BMap.Load(kd.address)

		if aesVal != nil && aseKey != nil {
			if aesStr, err = util.Base64Decode(aesVal.([]byte)); err != nil {
				logrus.Errorln("Base64Decode error,fileA  address: ", kd.address, " message: ", err)
				return
			}
			privKey, err = util.AesCrypt(aesStr, aseKey.([]byte), false)
			if err != nil {
				logrus.Errorf("AesDecrypt error message: %s", err.Error())
				return
			}
			//db.KeyStore.Store(kd.address, privKey)
			db.SetKeys(kd.address, string(privKey))
			//logrus.Printf("解密,address:%s,prv:%s", kd.address, string(privKey))
			logrus.Infof("解密,address:%s", kd.address)
		}

	}
}

//只能由主进程调用一次
func StopUploadJob() {
	if keyDataChan != nil {
		close(keyDataChan)
	}
}
