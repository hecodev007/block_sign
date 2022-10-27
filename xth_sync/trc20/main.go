package main

import (
	"github.com/shopspring/decimal"
	//"fmt"
	golog "log"
	"lunasync/common/log"
	"lunasync/models/po/atom"

	//"lunasync/common/conf"
	//"lunasync/common/log"
	"lunasync/common/db"
)

func init() {
	golog.SetFlags(golog.Llongfile)
}
func main() {
	amount1 := make([]*atom.FcAddressAmount1, 0)
	db.SyncDB.DB.Find(&amount1)
	amount2 := make([]*atom.FcAddressAmount2, 0)
	db.SyncDB.DB.Find(&amount2)
	log.Info(len(amount1), len(amount2))
	mapampunt1 := make(map[string]decimal.Decimal, 0)
	mapampunt2 := make(map[string]decimal.Decimal, 0)
	for _, v := range amount1 {
		mapampunt1[v.Address] = v.Amount
	}

	for _, v := range amount2 {
		mapampunt2[v.Address] = v.Amount
	}

	for k, _ := range mapampunt1 {
		if _, ok := mapampunt2[k]; !ok {

			mapampunt2[k] = decimal.Zero
		}
	}

	for k, _ := range mapampunt2 {
		if _, ok := mapampunt1[k]; !ok {
			//log.Info(k)
			mapampunt1[k] = decimal.Zero
		}
	}
	log.Info(len(mapampunt1), len(mapampunt2))

	for k, v := range mapampunt1 {
		if mapampunt2[k].Cmp(v) == 0 {
			continue
		}

		txlist := make([]*atom.FcTxClearDetail, 0)
		//db.SyncConn.ShowSQL(true)
		db.SyncConn.Where("addr=?", k).Find(&txlist)
		adds := decimal.Zero
		for _, tx := range txlist {
			if tx.Dir == 1 {
				adds = adds.Add(tx.Amount)
			} else if tx.Dir == 2 {
				adds = adds.Sub(tx.Amount)
			}
		}
		added := mapampunt2[k].Sub(v)
		if added.Cmp(adds) != 0 {
			log.Info(k, len(txlist), added.String(), adds.String())
		}
	}

}
