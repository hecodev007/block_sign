package keystore

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"starSign/common/conf"
	"starSign/common/log"
	"strings"
	"sync"
)

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
		//log.Infof("%s  get keys success, len : %d", fi.Name(), len(keys))

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
	//	log.Infof("%s: len :%d", k, len(v))
	//}

	log.Infof("%s  InitKeystore success", dirAbsPath)

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
