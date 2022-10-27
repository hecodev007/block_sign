package main

import (
	"mwSign/common/conf"
	"mwSign/common/log"
	"mwSign/utils/keystore"
	"mwSign/utils/mw"
	"strings"
)

func main() {
	//初始化log
	cfg := conf.GetConfig()
	log.InitLogger(cfg.Log.Level, cfg.Mode, cfg.Log.Formatter, cfg.Log.OutFile, cfg.Log.ErrFile)

	csvFiles, err := keystore.ListCsvFile("./csv/")
	//fmt.Println("csvFiles", csvFiles)
	if err != nil {
		panic(err.Error())
	}

	for _, c := range csvFiles {
		if !strings.HasSuffix(c, "c.csv") {
			continue
		}
		filename := strings.ReplaceAll(c, "/", "_")
		filename = strings.ReplaceAll(filename, "c.csv", "e.csv")
		log.Info(filename)
		//parentDir := keystore.GetParentDirectory(c)

		csvkeys, err := keystore.ReadCsvFile(c, false)

		if err != nil {
			panic(err.Error())
		}
		keys := make([]*keystore.CsvKey, 0)
		for _, key := range csvkeys {
			csvKey := new(keystore.CsvKey)
			//log.Infof("file:%s  addr:%v pri:%v", k2, key.Address, key.Key)
			csvKey.Key, _ = mw.PrivateToPub(key.Key)
			csvKey.Address, _ = mw.PubkeyToAddr(csvKey.Key)
			if csvKey.Address != key.Address {
				panic("")
			}
			keys = append(keys, csvKey)
		}
		keystore.WriteCsvFile(keys, filename)
	}

	//for k, v := range keystore.KeysDB {
	//	if !strings.HasSuffix(k, "c.csv") {
	//		continue
	//	}
	//	filename := strings.ReplaceAll(k, "c.csv", "e.csv")
	//	keys := make([]*keystore.CsvKey, 0)
	//	for k2, key := range v {
	//		csvKey := new(keystore.CsvKey)
	//		log.Infof("file:%s  addr:%v pri:%v", k+" "+k2, key.Address, key.Key)
	//		csvKey.Key, _ = mw.PrivateToPub(key.Key)
	//		csvKey.Address, _ = mw.PubkeyToAddr(csvKey.Key)
	//		if csvKey.Address != key.Address {
	//			panic("")
	//		}
	//		keys = append(keys, csvKey)
	//	}
	//	keystore.WriteCsvFile(keys, filename)
	//}

}
