package util

import (
	"fmt"
	"os"
)

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//单层级创建
func CreateDir(path string) (bool, error) {
	exist, err := PathExists(path)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return false, err
	}

	if exist {
		fmt.Printf("has dir![%v]\n", path)
		return true, nil
	} else {
		fmt.Printf("no dir![%v]\n", path)
		// 创建文件夹
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
			return false, err
		} else {
			fmt.Printf("mkdir success!\n")
			return true, nil
		}
	}
}

//多层级创建
func CreateDirAll(path string) (bool, error) {
	exist, err := PathExists(path)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return false, err
	}

	if exist {
		fmt.Printf("has dir![%v]\n", path)
		return true, nil
	} else {
		fmt.Printf("no dir![%v]\n", path)
		// 创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
			return false, err
		} else {
			fmt.Printf("mkdir success!\n")
			return true, nil
		}
	}
}
