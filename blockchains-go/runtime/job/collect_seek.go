package job

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectSeekJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectSeekJob(cfg conf.Collect2) cron.Job {
	return CollectSeekJob{
		coinName: "seek",
		cfg:      cfg,
	}
}

func (c CollectSeekJob) Run() {
	start := time.Now()
	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(start).Seconds())
	mchs, err := entity.FcMch{}.Find(builder.In("id", builder.Select("mch_id").From("fc_mch_service").
		Where(builder.Eq{
			"status":    0,
			"coin_name": c.coinName,
		})).And(builder.Eq{"status": 2}))
	if err != nil {
		log.Errorf("find platforms err %v", err)
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
func (c *CollectSeekJob) collect(mchId int, mchName string) error {
	//获取归集的目标冷地址
	toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   c.coinName,
	})
	if len(toAddrs) == 0 {
		return fmt.Errorf("%s don't hava cold address", mchName)
	}
	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount >= ? and forzen_amount = 0", c.cfg.MinAmount)).
		And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
	if err != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return fmt.Errorf("%s don't hava need collected address", mchName)
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
		Purpose:    "自动归集",
		Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		Source:     1,
	}
	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	for _, to := range toAddrs {
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     to,
			AddressFlag: "to",
			Status:      0,
		})
	}
	for _, from := range fromAddrs {
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     from,
			AddressFlag: "from",
			Status:      0,
		})
	}
	appId, err := cltApply.TransactionAdd(applyAddresses)
	if err != nil {
		//log.Errorf("apply create err : %v", err)
		return err
	}
	//开始请求钱包服务归集
	for i, from := range fromAddrs {
		//随机获取冷地址
		to := toAddrs[i%len(toAddrs)]
		orderReq := &transfer.OrderRequest2{}
		orderReq.ApplyId = appId
		orderReq.OuterOrderNo = cltApply.OutOrderid
		orderReq.OrderNo = fmt.Sprintf("%s_%d", cltApply.OrderId, i)
		orderReq.MchId = int64(mchId)
		orderReq.MchName = mchName
		orderReq.CoinName = c.coinName

		orderReq.OrderAddress = append(orderReq.OrderAddress, &transfer.OrderAddrRequest2{
			Dir:     transfer.DirTypeFrom,
			Address: from,
		})
		orderReq.OrderAddress = append(orderReq.OrderAddress, &transfer.OrderAddrRequest2{
			Dir:     transfer.DirTypeTo,
			Address: to,
			Amount:  decimal.New(-8, 0),
		})
		//直接发起交易
		err := c.walletServerCreate(orderReq)
		if err != nil {
			log.Errorf("%s 归集交易失败，%s", c.coinName, err.Error())
			//更新减少冻结金额
			continue
		}
		log.Infof("address：%s,归集交易成功", from)
	}
	return nil
}

//创建交易接口参数
func (c *CollectSeekJob) walletServerCreate(orderReq *transfer.OrderRequest2) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", c.coinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", c.coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}
