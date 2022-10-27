package service

import (
	"bnbsign/common"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	//"io/ioutil"
	"os"
	"runtime"

	"github.com/binance-chain/go-sdk/keys"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type AddressInfo struct {
	address    string
	encryptWif string
	wifkey     string
	key        string
}

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

func GenAddress(number int) (bool, map[string]string) {
	address := make(map[string]string, 0)
	//businessdir := fmt.Sprint(dirpath, "/", business, "/")
	//if exists, _ := PathExists(businessdir); exists == false {
	//	err := os.Mkdir(businessdir, os.ModePerm)
	//	if err != nil {
	//		return false, nil
	//	}
	//}
	//path := fmt.Sprint(dirpath, "/", business, "/", coin, "_a_usb_", orderid, ".csv")  // 加密私钥，地址
	//path2 := fmt.Sprint(dirpath, "/", business, "/", coin, "_b_usb_", orderid, ".csv") // 私钥秘钥，地址
	//path3 := fmt.Sprint(dirpath, "/", business, "/", coin, "_c_usb_", orderid, ".csv") // 地址
	//clientsFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	//if err != nil {
	//	beego.Debug(err)
	//	return false, nil
	//}
	//clientsFile2, err2 := os.OpenFile(path2, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	//if err2 != nil {
	//	beego.Debug(err)
	//	return false, nil
	//}
	//clientsFile3, err3 := os.OpenFile(path3, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	//if err3 != nil {
	//	beego.Debug(err)
	//	return false, nil
	//}
	//
	//clients := csv.NewWriter(clientsFile)
	//clients2 := csv.NewWriter(clientsFile2)
	//clients3 := csv.NewWriter(clientsFile3)

	numcpu := runtime.NumCPU()
	buildnummap := []int{}
	addressChan := make(chan *AddressInfo, number)
	if number <= numcpu {
		numcpu = 1
		buildnummap = append(buildnummap, number)
	} else {
		// 计算每个chan生成多少个
		avg := number / numcpu
		for j := 0; j < numcpu; j++ {
			buildnummap = append(buildnummap, avg)
		}
		buildnummap[numcpu-1] += (number % numcpu)
	}
	//fmt.Println(buildnummap)

	for i := 0; i < numcpu; i++ {
		buildnum := buildnummap[i]
		go func() {
			//fmt.Println(buildnum)
			for index := 0; index < buildnum; index++ {
				var address string = ""
				var privkey string = ""

				privKey1 := secp256k1.GenPrivKey()
				privkey = hex.EncodeToString(privKey1[:])

				keyManager, err := keys.NewPrivateKeyManager(privkey)
				if err != nil {
					beego.Error(err)
				}
				address = keyManager.GetAddr().String()
				//tmp, _ := keyManager.ExportAsPrivateKey()
				//beego.Debug(tmp, address, privkey)

				if address == "" || privkey == "" {
					addressChan <- &AddressInfo{address: "", encryptWif: "", wifkey: ""}
					return
				}

				key := make([]byte, 32)
				rand.Read(key)
				wifkey := base64.StdEncoding.EncodeToString(key)
				wifkey = wifkey[0:32]
				encryptWif, _ := common.AesEncrypt(privkey, []byte(wifkey))

				addrchan := &AddressInfo{
					address:    address,
					encryptWif: encryptWif,
					wifkey:     wifkey,
					key:        privkey,
				}
				addressChan <- addrchan
				//clients.Write([]string{encryptWif, address})
				//clients2.Write([]string{wifkey, address})
				//clients3.Write([]string{address})

				//fmt.Println("generate ", index, " ", address)
			}
		}()
	}

	offset := 0
	total := 0
	for {
		select {
		case addrchan := <-addressChan:
			{
				total++
				if addrchan.address == "" || addrchan.encryptWif == "" || addrchan.wifkey == "" {
					break
				}

				address[addrchan.address] = addrchan.key

				fmt.Println("generate ", offset, " ", addrchan.address)
				offset += 1
			}
		}
		if total >= number {
			break
		}
	}

	if offset >= number {
		return true, address
	} else {
		return false, nil
	}
}

// 生成地址
func GetNewAddress() (string, error) {
	// 生成地址
	resp, err := common.Request("getnewaddress", nil)
	if err != nil {
		return "", err
	}

	var addrresp map[string]interface{}
	err = json.Unmarshal(resp, &addrresp)
	if err != nil {
		return "", err
	}

	if addrresp["error"] != nil || addrresp["result"] == nil {
		return "", err
	}

	address := addrresp["result"].(string)
	return address, err
}

// 获取私钥
func DmpPrivateKey(address string) (string, error) {
	// 获取私钥
	resp, err := common.Request("dumpprivkey", []interface{}{address})
	if err != nil {
		return "", err
	}

	var privresp map[string]interface{}
	err = json.Unmarshal(resp, &privresp)
	if err != nil {
		return "", err
	}

	if privresp["error"] != nil || privresp["result"] == nil {
		return "", err
	}

	privkey := privresp["result"].(string)
	return privkey, err
}
