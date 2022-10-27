package job

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	//"custody-merchant-admin/middleware/verify"
	"custody-merchant-admin/module/blockChainsApi"
	"custody-merchant-admin/module/log"
)

//轮询更新币种列表
type WalletCoinListCallBackJob struct {
}

func (e WalletCoinListCallBackJob) Run() {
	var limit int
	var offset int
	var l int

	coins := make([]domain.BCCoinInfo, 0)
	subcoins := make([]domain.BCCoinInfo, 0)

	limit = 50
	l = 50
	//分页获取页面
	for limit <= l {
		var list []domain.BCCoinInfo
		var err error
		list, err = blockChainsApi.BlockChainGetCoinList(limit, offset, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
		if err != nil {
			log.Error("轮询更新币种列表 err = %v", err)
			return
		}
		l = len(list)
		if l <= 0 {
			log.Error("轮询更新币种列表 返回为空")
			return
		}
		offset = offset + limit
		if err != nil {
			log.Error("轮询更新币种列表 BlockChainGet err:", err.Error())
			l = 0
		}
		for _, item := range list {
			if item.Name == "" {
				continue
			}
			subcoins = append(subcoins, item)

			if item.Father == "" {
				coins = append(coins, item)
			}
			//else {
			//	subcoins = append(subcoins, item)
			//}
		}
	}

	//更新币数据
	err := service.UpdateCoinDB(coins)
	if err != nil {
		log.Error("轮询更新币种列表 更新币数据 err:", err.Error())
	}

	err = service.UpdateSubCoinDB(subcoins)
	if err != nil {
		log.Error("轮询更新币种列表 更新币数据 err:", err.Error())
	}

}
