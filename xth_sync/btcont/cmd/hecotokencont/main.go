package main

import (
	"btcont/common/conf"
	"btcont/common/log"
	"btcont/common/model"
	"os"
	"strings"
	"time"

	"github.com/onethefour/common/xutils"
)

var COINNAME = "eth"

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "diff_suit_cold_scan" {
			diff_suit_cold_scan()
		} else if os.Args[1] == "diff_cold_scan" {
			diff_cold_scan()
		} else if os.Args[1] == "diff_cold_scan_asset" {
			diff_cold_scan_asset()
		} else {
			log.Info("方法不存在")
		}
	} else {
		log.Info("diff_suit_cold_scan", "修复额度不一致")
		log.Info("diff_cold_scan", "额度与链上额度不同的所有地址")
		log.Info("diff_cold_scan_asset ", "额度与链上额度不同的所有地址(amount>0)")

	}
	log.Info("btcount end!!!")
}

func diff_cold_scan() {

	node := model.NewEthNode(conf.Cfg.Nodes["eth"])
	Coinsets := new(model.FcCoinSet).AllToken(COINNAME)
	if len(Coinsets) == 0 {
		log.Info("没有代币")
		return
	}
	addressAmounts := new(model.FcAddressAmount).AllByCoinSets(Coinsets)
	log.Info(len(addressAmounts))
	now := time.Now()
	for _, addressAmount := range addressAmounts {
		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
		contract := Coinsets[strings.ToLower(addressAmount.CoinType)].Token
		if contract == "" {
			panic("")
		}
	reBalanceOf:
		//log.Info(addressAmount.Address)
		amount, t, err := node.BalanceOf(addressAmount.Address, contract)
		if err != nil {
			log.Info(addressAmount.Address, contract, err.Error())
			goto reBalanceOf
		}
		amount = amount.Shift(0 - int32(Coinsets[strings.ToLower(addressAmount.CoinType)].Decimal))
		//log.Info(addressAmount.Address, "链,钱包", amount.String(), addressAmount.Amount.String())
		if amount.Cmp(addressAmount.Amount) != 0 {
			nowAddrMnt := new(model.FcAddressAmount)
			nowAddrMnt.Get(addressAmount.CoinType, addressAmount.Address)
			if nowAddrMnt.Amount.Cmp(amount) != 0 {
				//log.Info(t)
				tm := time.Unix(t, 0)

				//log.Info(t, time.Now().Unix())
				//if t < time.Now().Unix()-60*60*3 {
				log.Info(addressAmount.Address, "链,钱包", now.Sub(tm).String(), contract, addressAmount.CoinType, amount.String(), addressAmount.Amount.String())
			}
		}
	}
}

func diff_cold_scan_asset() {
	node := model.NewEthNode(conf.Cfg.Nodes["eth"])
	Coinsets := new(model.FcCoinSet).AllToken(COINNAME)
	if len(Coinsets) == 0 {
		log.Info("没有代币")
		return
	}
	addressAmounts := new(model.FcAddressAmount).AllByCoinSetsAsset(Coinsets)
	log.Info(len(addressAmounts))
	now := time.Now()
	for _, addressAmount := range addressAmounts {
		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
		contract := Coinsets[strings.ToLower(addressAmount.CoinType)].Token
		if contract == "" {
			panic("")
		}
	reBalanceOf:
		//log.Info(addressAmount.Address)
		amount, t, err := node.BalanceOf(addressAmount.Address, contract)
		if err != nil {
			log.Info(err.Error())
			goto reBalanceOf
		}
		amount = amount.Shift(0 - int32(Coinsets[strings.ToLower(addressAmount.CoinType)].Decimal))
		//log.Info(addressAmount.Address, "链,钱包", amount.String(), addressAmount.Amount.String())
		if amount.Cmp(addressAmount.Amount) != 0 {
			//log.Info(t)
			tm := time.Unix(t, 0)

			//log.Info(t, time.Now().Unix())
			//if t < time.Now().Unix()-60*60*3 {
			log.Info(addressAmount.Address, "链,钱包", now.Sub(tm).String(), contract, addressAmount.CoinType, amount.String(), addressAmount.Amount.String())
			//}
		}
	}
}

func diff_suit_cold_scan() {
	node := model.NewEthNode(conf.Cfg.Nodes["eth"])
	Coinsets := new(model.FcCoinSet).AllToken(COINNAME)
	if len(Coinsets) == 0 {
		log.Info("没有代币")
		return
	}
	addressAmounts := new(model.FcAddressAmount).AllByCoinSets(Coinsets)
	log.Info(len(addressAmounts))
	now := time.Now()
	for _, addressAmount := range addressAmounts {
		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
		contract := Coinsets[strings.ToLower(addressAmount.CoinType)].Token
		if contract == "" {
			panic("")
		}
	retry:
		amount, t, err := node.BalanceOf(addressAmount.Address, contract)
		if err != nil {
			log.Info(err.Error())
			goto retry
		}
		amount = amount.Shift(0 - int32(Coinsets[strings.ToLower(addressAmount.CoinType)].Decimal))
		if amount.Cmp(addressAmount.Amount) != 0 {
			//log.Info(t)
			tm := time.Unix(t, 0)
			log.Info(addressAmount.Address, "链,钱包", now.Sub(tm).String(), contract, addressAmount.CoinType, amount.String(), addressAmount.Amount.String())
			//log.Info(t, time.Now().Unix())
			if t < time.Now().Unix()-60*60*3 {
				err = new(model.FcAddressAmount).SetMount(addressAmount.CoinType, addressAmount.Address, amount.String())
				if err != nil {
					log.Info(err.Error())
					log.Error(err.Error())
				}
			} else {
				log.Info("有比较新的交易不修复")
			}
		}
	}
}
