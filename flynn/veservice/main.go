package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"veservice/common"
	"veservice/controllers"
	_ "veservice/routers"
	"veservice/service"

	"github.com/astaxie/beego"
)

func main() {

	if len(os.Args) > 1 && os.Args[1] == "gen" {
		var num int = 1
		var coin string = beego.AppConfig.String("coin")

		if len(os.Args) > 2 {
			num = common.StrToInt(os.Args[2])
		}
		if len(os.Args) > 3 {
			coin = os.Args[3]
		}

		// 校验地址
		if coin != beego.AppConfig.String("coin") {
			fmt.Println("coin config error")
			return
		}

		result := service.GenAddress(num, coin)
		if beego.AppConfig.DefaultBool("delwallet", false) {
			os.RemoveAll(beego.AppConfig.String("walletpath"))
			os.Remove(beego.AppConfig.String("walletpath"))
		}

		if result {
			fmt.Println("finish success")
		} else {
			fmt.Println("finish fail")
		}

		return
	}

	var folderpath string = beego.AppConfig.String("csvdir")
	if len(os.Args) > 1 {
		folderpath = os.Args[1]
	}
	beego.Debug("file: ", folderpath)
	// 读取csv目录
	files, _ := ioutil.ReadDir(folderpath)
	for _, file := range files {
		if !file.IsDir() {
			// 解析文件名
			list := strings.Split(file.Name(), "_")
			if len(list) < 3 || (list[1] != "a" && list[1] != "b") {
				continue
			}

			filepath := folderpath + "/" + file.Name()

			// 读取配置文件
			fmt.Println("load ", filepath)
			if list[1] == "a" {
				controllers.ReadUsbAConfig(filepath)
			} else if list[1] == "b" {
				controllers.ReadUsbBConfig(filepath)
			}
		}
	}

	beego.Run()
}
