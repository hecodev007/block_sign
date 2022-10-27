package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/robfig/cron/v3"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

/*
指定币种以及指定地址归集
*/
type CollectEthToAssignAddressJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectEthToAssignAddressJob(cfg conf.Collect2) cron.Job {
	return CollectEthToAssignAddressJob{
		coinName: "eth",
		cfg:      cfg,
	}
}

func (c CollectEthToAssignAddressJob) Run() {
	var (
		mchs []*entity.FcMch
		err  error
	)
	start := time.Now()
	mchs = make([]*entity.FcMch, 0)
	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(start).Seconds())
	if len(c.cfg.Mchs) != 0 {
		//mchs, err = entity.FcMch{}.Find(builder.In("platform", c.cfg.Mchs).And(builder.Eq{"status": 2}))
		err = db.Conn.Where("status = ?", 2).In("platform", c.cfg.Mchs).Find(&mchs)
	} else {
		mchs, err = entity.FcMch{}.Find(builder.In("id", builder.Select("mch_id").From("fc_mch_service").
			Where(builder.Eq{
				"status":    0,
				"coin_name": c.coinName,
			})).And(builder.Eq{"status": 2}))
	}
	if err != nil {
		log.Errorf("find platforms err==》 %s", err.Error())
		return
	}

	wg := &sync.WaitGroup{}
	for _, tmp := range mchs {
		go func(mch *entity.FcMch) {
			wg.Add(1)
			defer wg.Done()
			if err := c.collect(mch.Id, mch.Platform); err != nil {
				log.Errorf(" %s ## collect err: %v", mch.Platform, err)
			}
		}(tmp)
	}
	wg.Wait()
}

func (c *CollectEthToAssignAddressJob) collect(mchId int, mchName string) error {
	if len(c.cfg.AssignCoins) <= 0 {
		return errors.New("do not find any assign coin in config")
	}
	if len(c.cfg.AssignAddress) <= 0 {
		return errors.New("do not find any assign address in config")
	}
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": 5}.Or(builder.Eq{"id": 5})).And(builder.In("name", c.cfg.AssignCoins)))
	if err != nil {
		return fmt.Errorf("get coinset error,Err=[%v]", err)
	}
	for _, coinName := range c.cfg.AssignCoins {

		for _, coin := range coins {
			if strings.ToLower(coin.Name) == strings.ToLower(coinName) {
				//获取to地址
				//获取归集的目标冷地址
				toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
					"type":        address.AddressTypeCold,
					"status":      address.AddressStatusAlloc,
					"platform_id": mchId,
					"coin_name":   c.coinName,
				})
				if err != nil || len(toAddrs) == 0 {
					//fmt.Errorf("%s don't hava cold address", mchName)
					continue
				}
				//生成归集订单
				cltApply := &entity.FcTransfersApply{
					Username:   "Robot",
					CoinName:   c.coinName,
					Department: "blockchains-go",
					OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
					OrderId:    util.GetUUID(),
					Applicant:  mchName,
					Operator:   "Robot",
					AppId:      mchId,
					Type:       "gj",
					Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
					Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
					Createtime: time.Now().Unix(),
					Lastmodify: util.GetChinaTimeNow(),
					Source:     1,
				}
				if coin.Name != c.coinName { //代表是代币
					cltApply.Eostoken = coin.Token
					cltApply.Eoskey = coin.Name
				}

				applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
				to := toAddrs[0]
				//randNum := 4 //随机取前4个地址
				//if len(toAddrs) < randNum {
				//	randNum = len(toAddrs)
				//}
				//to := toAddrs[rand.Intn(randNum)] //随机地址
				////减少某个地址的归集次数，因为被浏览器标记了
				//if strings.ToLower(to) == "0x980a4732c8855ffc8112e6746bd62095b4c2228f" {
				//	//再次随机两次
				//	for r := 0; r < 2; r++ {
				//		to = toAddrs[rand.Intn(randNum)]
				//		if strings.ToLower(to) != "0x980a4732c8855ffc8112e6746bd62095b4c2228f" {
				//			break
				//		}
				//	}
				//}
				//for _, to := range toAddrs {
				applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
					Address:     to,
					AddressFlag: "to",
					Status:      0,
					Lastmodify:  cltApply.Lastmodify,
				})
				collectAddrs := make([]string, 0)
				for _, from := range c.cfg.AssignAddress {
					applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
						Address:     from,
						AddressFlag: "from",
						Status:      0,
						Lastmodify:  cltApply.Lastmodify,
					})
					collectAddrs = append(collectAddrs, from)
				}

				appId, err := cltApply.TransactionAdd(applyAddresses)
				if err == nil {
					//开始请求钱包服务归集
					orderReq := &transfer.EthCollectReq{}
					orderReq.ApplyId = appId
					orderReq.OuterOrderNo = cltApply.OutOrderid
					orderReq.OrderNo = cltApply.OrderId
					orderReq.MchId = int64(mchId)
					orderReq.MchName = mchName
					orderReq.CoinName = c.coinName
					orderReq.FromAddrs = collectAddrs
					orderReq.ToAddr = to
					if coin.Name != c.coinName { //如果是代币归集
						orderReq.ContractAddr = coin.Token
						orderReq.Decimal = coin.Decimal
					}
					if err := c.walletServerCollect(orderReq); err != nil {
						log.Errorf("err : %v", err)
					}
				}
			}
		}
	}
	return nil
}

//创建交易接口参数
func (c *CollectEthToAssignAddressJob) walletServerCollect(orderReq *transfer.EthCollectReq) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/collect", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", c.coinName, string(dd))
	log.Infof("%s Collect resp :%s", c.coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("walletServerCollect 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}
