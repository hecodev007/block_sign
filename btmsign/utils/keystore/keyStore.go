package keystore

import (
	"btmSign/common/conf"
	"btmSign/common/log"
	"crypto/aes"
	"crypto/cipher"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

//type Service struct {
//	aesKeyMap     map[string]string
//	encryptKeyMap map[string]string
//	publicKeyMap  map[string]string
//}

var aesKeyMap map[string]string
var encryptKeyMap map[string]string
var publicKeyMap map[string]string

func New() {
	aesKeyMap = make(map[string]string)
	encryptKeyMap = make(map[string]string)
	publicKeyMap = make(map[string]string)
	initKeyMap()
}

func initKeyMap() {
	paths := loadKeyFilePath()
	if paths == nil || len(paths) == 0 {
		// log.Error("Load private key error,maybe is can find .csv,please check you path")
		return
	}
	err := loadKeyFile(paths)
	if err != nil {

	}
}

func loadKeyFilePath() []string {
	var (
		paths []string
	)

	filepath.Walk("./"+conf.GetConfig().Csv.Dir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".csv") {
			if strings.Contains(path, "_a") || strings.Contains(path, "_b") {
				paths = append(paths, path)
			}
		}
		return nil
	})
	return paths
}

func loadKeyFile(paths []string) error {

	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		// 这个方法体执行完成后，关闭文件
		defer file.Close()
		reader := csv.NewReader(file)
		// reader.FieldsPerRecord = -1
		// idx:=1
		for {
			// log.Println(path,idx)
			// Read返回的是一个数组，它已经帮我们分割了，
			record, err := reader.Read()
			// 如果读到文件的结尾，EOF的优先级比nil高！
			if err == io.EOF {
				break
			} else if err != nil {
				log.Info(path)
				log.Info(reader.FieldsPerRecord)
				log.Error("记录集错误:", err)
				return nil
			}
			if strings.Contains(path, "_a") {
				encryptKeyMap[record[0]] = record[1]
			} else if strings.Contains(path, "_b") {
				aesKeyMap[record[0]] = record[1]
			}
			// idx++
		}
	}
	return nil
}

func GetKeyByAddress(address string) (string, error) {
	aesKey := aesKeyMap[address]
	encryptKey := encryptKeyMap[address]
	if aesKey == "" || encryptKey == "" {
		return "", fmt.Errorf("Load aes key or encrypt key is null,AES=[%s],ENCRYPT=[%s],Address=[%s]", aesKey,
			encryptKey, address)
	}
	privateBytes, err := AesBase64Crypt([]byte(encryptKey), []byte(aesKey), false)
	if err != nil {
		return "", err
	}
	wif := string(privateBytes)
	return wif, nil
}

func AesBase64Crypt(data, key []byte, encry bool) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var iv = key[:aes.BlockSize]
	if encry {
		crypted := make([]byte, len(data))
		cipher.NewCFBEncrypter(aesBlock, iv).XORKeyStream(crypted, data)
		return Base64Encode(crypted), nil
	} else {
		baseData, err := Base64Decode(data)
		if err != nil {
			return nil, err
		}
		crypted := make([]byte, len(baseData))
		cipher.NewCFBDecrypter(aesBlock, iv).XORKeyStream(crypted, baseData)
		return crypted, nil
	}
}

type Keystore map[string]map[string]*CsvKey

var KeysDB Keystore

//KeysDB 读写锁
var KeysLock *sync.RWMutex

func init() {
	KeysLock = new(sync.RWMutex)
	cfg := conf.GetConfig()
	if err := InitKeystore(cfg.Csv.Dir, cfg.Name); err != nil {
		panic(err.Error())
	}
}

func Have(keyDir, mchName, orderId string) bool {
	cvsFileA := path.Join(keyDir, mchName, fmt.Sprintf("%s_a.csv", orderId))
	if _, err := os.Stat(cvsFileA); err == nil {
		return true
	} else if os.IsExist(err) {
		return true
	} else {
		return false
	}
}
func InitKeystore(dirName, coinName string) error {

	if KeysDB == nil {
		KeysDB = make(Keystore)
	}

	//log.Infof("%s", dirName)
	//fmt.Println("dirName", dirName)

	dirAbsPath, err := filepath.Abs(dirName)
	if err != nil {
		log.Infof("don't find csv dir %s ", dirAbsPath)
		return fmt.Errorf("don't find csv dir %s ", dirAbsPath)
	}
	//fmt.Println("dirAbsPath", dirAbsPath)
	csvFiles, err := ListCsvFile(dirAbsPath)
	//fmt.Println("csvFiles", csvFiles)
	if err != nil {
		log.Infof("%s don't find csv file", dirAbsPath)
		return fmt.Errorf("%s don't find csv file", dirAbsPath)
	}

	for _, c := range csvFiles {
		parentDir := GetParentDirectory(c)
		pf, err := os.Stat(parentDir)
		if err != nil {
			log.Infof("%s Stat err : %v", parentDir, err)
			continue
		}

		fi, err := os.Stat(c)
		if err != nil {
			log.Infof("%s Stat err : %v", c, err)
			continue
		}
		dbkey := pf.Name() + fi.Name()[strings.Index(fi.Name(), "_"):]
		keys, err := ReadCsvFile(c, false)
		if err != nil {
			log.Infof("ReadCsvFile err : %v", err)
			continue
		}
		//fmt.Printf("%s  get keys success, len : %d", fi.Name(), len(keys))

		if _, ok := KeysDB[dbkey]; ok {
			//比较两个长度，range短的map
			if len(KeysDB[dbkey]) < len(keys) {
				KeysDB[dbkey], keys = keys, KeysDB[dbkey]
			}
			for k, v := range keys {
				KeysDB[dbkey][k] = v
			}
		} else {
			KeysDB[dbkey] = keys

		}
	}
	//fmt.Println(KeysDB)
	//for k, v := range KeysDB {
	//	fmt.Printf("%s: len :%d ", k, len(v))
	//}
	//
	//log.Infof("%s  InitKeystore success", dirAbsPath)

	return nil
}

