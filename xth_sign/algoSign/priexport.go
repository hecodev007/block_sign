package main

import (
	"algoSign/common/keystore"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/algorand/go-algorand-sdk/types"

	"log"

	"github.com/algorand/go-algorand-sdk/client/kmd"
	"golang.org/x/crypto/ed25519"
)

//go run priexport.go http://algo.rylink.io:27833 2dc5bfd8ed194b0ab271f4bcee568cacffd1adb1beadd150e594aa3e35e3729a test 123456
func init() {
	log.SetFlags(log.Llongfile)
}
func main() {
	if len(os.Args) < 4 {
		fmt.Println("////////////////////////")
		fmt.Println("启动参数: ./priexport $url $token $walletPassword")
		fmt.Println("eg:       ./priexport http://algo.rylink.io:27833 2dc5bfd8ed194b0ab271f4bcee568cacffd1adb1beadd150e594aa3e35e3729a 123456")
		fmt.Println("////////////////////////")
		return
	}
	url := os.Args[1] //"http://algo.rylink.io:27833"
	apitoken := os.Args[2]
	walletPassword := os.Args[3]
	client, err := kmd.MakeClient(url, apitoken)
	if err != nil {
		panic(err.Error())
	}
	/*
		if v, err := client.Version(); err != nil {
			panic(err.Error())
		} else {
			log.Println(String(v))
		}
	*/

	walletsresp, err := client.ListWallets()
	if err != nil {
		panic(err.Error())
	}

	for _, wallet := range walletsresp.Wallets {
		var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey
		iniwalletresp, err := client.InitWalletHandle(wallet.ID, walletPassword)
		if err != nil {
			panic(err.Error())
		}
		if iniwalletresp.Error {
			panic(iniwalletresp.Message)
		}
		walletHandle := iniwalletresp.WalletHandleToken
		lstKeyresp, err := client.ListKeys(walletHandle)
		if err != nil {
			panic(err.Error())
		} else if lstKeyresp.Error {
			panic(lstKeyresp.Message)
		}
		i := 0
		for _, addr := range lstKeyresp.Addresses {
			log.Println(wallet.Name, i, len(lstKeyresp.Addresses))
			if i%1000 == 0 {
				iniwalletresp, err := client.InitWalletHandle(wallet.ID, walletPassword)
				if err != nil {
					panic(err.Error())
				}
				if iniwalletresp.Error {
					panic(iniwalletresp.Message)
				}
				walletHandle = iniwalletresp.WalletHandleToken
			}

			exresp, err := client.ExportKey(walletHandle, walletPassword, addr)
			if err != nil {
				panic(err.Error())
			}
			pri := hex.EncodeToString(exresp.PrivateKey[:])
			public := exresp.PrivateKey.Public()
			addr := types.Address{}
			copy(addr[:], public.(ed25519.PublicKey)[:])
			pub := addr.String()
			aesKey := keystore.RandBase64Key()
			aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(pri), aesKey, true)
			if err != nil {
				panic(err.Error())
			}
			cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: pub, Key: string(aesPrivKey)})
			cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: pub, Key: string(aesKey)})
			cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: pub, Key: pri})
			cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: pub, Key: ""})
		}
		keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, wallet.Name, "20201222001")
	}

	println("导出成功!\n拷贝./csv文件")
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
