package keystore

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"zenSign/common/log"

	"os"
	"strings"
)

//toLower,地址有没有大小写的区别
func ReadCsvFile(fileName string, toLower bool) (map[string]*CsvKey, error) {

	var cvsKeys = make(map[string]*CsvKey)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("read csv file %v", err)
		return nil, err
	}
	//返回的是csv.Reader
	r := csv.NewReader(strings.NewReader(string(data)))
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		cvsKey := &CsvKey{}
		if len(line) > 0 {
			cvsKey.Address = line[0]
		}
		if len(line) > 1 {
			cvsKey.Key = line[1]
		}
		if toLower {
			cvsKey.Address = strings.ToLower(cvsKey.Address)
		}
		cvsKeys[cvsKey.Address] = cvsKey
		//cvsKeys = append(cvsKeys, cvsKey)
	}

	return cvsKeys, nil
}

// 返回任何可能发生的错误
func WriteCsvFile(cvsKeys []*CsvKey, fileName string) error {

	csvFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Infof("Open file : %s, error: %v", fileName, err)
		return err
	}
	defer csvFile.Close()

	n := csv.NewWriter(csvFile)
	for _, cvsKey := range cvsKeys {
		err := n.Write([]string{cvsKey.Address, cvsKey.Key})
		if err != nil {
			return err
		}
	}

	n.Flush()
	return n.Error()
}
