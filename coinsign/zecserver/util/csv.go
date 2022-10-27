package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

//import "os"
//
//func writeCSV(strFile, strFiledKey, strFiledValue string) {
//	file, err := os.OpenFile(strFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		log.Fatal(err)
//	}
//	w := csv.NewWriter(file)
//	w.Write([]string{strFiledKey, strFiledValue})
//	w.Flush()
//	err = file.Close()
//	if err != nil {
//		log.Fatal(err)
//	}
//	//fmt.Printf("write csv file ok")
//}

//读取csv,返回某一列的内容，强制转换为string
func ReadCsv(execlFileName string, index int) ([]string, error) {
	result := make([]string, 0)
	file, err := os.Open(execlFileName)
	if err != nil {
		return nil, err
	}
	// 这个方法体执行完成后，关闭文件
	defer file.Close()
	reader := csv.NewReader(file)
	for {
		// Read返回的是一个数组，它已经帮我们分割了，
		record, err := reader.Read()
		// 如果读到文件的结尾，EOF的优先级居然比nil还高！
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("记录集错误:", err)
			return nil, err
		}
		if index < 0 || (len(record)-1) < index {
			return nil, fmt.Errorf("error index:%d", index)
		}
		result = append(result, record[index])
	}
	return result, nil
}

//解码
func DecodeAesToStr(params string, key string) string {
	if dst, err := Base64Decode([]byte(params)); err == nil {
		if result, err := AesCrypt(dst, []byte(key), false); err == nil {
			return string(result)
		}
	}
	return ""
}