func KeystoreGetKeyA(mchName, address string) (string, error) {
	KeysLock.RLock()
	defer KeysLock.RUnlock()

	fileName := fmt.Sprintf("%s_a.csv", mchName)
	//fmt.Println("fileName", fileName, address, KeysDB[fileName]["t1evM1XTsW1BayiT5qPqbfKU37bWb3rRWHc"], KeysDB[fileName])
	if keys, ok := KeysDB[fileName]; !ok {
		return "", fmt.Errorf("doesn't find keys for mch: %s", fileName)
	} else if v, ok := keys[address]; ok {
		return v.Key, nil
	}

	return "", fmt.Errorf("doesn't find key for address: %s", address)
}

func KeystoreGetKeyB(mchName, address string) (string, error) {
	KeysLock.RLock()
	defer KeysLock.RUnlock()

	fileName := fmt.Sprintf("%s_b.csv", mchName)

	if keys, ok := KeysDB[fileName]; !ok {
		return "", fmt.Errorf("can't find keys for mch: %s", fileName)
	} else if v, ok := keys[address]; ok {
		return v.Key, nil
	}

	return "", fmt.Errorf("can't find key for address: %s", address)
}

func GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*CsvKey, mchName, orderNo string) error {
	if len(cvsKeysA) == 0 || len(cvsKeysB) == 0 || len(cvsKeysC) == 0 || len(cvsKeysD) == 0 {
		return fmt.Errorf("csv don't allow empty")
	}

	mydir, _ := os.Getwd()
	csvdir := path.Join(mydir, "csv", mchName)
	if err := os.MkdirAll(csvdir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %s , err %v", path.Join(csvdir, mchName), err)
	}

	cvsFileA := path.Join(csvdir, fmt.Sprintf("%s_a.csv", orderNo))
	cvsFileB := path.Join(csvdir, fmt.Sprintf("%s_b.csv", orderNo))
	cvsFileC := path.Join(csvdir, fmt.Sprintf("%s_c.csv", orderNo))
	cvsFileD := path.Join(csvdir, fmt.Sprintf("%s_d.csv", orderNo))

	if err := WriteCsvFile(cvsKeysA, cvsFileA); err != nil {
		return fmt.Errorf("create file : %s , err: %v", cvsFileA, err)
	}

	if err := WriteCsvFile(cvsKeysB, cvsFileB); err != nil {
		return fmt.Errorf("create file : %s , err: %v", cvsFileB, err)
	}

	if err := WriteCsvFile(cvsKeysC, cvsFileC); err != nil {
		return fmt.Errorf("create file : %s , err: %v", cvsFileC, err)
	}

	if err := WriteCsvFile(cvsKeysD, cvsFileD); err != nil {
		return fmt.Errorf("create file : %s , err: %v", cvsFileD, err)
	}

	//异步加载
	go SubKeysDBfunc(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, mchName)
	return nil
}

//写入到全局变量KeysDB
func SubKeysDBfunc(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*CsvKey, mchName string) error {
	KeysLock.Lock()
	defer KeysLock.Unlock()
	fileNameA := fmt.Sprintf("%s_a.csv", mchName)
	fileNameB := fmt.Sprintf("%s_b.csv", mchName)
	fileNameC := fmt.Sprintf("%s_c.csv", mchName)
	fileNameD := fmt.Sprintf("%s_d.csv", mchName)
	if _, ok := KeysDB[fileNameA]; !ok {
		KeysDB[fileNameA] = make(map[string]*CsvKey)
	}
	if _, ok := KeysDB[fileNameB]; !ok {
		KeysDB[fileNameB] = make(map[string]*CsvKey)
	}
	if _, ok := KeysDB[fileNameC]; !ok {
		KeysDB[fileNameC] = make(map[string]*CsvKey)
	}
	if _, ok := KeysDB[fileNameD]; !ok {
		KeysDB[fileNameD] = make(map[string]*CsvKey)
	}
	for _, v := range cvsKeysA {
		KeysDB[fileNameA][v.Address] = v
	}
	for _, v := range cvsKeysB {
		KeysDB[fileNameB][v.Address] = v
	}
	for _, v := range cvsKeysC {
		KeysDB[fileNameC][v.Address] = v
	}
	for _, v := range cvsKeysD {
		KeysDB[fileNameD][v.Address] = v
	}

	return nil
}
