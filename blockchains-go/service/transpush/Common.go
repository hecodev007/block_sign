package transpush

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/shopspring/decimal"
)

// 推送任务接口
type TransPushTask interface {
	Run(reqexit <-chan bool)
}

var WaitGroupTransPush sync.WaitGroup
var PushTaskList []interface{}
var exectasknum int = 0
var execexitchan chan bool

func RunAllTask() {
	PushTaskList = append(PushTaskList, new(ExecImport))
	PushTaskList = append(PushTaskList, new(ExecConfirm))
	PushTaskList = append(PushTaskList, new(ExecPustList))
	PushTaskList = append(PushTaskList, new(AccountModel))
	PushTaskList = append(PushTaskList, new(EosModel))
	PushTaskList = append(PushTaskList, new(UtxoModel))
	PushTaskList = append(PushTaskList, new(BtmModel))
	//PushTaskList = append(PushTaskList, new(Sync)) //统计余额任务
	//PushTaskList = append(PushTaskList, new(ExecAmount))

	exectasknum = len(PushTaskList)
	execexitchan = make(chan bool, exectasknum)
	WaitGroupTransPush.Add(exectasknum)

	for _, v := range PushTaskList {
		go v.(TransPushTask).Run(execexitchan)
	}
}

func ExitAllTask() {
	for i := 0; i < exectasknum; i++ {
		execexitchan <- true
	}
	WaitGroupTransPush.Wait()
}

func in_array(need interface{}, needArr []interface{}) bool {
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}

type bcadd_multi_func func(interface{}) string

func bcadd_multi(_func bcadd_multi_func, _list interface{}) string {
	total := decimal.NewFromInt(0)
	for _, v := range _list.([]interface{}) {
		stramount := _func(v)
		tmp, _ := decimal.NewFromString(stramount)
		total.Add(tmp)
	}
	return total.String()
}

func bcadd_multi_utxoinput(_func bcadd_multi_func, _list interface{}) string {
	total := decimal.NewFromInt(0)
	for _, v := range _list.([]model.PushTxInput) {
		stramount := _func(v)
		tmp, _ := decimal.NewFromString(stramount)
		total.Add(tmp)
	}
	return total.String()
}

func bcadd_multi_utxooutput(_func bcadd_multi_func, _list interface{}) string {
	total := decimal.NewFromInt(0)
	for _, v := range _list.([]model.PushTxOutput) {
		stramount := _func(v)
		tmp, _ := decimal.NewFromString(stramount)
		total.Add(tmp)
	}
	return total.String()
}

func bcadd_multi_account(_func bcadd_multi_func, _list interface{}) string {
	total := decimal.NewFromInt(0)
	for _, v := range _list.([]model.PushAccountTx) {
		stramount := _func(v)
		tmp, _ := decimal.NewFromString(stramount)
		total.Add(tmp)
	}
	return total.String()
}

func format_decimal(str []byte, flag int, app_id int) []byte {
	var vo map[string]interface{}
	json.Unmarshal(str, &vo)

	////1接收，2发送
	if vo["is_in"] != nil {
		vo["is_in"] = 1
	} else {
		vo["is_in"] = 2
	}
	vo["amount"] = fmt.Sprintf("%s", vo["amount"])
	vo["fee"] = fmt.Sprintf("%s", vo["fee"])
	if flag > 0 {
		vo["is_in"] = flag
	}
	if vo["is_in"] == 2 && vo["transaction_id"] != nil {
		log.Infof("transaction_id=====>:%s", vo["transaction_id"].(string))
		log.Infof("coin=====>:%s", vo["coin"].(string))
		if in_array(strings.ToLower(vo["coin"].(string)), []interface{}{"iotx", "matic-matic", "rub", "cocos", "wbc", "ont", "ong", "nas", "zvc", "stx", "mdu", "stg", "klay", "gxc", "kava", "luna", "lunc",
			"cds", "ar", "ksm", "bnc", "hnt", "crab", "vet", "uca", "celo", "fio", "mtr", "sol", "tlos", "pcx", "ghost", "dot", "azero", "sgb-sgb", "kar", "nas", "dcr", "doge", "avax", "bsc",
			"dash", "fil", "wd", "biw", "atom", "near", "yta", "cfx", "star", "fis", "oneo", "atp", "cph-cph", "xlm", "bcha", "xec", "ada", "trx", "zen", "mw", "dip", "algo", "ori", "bos", "okt", "glmr", "avaxcchain",
			"heco", "nyzo", "xdag", "iost", "hsc", "dhx", "dom", "wtc", "moac", "satcoin", "eac", "iota", "kai", "rbtc", "movr", "sep20", "brise-brise", "ccn", "optim", "ftm", "welups", "rose", "one", "rev", "tkm", "ron",
			"neo", "icp", "flow", "uenc", "btm", "cspr", "crust", "rei", "evmos", "aur", "dscc", "mob", "dscc1", "deso", "lat", "nodle", "hbar"}) {
			//order_hot := &entity.FcOrderHot{}
			//dao.TransPushGet("select outer_order_no from fc_order_hot where tx_id = ? and status = 4", order_hot, vo["transaction_id"].(string))
			order_hot, errOrder := dao.FcOrderHotGetByTxid(vo["transaction_id"].(string), int(status.BroadcastStatus))
			if errOrder != nil {
				log.Infof("外部订单,填充hot order订单失败：err=[%s]", errOrder.Error())
			}
			if order_hot != nil {
				vo["outer_order_no"] = order_hot.OuterOrderNo
				vo["outOrderId"] = order_hot.OuterOrderNo
			}

		} else {
			//order := &entity.FcOrder{}
			//dao.TransPushGet("select outer_order_no from fc_order where tx_id = ? and status = 4", order, vo["transaction_id"].(string))
			order, errOrder := dao.FcOrderGetByTxid(vo["transaction_id"].(string), int(status.BroadcastStatus))
			if errOrder != nil {
				log.Error("外部订单填充cold order订单失败，err=[%s]", errOrder.Error())
			}
			if order != nil {
				vo["outer_order_no"] = order.OuterOrderNo
				vo["outOrderId"] = order.OuterOrderNo
			}
		}
	}
	if app_id > 0 {
		vo["user_sub_id"] = app_id
	}
	data, _ := json.Marshal(vo)
	log.Info("vo:", vo)
	return data
}
