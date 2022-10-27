package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type CsvData map[string]string

func ReadCsvFile(fileName string) ([][]string, error) {

	var res [][]string
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("read csv file %v", err)
		return nil, err
	}
	// 返回的是csv.Reader
	r := csv.NewReader(strings.NewReader(string(data)))
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}

		res = append(res, line)
	}

	return res, nil
}

// 返回任何可能发生的错误
func WriteCsvFile(cvsKeys [][]string, fileName string) error {

	csvFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Open file : %s, error: %v", fileName, err)
		return err
	}
	defer csvFile.Close()

	n := csv.NewWriter(csvFile)
	for _, cvsKey := range cvsKeys {
		err := n.Write(cvsKey)
		if err != nil {
			return err
		}
	}

	n.Flush()
	return n.Error()
}
