package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"veservice/common"

	"github.com/astaxie/beego"
	"github.com/ethereum/go-ethereum/crypto"
)

type AddressInfo struct {
	address    string
	encryptWif string
	wifkey     string
	privateKey string
}

func GenAddress(number int, coin string) bool {
	path := fmt.Sprint("./", coin, "_a_usb_", time.Now().Unix(), ".csv")  // 加密私钥，地址
	path2 := fmt.Sprint("./", coin, "_b_usb_", time.Now().Unix(), ".csv") // 私钥秘钥，地址
	path3 := fmt.Sprint("./", coin, "_c_usb_", time.Now().Unix(), ".csv") // 地址
	path4 := fmt.Sprint("./", coin, "_d_usb_", time.Now().Unix(), ".csv") // 地址
	clientsFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return false
	}
	clientsFile2, err2 := os.OpenFile(path2, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err2 != nil {
		return false
	}
	clientsFile3, err3 := os.OpenFile(path3, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err3 != nil {
		return false
	}
	clientsFile4, err4 := os.OpenFile(path4, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err4 != nil {
		return false
	}
	clients := csv.NewWriter(clientsFile)
	clients2 := csv.NewWriter(clientsFile2)
	clients3 := csv.NewWriter(clientsFile3)
	clients4 := csv.NewWriter(clientsFile4)

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

				b := make([]byte, 32)
				_, err := rand.Read(b)

				if err != nil {
					beego.Debug(err)
					addressChan <- &AddressInfo{address: "", encryptWif: "", wifkey: ""}
					return
				}
				priv, _ := crypto.ToECDSA(b)
				addr := crypto.PubkeyToAddress(priv.PublicKey)
				address = addr.Hex()
				privkey = hex.EncodeToString(b)

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
					address:    strings.ToLower(address),
					encryptWif: encryptWif,
					wifkey:     wifkey,
					privateKey: privkey,
				}
				addressChan <- addrchan
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
				clients.Write([]string{addrchan.encryptWif, addrchan.address})
				clients2.Write([]string{addrchan.wifkey, addrchan.address})
				clients3.Write([]string{addrchan.privateKey})
				clients4.Write([]string{addrchan.address})

				fmt.Println("generate ", offset, " ", addrchan.address)
				offset += 1
			}
		}
		if total >= number {
			break
		}
	}

	clients.Flush()
	clients2.Flush()
	clients3.Flush()
	clients4.Flush()

	clientsFile.Close()
	clientsFile2.Close()
	clientsFile3.Close()
	clientsFile4.Close()

	if offset >= number {
		return true
	} else {
		os.Remove(path)
		os.Remove(path2)
		os.Remove(path3)
		os.Remove(path4)
		return false
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
