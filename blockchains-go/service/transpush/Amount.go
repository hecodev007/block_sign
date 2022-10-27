package transpush

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"time"
)

// 同步金额模型
type ExecAmount struct {
	// nothing
}

func (r *ExecAmount) Run(reqexit <-chan bool) {
	log.Debug("Run ExecAmount")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("ExecAmount exit", s)
			run = false
			break

		default:
			dispossAmount()
			time.Sleep(time.Second * 10)
		}
	}
	WaitGroupTransPush.Done()
}

func dispossAmount() {
	mch_amount()
}

func mch_amount() {
	list := make([]*entity.FcAddressAmount, 0)
	err := dao.TransPushFind(&list, "select coin_type,app_id,sum(amount) as amount from fc_address_amount where app_id > 0 and type in (1, 2, 6, 5) group by coin_type, app_id")
	if err != nil {
		log.Error(err)
		return
	}

	if len(list) == 0 {
		return
	}

	for _, vo := range list {
		mchAmount := &entity.FcMchAmount{}
		isfind, err := dao.TransPushGet(mchAmount, "select id from fc_mch_amount where coin_type = ? and app_id = ?", vo.CoinType, vo.AppId)
		if err != nil {
			log.Error(err)
			break
		}

		if isfind {
			dao.TransPushUpdate("update fc_mch_amount set amount = ? where coin_type = ? and app_id = ?", vo.Amount, vo.CoinType, vo.AppId)
		} else {
			coin_id := 0
			if global.CoinDecimal[vo.CoinType] != nil {
				coin_id = global.CoinDecimal[vo.CoinType].Id
			}
			dao.TransPushInsert("insert into fc_mch_amount(coin_id, coin_type, amount, app_id) values(%d, %s, %s, %d)", coin_id, vo.CoinType, vo.Amount, vo.AppId)
		}
	}
}
