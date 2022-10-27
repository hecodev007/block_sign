package files

import (
	"bytes"
	. "custody-merchant-admin/config"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"
)

func SingleFile(file *multipart.FileHeader, url, host string) (string, error) {
	static, err := createDir(url)
	if err != nil {
		return "", err
	}
	//打开用户上传的文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	// 创建目标文件，就是我们打算把用户上传的文件保存到什么地方
	// files.Filename 参数指的是我们以用户上传的文件名，作为目标文件名，也就是服务端保存的文件名跟用户上传的文件名一样
	path := getUrl(static, file.Filename)
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	// 这里将用户上传的文件复制到服务端的目标文件
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", host, path), nil
}

func MultipleFile(form *multipart.Form, url, host string) (string, error) {
	static, err := createDir(url)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	files := form.File["files"]
	i := 0
	for _, file := range files {
		if i != 0 {
			buf.WriteString(",")
		}
		i++
		// Source
		src, err := file.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()
		// Destination
		path := getUrl(static, file.Filename)
		dst, err := os.Create(path)
		if err != nil {
			return "", err
		}
		defer dst.Close()
		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return "", err
		}
		buf.WriteString(fmt.Sprintf("%s/%s", host, path))
	}
	return buf.String(), nil
}

func createDir(url string) (string, error) {
	static := Conf.StaticFile + url
	err := os.MkdirAll(static, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		log.Errorf("创建文件夹失败：%v", err.Error())
	}
	return static, nil
}

func getUrl(static, filename string) string {
	str := strings.Split(filename, ".")
	return static + "/" + xkutils.IntToString(int(time.Now().Local().Unix())) + "." + str[len(str)-1]
}
