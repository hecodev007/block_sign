package main

import (
	"encoding/hex"
	"fmt"
	"moacSign/common/keystore"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main2() {
	file_old := "csv"
	dirAbsPath, err := filepath.Abs("./" + file_old)
	if err != nil {
		panic(err.Error())
	}
	csvFiles, err := keystore.ListCsvFile(dirAbsPath)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Println(csvFiles)
	for _, acsv := range csvFiles {

		if !strings.HasSuffix(acsv, "_a.csv") {
			continue
		}
		fmt.Println(acsv)
		bcsv := strings.Replace(acsv, "a.csv", "b.csv", 1)

		keylist := make([]*keystore.CsvKey, 0)
		bkeylist := make([]*keystore.CsvKey, 0)
		akeys, err := keystore.ReadCsvFile(acsv, false)
		if err != nil {
			panic(err.Error())
		}
		bkeys, err := keystore.ReadCsvFile(bcsv, false)
		for _, v := range akeys {
			//fmt.Println(v.Address)
			akey, err := keystore.Base64Decode([]byte(v.Key))
			if err != nil {
				panic(err.Error())
			}
			pri, err := keystore.AesCryptCfb([]byte(akey), []byte(bkeys[v.Address].Key), false)
			if err != nil {
				panic(err.Error())
			}

			if len(pri) == 64 {
				pri := []byte(hex.EncodeToString(pri))
				tKey, err := keystore.AesBase64CryptCfb([]byte(pri), []byte(bkeys[v.Address].Key), true)
				if err != nil {
					panic(err.Error())
				}
				v.Key = string(tKey)

			}
			keylist = append(keylist, v)
			bkeylist = append(bkeylist, bkeys[v.Address])
		}
		anewpath := strings.Replace(acsv, "/"+file_old+"/", "/csvNew/", 1)
		bnewpath := strings.Replace(bcsv, "/"+file_old+"/", "/csvNew/", 1)

		if err := os.MkdirAll(path.Dir(anewpath), os.ModePerm); err != nil {
			panic(err.Error())
		}
		err = keystore.WriteCsvFile(keylist, anewpath)
		if err != nil {
			panic(err.Error())
		}
		err = keystore.WriteCsvFile(bkeylist, bnewpath)
		if err != nil {
			panic(err.Error())
		}
	}

}
