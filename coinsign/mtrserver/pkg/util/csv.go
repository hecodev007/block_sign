package util

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

//根据币种 扩展AddrInfo信息
type AddrInfo struct {
	Address string `json:"address"`
	PrivKey string `json:"privKey"`
}

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
		// 如果读到文件的结尾，EOF的优先级比nil高！
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

//指定路径生成地址csv文件
//params: createPath 生成路径
//params: mchId 商户ID
//params: orderId 订单ID
//params: coinName 币种
//params: []AddrInfo 地址信息
//params: privkeySize 私钥长度,用于校验
//params: addrSize 地址长度,用于校验

//a 文件 密文文件
//b 文件 密钥文件
//c 文件 明文文件
//d 文件 地址文件
func CreateAddrCsv(createPath, mchId, orderId, coinName string, addrInfos []AddrInfo, addrSize, privkeySize int) (addrs []string, err error) {
	if createPath == "" || mchId == "" || orderId == "" || coinName == "" {
		return nil, errors.New("empty params")
	}

	if len(addrInfos) == 0 {
		return nil, errors.New("empty addrs")
	}

	filename := createPath + "/" + mchId + "/" + coinName + "_%s_usb_" + orderId + ".csv"
	fileAPath := fmt.Sprintf(filename, "a")
	fileBPath := fmt.Sprintf(filename, "b")
	fileCPath := fmt.Sprintf(filename, "c")
	fileDPath := fmt.Sprintf(filename, "d")

	//判断是否存在同名文件
	_, err = os.Stat(fileAPath)

	if err == nil {
		return nil, errors.New("已经存在相同订单号文件")
	} else {
		//创建多层级目录
		_, err = CreateDirAll(createPath + "/" + mchId)
		if err != nil {
			log.Println("create create dir error ", err)
			return nil, err
		}
	}

	fileA, err := os.Create(fileAPath)
	if err != nil {
		return nil, err
	}
	defer fileA.Close()

	//B文件
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

	for _, info := range addrInfos {
		aesKey := RandBase64Key()
		ciphertext, err := AesBase64Crypt([]byte(info.PrivKey), aesKey, true)
		if err != nil {
			err = fmt.Errorf("AesBase64Crypt error:%s ", err)
			//不使用return,break之后直接手动释放写入流
			break
		}
		wa.Write([]string{info.Address, string(ciphertext)})
		wb.Write([]string{info.Address, string(aesKey)})
		wc.Write([]string{info.Address, string(info.PrivKey)})
		wd.Write([]string{info.Address})
		addrs = append(addrs, info.Address)
	}
	wa.Flush()
	wb.Flush()
	wc.Flush()
	wd.Flush()

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			log.Println(fmt.Sprintf("发生异常,删除文件,A文件:%s,B文件:%s,C文件:%s,D文件:%s",
				fileAPath, fileBPath, fileCPath, fileDPath))
			//删除文件
			os.Remove(fileAPath)
			os.Remove(fileBPath)
			os.Remove(fileCPath)
			os.Remove(fileDPath)
		}
	}()
	err = CheckPrivekeyFile(fileCPath, privkeySize)
	if err != nil {
		return nil, err
	}
	//理论上校验CD文件即可
	err = CheckAddressFile(fileCPath, addrSize)
	if err != nil {
		return nil, err
	}
	err = CheckAddressFile(fileDPath, addrSize)
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

//校验文件地址长度。避免出现写入不完整
func CheckAddressFile(filePath string, addrSize int) error {
	if filePath == "" {
		return errors.New("empty path")
	}

	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		lineArr := strings.Split(string(line), ",")
		if len(lineArr) < 1 {
			return fmt.Errorf("文件长度不对:%s", string(line))
		}
		if len(lineArr[0]) != addrSize {
			return fmt.Errorf("address :%s,地址长度异常，期望长度：%d,实际长度：%d",
				lineArr[0], addrSize, len(lineArr[0]))
		}
	}
	return nil
}

//校验c文件密钥长度。避免出现写入不完整
func CheckPrivekeyFile(filePath string, privKeySize int) error {
	if filePath == "" {
		return errors.New("empty path")
	}

	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		lineArr := strings.Split(string(line), ",")
		if len(lineArr) < 2 {
			return fmt.Errorf("文件长度不对:%s", string(line))
		}
		if len(lineArr[1]) != privKeySize {
			return fmt.Errorf("address :%s,私钥长度异常，期望长度：%d,实际长度：%d",
				lineArr[0], privKeySize, len(lineArr[1]))
		}
	}
	return nil
}
