package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

//读取csv,返回某一列的内容，强制转换为string
//fileName 全路径文件名
//index 列下标，从0开始
func ReadCsv(fileName string, index int) ([]string, error) {
	result := make([]string, 0)
	file, err := os.Open(fileName)
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
			return nil, err
		}
		if index < 0 || (len(record)-1) < index {
			return nil, fmt.Errorf("error index:%d", index)
		}
		result = append(result, record[index])
	}
	return result, nil
}
