package main

/*

import (
	"flag"
	"fmt"
	"github.com/shopspring/decimal"
	"dataserver/conf"
	"dataserver/utils"
	"dataserver/utils/chain"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

func main() {
	var (
		routineNum int
		cfgFile    string
		sconf      conf.FixConfig
	)

	debug := true
	defer func() {
		if err := recover(); err != nil {
			log.Printf("process exit err : %v \n", err)
		}
	}()

	flag.IntVar(&routineNum, "n", 10, "each cpu's routine num")
	flag.StringVar(&cfgFile, "c", "", "set the yaml conf file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	err := conf.LoadConfig(cfgFile, &sconf)
	if err != nil {

	}

	if debug {
		log.Printf("LoadConfig : %v", sconf)
	}

	beginTime := time.Now()
	dirAbsPath, err := filepath.Abs(sconf.Csv.Dir)
	if err != nil {
		log.Printf("don't find csv dir %s ", dirAbsPath)
		panic(fmt.Errorf("don't find csv dir %s ", dirAbsPath))
	}

	keys, err := utils.ReadCsvFile(fmt.Sprintf("%s/%s.csv", dirAbsPath, sconf.Csv.ReadFile))
	if err != nil {
		log.Printf("ReadCsvFile err : %v", err)
		panic(fmt.Errorf("ReadCsvFile err : %v", err))
	}

	var writekeys [][]string
	client := chain.NewRpcClient(sconf.Nodes[sconf.Name].Url)
	{
		for _, key := range keys {

			tmp, err := client.GetBalance(key[0], chain.BlockLatest)
			if err != nil {
				continue
			}
			log.Printf("%s get balance %v", key[0], tmp)
			ethbalance := decimal.NewFromBigInt(tmp, -18)
			key = append(key, ethbalance.String())

			balanceOfData, _ := chain.ERC20{}.GetBalanceOf(key[0])
			tmp, err = client.GetBalanceToken("0xdac17f958d2ee523a2206206994597c13d831ec7", balanceOfData)
			if err != nil {
				continue
			}
			log.Printf("%s get balance token %v", key[0], tmp)
			usdtbalance := decimal.NewFromBigInt(tmp, -6) // .Shift(-6)
			key = append(key, usdtbalance.String())

			log.Printf("%s get chain : %v, usdt : %v", key[0], ethbalance, usdtbalance)
			// time.Sleep(time.Millisecond*20)
			writekeys = append(writekeys, key)
		}
	}

	err = utils.WriteCsvFile(writekeys, fmt.Sprintf("%s/%s.csv", dirAbsPath, sconf.Csv.WriteFile))
	if err != nil {
		log.Printf("WriteCsvFile err : %v", err)
		panic(fmt.Errorf("WriteCsvFile err : %v", err))
	}

	endTime := time.Since(beginTime)
	log.Printf(" %d keys,used time : %f s", len(writekeys), endTime.Seconds())
}
*/
