package main

import (
	"btcont/common/log"
	"btcont/common/model"
	"os"
	"time"

	"github.com/shopspring/decimal"

	"github.com/onethefour/common/xutils"
)

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "AmountBylow" {
			//log.Info(AmountBylow(os.Args[2]))
		} else if os.Args[1] == "diff_txs" {
			//diff_txs(os.Args[2])
		} else if os.Args[1] == "diff_cold_flow" {
			//diff_cold_flow()
		} else if os.Args[1] == "diff_suit_cold_scan" {
			diff_suit_cold_scan()
		} else if os.Args[1] == "diff_cold_scan" {
			diff_cold_scan()
		} else if os.Args[1] == "diff_cold_scan_asset" {
			diff_cold_scan_asset()
		} else {
			log.Info("方法不存在")
		}
	} else {
		log.Info("diff_suit_cold_scan", "addr 漏推送的交易id")
		log.Info("diff_cold_flow ", "额度与流水不一致的所有地址")
		log.Info("diff_cold_scan", "额度与链上额度不同的所有地址")
	}
	log.Info("btcount end!!!")
}

func diff_cold_scan() {
	addressAmounts, err := new(model.FcAddressAmount).All("USDT-TRC20")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(len(addressAmounts))
	now := time.Now()
	for _, addressAmount := range addressAmounts {
		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
	reBalanceOf:
		amount, t, err := new(model.TrxScan).BalanceOf(addressAmount.Address, "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t")
		if err != nil {
			log.Info(err.Error())
			goto reBalanceOf
		}
		if amount.Cmp(addressAmount.Amount) != 0 {
			//log.Info(t)
			tm := time.Unix(t, 0)

			//log.Info(t, time.Now().Unix())
			//if t < time.Now().Unix()-60*60*3 {
			log.Info(addressAmount.Address, "链,钱包", now.Sub(tm).String(), amount.String(), addressAmount.Amount.String())
			//}
		}
	}
}

func diff_cold_scan_asset() {
	addressAmounts, err := new(model.FcAddressAmount).AllAssert("USDT-TRC20")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(len(addressAmounts))
	now := time.Now()
	for _, addressAmount := range addressAmounts {
		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
	reBalanceOf:
		amount, t, err := new(model.TrxScan).BalanceOf(addressAmount.Address, "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t")
		if err != nil {
			log.Fatal(err.Error())
			goto reBalanceOf
		}
		if amount.Cmp(addressAmount.Amount) != 0 {
			//log.Info(t)
			tm := time.Unix(t, 0)

			//log.Info(t, time.Now().Unix())
			if t < time.Now().Unix()-60*60*3 {
				log.Info(addressAmount.Address, "链,钱包", now.Sub(tm).String(), amount.String(), addressAmount.Amount.String())
			}
		}
	}
}

func diff_suit_cold_scan() {
	addressAmounts, err := new(model.FcAddressAmount).All("usdt-trc20")
	if err != nil {
		log.Fatal(err.Error())
	}
	now := time.Now()
	for _, addressAmount := range addressAmounts {
		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
		amount, t, err := new(model.TrxScan).BalanceOf(addressAmount.Address, "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t")
		if err != nil {
			log.Fatal(err.Error())
		}
		if amount.Cmp(addressAmount.Amount) != 0 {
			//log.Info(t)
			tm := time.Unix(t, 0)
			log.Info(addressAmount.Address, "链,钱包", now.Sub(tm).String(), amount.String(), addressAmount.Amount.String())
			//log.Info(t, time.Now().Unix())
			if t < time.Now().Unix()-60*60*3 {
				err = new(model.FcAddressAmount).SetMount("usdt-trc20", addressAmount.Address, amount.String())
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
func diff_cold_flow() {
	addressAmounts, err := new(model.FcAddressAmount).All("usdt-trc20")
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, addressAmount := range addressAmounts {
		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
		flowAmount, err := AmountBylow(addressAmount.Address)
		if err != nil {
			log.Fatal(err.Error())
		}
		if flowAmount.Cmp(addressAmount.Amount) != 0 {
			scAmount, _, err := new(model.Scan).BalanceOf(addressAmount.Address)
			if err != nil {
				log.Error(err.Error())
			}
			log.Info(addressAmount.Address, "流水校验不一致", scAmount.String(), flowAmount.String(), addressAmount.Amount.String())
			if scAmount.Cmp(flowAmount) != 0 {
				diff_txs(addressAmount.Address)
				log.Info("")
			}
		}
	}
}

func diff_txs(addr string) {
	_, txs, err := new(model.Scan).AllTxsByAddr(addr)
	if err != nil {
		log.Fatal(err.Error())
	}

	txsMap := new(model.Scan).ToMap(txs)
	fctxs, err := new(model.FcTxClearDetail).AllTxIdByAddr("usdt-trc20", addr)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(len(txs), len(txsMap))
	for _, fctx := range fctxs {
		if _, has := txsMap[fctx.TxId]; !has {
			log.Info(addr, fctx.TxId, "浏览器没有")
			continue
		}
		delete(txsMap, fctx.TxId)
	}
	if len(txsMap) != 0 {
		for _, v := range txsMap {
			if v.Height > 674386 {
				log.Info(addr, v.Txid, "钱包没有没有")
			}
		}
	}
}
func AmountBylow(addr string) (decimal.Decimal, error) {
	utxos, err := new(model.FcTxClearDetail).AllByAddr("usdt-trc20", addr)
	if err != nil {
		return decimal.Decimal{}, err
	}
	ret := decimal.Zero
	for _, v := range utxos {
		if v.Dir == 1 {
			ret = ret.Add(v.Amount)
		} else {
			ret = ret.Sub(v.Amount)
		}
	}
	return ret, nil
}
func test() {
	addressAmounts, err := new(model.FcAddressAmount).All("usdt-trc20")
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, addressAmount := range addressAmounts {

		if addressAmount.Address == "" {
			log.Info(xutils.String(addressAmount))
			continue
		}
		amount, txs, err := new(model.Scan).AllTxsByAddr(addressAmount.Address)
		if err != nil {
			log.Fatal(err.Error())
		}
		if amount.Cmp(addressAmount.Amount) != 0 {
			log.Info(addressAmount.Address, "链上额度不一致", amount.String(), addressAmount.Amount.String())
		}
		txsMap := new(model.Scan).ToMap(txs)
		fctxs, err := new(model.FcTxClearDetail).AllTxIdByAddr("usdt-trc20", addressAmount.Address)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, fctx := range fctxs {
			if _, has := txsMap[fctx.TxId]; !has {
				log.Info(addressAmount.Address, fctx.TxId, "浏览器没有")
				continue
			}
			delete(txsMap, fctx.TxId)
		}
		if len(txsMap) != 0 {
			for _, v := range txsMap {
				log.Info(addressAmount.Address, v.Txid, "钱包没有没有")
			}
		}

	}
}
