package job

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"math/rand"
	"strconv"
	"xorm.io/builder"
)

//查找符合归集条件的币种，生成归集订单，如果该地址没有合适的手续费，那么生成打手续费订单

// Job Specific Functions
type CollectErc20Job struct {
	cfg conf.Collect2
}

//
func (s CollectErc20Job) Run() {
	//获取币种信息
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"pid": 5, "status": 1})
	if err != nil {
		log.Errorf("%v", err)
		return
	}
	//todo 开启每个币种的归集策略
	for _, coin := range coins {
		for _, ignore := range s.cfg.IgnoreCoins {
			if coin.Name == ignore {
				continue
			}
		}
		//查找相关商户
		mchs, err := entity.FcMch{}.Find(builder.Eq{"status": 2}.And(builder.In("platform", s.cfg.Mchs)))
		if err != nil {
			log.Errorf("find platforms err %v", err)
			return
		}
		//轮训每个商户需要归集的订单
		addrList := entity.FcGenerateAddressList{}
		addrAmount := entity.FcAddressAmount{}
		for _, mch := range mchs {
			if toAddresses, err := addrList.FindAddress(builder.Eq{
				"type":        1,
				"status":      2,
				"platform_id": mch.Id,
				"coin_name":   "eth",
			}.And(builder.Neq{"address": "0x9760862d09b70433a91ff27cbd069f51ef1cbd5c"})); err == nil {
				toAddress := toAddresses[rand.Intn(len(toAddresses))]
				//查找有需要进行归集的源地址
				if fromAddresses, err := addrAmount.FindAddress(builder.Eq{
					"status":      2,
					"platform_id": mch.Id,
					"coin_name":   coin.Name,
				}.And(builder.Gte{
					"amount": s.cfg.MinAmount,
				}.And(builder.NotIn("address", toAddresses)).
					And(builder.In("type", 2, 6))), 100); err == nil {
					collectAddrs := make([]string, 0)
					feeAddrs := make([]string, 0)
					//过滤出来需要打手续费的地址
					for _, fromAddr := range fromAddresses {
						if needTransferFee(fromAddr, mch.Id, s.cfg.MinAmount) {
							feeAddrs = append(feeAddrs, fromAddr)
						} else {
							collectAddrs = append(collectAddrs, fromAddr)
						}
					}
					if len(collectAddrs) > 0 {
						//生成归集订单
						cltApply := &entity.FcTransfersApply{
							Username:   "Robot",
							Department: "冷钱包组",
							Applicant:  mch.Platform,
							Operator:   "Robot",
							AppId:      mch.Id,
							Type:       "gj",
							Purpose:    "自动归集",
						}
						if coin.Type != 2 {
							cltApply.Eoskey = coin.Name
							cltApply.Eostoken = coin.Token
						}
						applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
						applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
							Address:     toAddress,
							AddressFlag: "to",
							Status:      0,
						})
						for _, cltAddr := range collectAddrs {
							applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
								Address:     cltAddr,
								AddressFlag: "from",
								Status:      0,
							})
						}
						if _, err := cltApply.TransactionAdd(applyAddresses); err != nil {
							log.Errorf("apply create err : %v", err)
						}
					}
					if len(feeAddrs) > 0 {
						//生成手续费订单
						feeApply := &entity.FcTransfersApply{
							Username:   "Robot",
							Department: "冷钱包组",
							Applicant:  mch.Platform,
							Operator:   "Robot",
							AppId:      mch.Id,
							Type:       "fee",
							Purpose:    "自动归集",
						}
						//查找手续费地址
						feeAddress := &entity.FcAddressAmount{}
						has, err := feeAddress.Get(builder.In("address",
							builder.Select("address").From("fc_generate_address_list").
								Where(builder.Eq{
									"type":        1,
									"status":      2,
									"platform_id": mch.Id,
									"coin_name":   "eth",
								})).And(builder.Eq{
							"app_id":    mch.Id,
							"coin_type": "eth",
							"type":      1,
						}))
						if err != nil {
							continue
						}
						if !has {
							continue
						}
						//if feeAddress.Amount > s.cfg.AlarmFee {
						//	//todo 商户手续费告警
						//	continue
						//}
						applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
						applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
							Address:     feeAddress.Address,
							AddressFlag: "from",
							Status:      0,
						})
						for _, feeAddr := range feeAddrs {
							applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
								Address:     feeAddr,
								AddressFlag: "to",
								Status:      0,
							})
						}
						if _, err := feeApply.TransactionAdd(applyAddresses); err != nil {
							log.Errorf("apply create err : %v", err)
						}
					}
				}
			}
		}
	}
}

func needTransferFee(address string, mchId int, minAmount float64) bool {
	amt := &entity.FcAddressAmount{}
	has, err := amt.Get(builder.Eq{
		"coin_type": "eth",
		"app_id":    mchId,
		"address":   address,
	})
	if err != nil {
		return false
	}
	if !has {
		return true
	}
	amount, err := strconv.ParseFloat(amt.Amount, 64)
	if err != nil {
		return false
	}
	if amount < minAmount {
		return true
	}
	return false
}
