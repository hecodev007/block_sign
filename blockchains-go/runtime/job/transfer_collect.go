package job

import (
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"xorm.io/builder"
)

// Job Specific Functions
type CollectApplyJob struct {
}

// 轮询数据库数据 进行交易
func (c CollectApplyJob) Run() {

	//查询数据,每次查询10条
	applies, err := entity.FcTransfersApply{}.Find(builder.Eq{}, 10)
	if err != nil {
		log.Errorf("查询订单数据异常:%s", err.Error())
		return
	}
	if len(applies) > 0 {
		log.Info("=======执行创建回调=======")
	}

	for _, apply := range applies {
		isToken := apply.Eostoken != "" || apply.Eoskey != ""
		decimal := 18
		coinName := apply.CoinName
		if isToken {
			//查找币种精度
			coin := &entity.FcCoinSet{}
			has, err := coin.Get(builder.Eq{"name": apply.Eoskey, "token": apply.Eostoken})
			if err != nil || !has {
				continue
			}
			decimal = coin.Decimal
		}
		//获取该apply的地址信息
		addresses, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": apply.Id})
		if err != nil {
			continue
		}
		fromAddrs := make([]string, 0)
		toAddrs := make([]string, 0)
		for _, addr := range addresses {
			if addr.AddressFlag == "from" {
				fromAddrs = append(fromAddrs, addr.Address)
			} else if addr.AddressFlag == "to" {
				toAddrs = append(toAddrs, addr.Address)
			}
		}
		if len(fromAddrs) == 0 || len(toAddrs) == 0 {
			continue
		}
		log.Infof("deciaml %d , coin name %s", decimal, coinName)
	}
}
