package sync_sys

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func ClearLog() {
	xkutils.NewJobsSchedule(func() {
		ph, err := os.Getwd()
		if err != nil {
			return
		}
		remove(ph + "/" + Conf.LogFile)
	}, time.Hour*24)
}

func remove(file string) error {
	//获取文件或目录相关信息
	fileInfoList, err := ioutil.ReadDir(file)
	if err != nil {
		log.Fatal(err)
	}
	for i := range fileInfoList {
		ph := file + "/" + fileInfoList[i].Name()
		now := time.Now().Local().Unix()
		tm := strings.Replace(strings.Replace(fileInfoList[i].Name(), "log-", "", 1), ".log", "", 1)
		location, err := time.ParseInLocation("2006-01-02", tm, time.Local)
		if err != nil {
			return err
		}
		a := now - location.Local().Unix()
		if a > int64(60*60*24*30) {
			err = os.Remove(ph)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
