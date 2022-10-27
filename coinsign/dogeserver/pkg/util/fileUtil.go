package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//读取目录下以及子目录所有文件
func GetAllFile(pathname string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(pathname,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if f.IsDir() {
				//fmt.Println("dir:", path)
				return nil
			}
			//fmt.Println("file:", path)
			files = append(files, path)
			return nil
		})
	return files, err
}

//判断文件是否存在：存在，返回true，否则返回false
func IsFileExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func FileCopy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
