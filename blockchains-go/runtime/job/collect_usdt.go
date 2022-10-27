package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"strings"

	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectUsdtJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectUsdtJob(cfg conf.Collect2) cron.Job {
	return CollectUsdtJob{
		coinName: "usdt",
		cfg:      cfg,
	}
}

func (c CollectUsdtJob) Run() {
	var (
		mchs []*entity.FcMch
		err  error
	)
	stUsdtt := time.Now()

	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(stUsdtt).Seconds())
	if len(c.cfg.Mchs) != 0 {
		mchs, err = entity.FcMch{}.Find(builder.In("platform", c.cfg.Mchs).And(builder.Eq{"status": 2}))
	} else {
		mchs, err = entity.FcMch{}.Find(builder.In("id", builder.Select("mch_id").From("fc_mch_service").
			Where(builder.Eq{
				"status":    0,
				"coin_name": c.coinName,
			})).And(builder.Eq{"status": 2}))
	}
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
func (c *CollectUsdtJob) collect(mchId int, mchName string) error {
	//1. 先构建归集的交易订单
	orderReqs, err := c.walletServerCollect(mchId)
	if err != nil {
		return err
	}
	if len(orderReqs) > 0 {
		for _, orderReq := range orderReqs {

			if orderReq.ToAddress == "" {
				return errors.New("to address is null")
			}

			if mchId == 1 {
				if orderReq.ToAddress != "1PXC1aiZGXy4j81wt16g2rrUAsYVGUbuA7" {
					return errors.New("to address is error")
				}
			}

			//2. 判断to地址是不是这个商户下的冷地址
			result, err := dao.FcGenerateAddressGetByAddressAndMchId(orderReq.ToAddress, mchId)
			if err != nil || result == nil {
				return fmt.Errorf("find mchId (%d) address error: %v", mchId, err)
			}
			if result.Type != 1 {
				return fmt.Errorf("this mchId(%d) do not find this address(%s)", mchId, orderReq.ToAddress)
			}
			//3. 保存到apply表，生成applyId
			//生产归集订单
			cltApply := &entity.FcTransfersApply{
				Username:   "Robot",
				Department: "blockchains-go",
				Applicant:  mchName,
				OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
				OrderId:    util.GetUUID(),
				Operator:   "Robot",
				CoinName:   c.coinName,
				Type:       "gj",
				Purpose:    fmt.Sprintf("%s自动归集", c.coinName),
				Lastmodify: util.GetChinaTimeNow(),
				AppId:      mchId,
				Source:     1,
				Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
				Createtime: time.Now().Unix(),
			}
			applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
			applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
				Address:     orderReq.ToAddress,
				AddressFlag: "to",
				Status:      0,
				Lastmodify:  cltApply.Lastmodify,
			})
			applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
				Address:     orderReq.FromAddress,
				AddressFlag: "from",
				Status:      0,
				Lastmodify:  cltApply.Lastmodify,
			})
			appId, err := cltApply.TransactionAdd(applyAddresses)
			if err != nil {
				return fmt.Errorf("create app id  error :%v", err)
			}
			orderReq.ApplyId = appId
			orderReq.OuterOrderNo = cltApply.OutOrderid
			orderReq.OrderNo = cltApply.OrderId
			orderReq.MchId = int64(mchId)
			orderReq.MchName = mchName
			orderReq.CoinName = strings.ToUpper(c.coinName)
			//4. 将交易提交到链上
			_, err = c.walletServerCreate(orderReq)
			if err != nil {
				return fmt.Errorf("%s collect error: %v", c.coinName, err)
			}
			log.Infof("%s collect success,from(%s),to(%s),amount(%d)",
				mchName, orderReq.FromAddress, orderReq.ToAddress, orderReq.Amount)
			time.Sleep(time.Second * 10)
		}
		return nil
	} else {
		return fmt.Errorf("mch:%s empty", mchName)
	}
}

func (c *CollectUsdtJob) walletServerCreate(orderReq *transfer.UsdtOrderCollectRequest) (string, error) {
	//dd, _ := json.Marshal(orderReq)
	//log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/usdt/create", c.cfg.Url), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%d],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Amount, err)
	}
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	thr, err := transfer.DecodeCreateUsdtResp(data)
	if err != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,error: %v", orderReq.OuterOrderNo, err)
	}
	if thr.Code != 0 || thr.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Hash, nil
}

func (c *CollectUsdtJob) walletServerCollect(mchId int) ([]*transfer.UsdtOrderCollectRequest, error) {

	params := make(map[string]interface{})
	params["appId"] = mchId
	dd, _ := json.Marshal(params)
	log.Infof("%s Collect send :%s", c.coinName, string(dd))
	data, err := util.PostJsonByAuthAndTime(fmt.Sprintf("%s/usdt/collecttpl", c.cfg.Url), c.cfg.User, c.cfg.Password, params, 360)
	if err != nil {
		return nil, fmt.Errorf("rpc get %s collect data error: %v", c.coinName, err)
	}
	log.Infof("%s Collect resp :%s", c.coinName, string(data))
	cd, err := transfer.DecodeCreateTxResp(data)
	if err != nil {
		return nil, fmt.Errorf("decode %s collect data error: %v", c.coinName, err)
	}
	if cd.Code != 0 || cd.Data == nil {
		return nil, fmt.Errorf("rpc resp error: code=%d or data=null ,MchId=%d", cd.Code, mchId)
	}
	od, _ := json.Marshal(cd.Data)
	uros := make([]*transfer.UsdtOrderCollectRequest, 0)
	err = json.Unmarshal(od, &uros)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal usdt order request error: %v", err)
	}
	return uros, nil
}
