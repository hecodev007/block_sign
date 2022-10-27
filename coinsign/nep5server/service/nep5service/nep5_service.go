package nep5service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/group-coldwallet/nep5server/model/bo"
	"github.com/group-coldwallet/nep5server/model/global"
	"github.com/group-coldwallet/nep5server/model/vo"
	"github.com/group-coldwallet/nep5server/service"
	"github.com/group-coldwallet/nep5server/util"
	"github.com/group-coldwallet/nep5server/util/neputil"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

type Nep5Service struct {
}

func (n *Nep5Service) TokenSign(from, to, scriptAddr string, amount int64) (raw, txid string, err error) {
	fromPrivKey, _ := global.GetValue(from)
	if fromPrivKey == "" {
		err = fmt.Errorf(" address:%s, miss privkey", from)
		return "", "", err
	}
	return neputil.Nep5Transfer(from, to, fromPrivKey, scriptAddr, amount)
}

func (n *Nep5Service) CreateAddr(param *bo.CreateAddrBO) (*vo.CreateAddrVO, error) {
	createPath := param.CreatePath + "/" + param.MchId
	result := &vo.CreateAddrVO{
		Num:      param.Num,
		OrderId:  param.OrderId,
		MchId:    param.MchId,
		CoinName: param.CoinName,
	}
	//临时存储地址
	resultAddrs := make([]string, 0)
	addrs := make([]*bo.AddressInfo, 0)

	filename := createPath + "/" + param.CoinName + "_%s_usb_" + param.OrderId + ".csv"
	fileAPath := fmt.Sprintf(filename, "a")
	fileBPath := fmt.Sprintf(filename, "b")
	fileCPath := fmt.Sprintf(filename, "c")
	fileDPath := fmt.Sprintf(filename, "d")

	//A文件判断,
	_, err := os.Stat(fileAPath)
	if err == nil {
		logrus.Info("已经存在a文件 ", err)
		//读取A文件，直接读取地址
		resultAddrs, err := util.ReadCsv(fileAPath, 0)
		if err != nil {
			return nil, err
		}
		result.Addrs = resultAddrs
		return result, nil
	}

	//创建多层级目录
	_, err = util.CreateDirAll(createPath + "/")
	if err != nil {
		log.Println("create create dir error ", err)
		return nil, err
	}
	//生成地址
	for i := uint(0); i < param.Num; i++ {
		addr, privkey := neputil.CreateAddr()
		if addr == "" || privkey == "" {
			return nil, errors.New("create address error")
		}
		addrs = append(addrs, &bo.AddressInfo{
			Address:    addr,
			PrivateKey: privkey,
		})
	}
	if len(addrs) != int(param.Num) {
		logrus.Infof("createn address error,len :%d", len(addrs))
		return nil, fmt.Errorf("createn address error,len :%d", len(addrs))
	}

	fileA, err := os.Create(fileAPath)
	if err != nil {
		return nil, err
	}
	defer fileA.Close()

	fileB, err := os.Create(fileBPath)
	if err != nil {
		return nil, err
	}
	defer fileB.Close()

	//C文件
	fileC, err := os.Create(fileCPath)
	if err != nil {
		return nil, err
	}
	defer fileC.Close()

	//D文件
	_, err = os.Stat(fileDPath)
	fileD, err := os.Create(fileDPath)
	defer fileD.Close()

	wa := csv.NewWriter(fileA) //创建一个新的写入文件流
	wb := csv.NewWriter(fileB)
	wc := csv.NewWriter(fileC)
	wd := csv.NewWriter(fileD)
	for _, info := range addrs {
		aesKey := util.RandBase64Key()
		aesPrivateKey, err := util.AesBase64Crypt([]byte(info.PrivateKey), aesKey, true)
		if err != nil {
			log.Println("util.AesBase64Crypt error: ", err)
			continue
		}
		wa.Write([]string{info.Address, string(aesPrivateKey)})
		wb.Write([]string{info.Address, string(aesKey)})
		wc.Write([]string{info.Address, string(info.PrivateKey)})
		wd.Write([]string{info.Address})
		resultAddrs = append(resultAddrs, info.Address)
	}
	wa.Flush()
	wb.Flush()
	wc.Flush()
	wd.Flush()
	logrus.Infof(fileAPath, fileBPath, fileCPath, fileDPath)
	result.Addrs = resultAddrs
	return result, nil

}

func NewNep5Service() service.TransfService {
	return &Nep5Service{}
}
