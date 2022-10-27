package util

import (
	"os"
	"path/filepath"
)

//读取目录下以及子目录所有文件,s为接收的数组
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
