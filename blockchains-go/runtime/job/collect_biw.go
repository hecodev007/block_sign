package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math/rand"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectBiwJob struct {
	coinName string
	cfg      conf.Collect2
	limitMap sync.Map
}

func NewCollectBiwJob(cfg conf.Collect2) cron.Job {
	return CollectBiwJob{
		coinName: "biw",
		cfg:      cfg,
		limitMap: sync.Map{}, //初始化限制表
	}
}

func (c CollectBiwJob) Run() {
	var (
		mchs []*entity.FcMch
		err  error
	)
	start := time.Now()
	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(start).Seconds())

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

func (c *CollectBiwJob) collect(mchId int, mchName string) error {

	var (
		fromAmountInt64 decimal.Decimal                    //from金额
		toAmountInt64   decimal.Decimal                    //to金额
		txins           = make([]transfer.BchaTxInTpl, 0)  //utxo模板
		txouts          = make([]transfer.BchaTxOutTpl, 0) //utxo模板
		tpl             *transfer.BchaTxTpl                //模板
	)
	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount > ? and forzen_amount = 0", c.cfg.MinAmount)), c.cfg.MaxCount)
	if err != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil
		//return fmt.Errorf("%s don't have need collected from address", mchName)
	}
	//获取归集的目标冷地址
	toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   c.coinName,
	})
	if len(toAddrs) == 0 {
		return fmt.Errorf("%s don't have cold address", mchName)
	}
	if mchId == 1 && toAddrs[0] != "18gkMAPosZdXbihLNXUyh6qd1p75M98D5A" {
		return fmt.Errorf("%s err address", mchName)
	}

	to := toAddrs[rand.Intn(len(toAddrs))] //随机取个地址
	//fee := decimal.NewFromFloat(c.cfg.NeedFee)
	fee := decimal.NewFromFloat(0.01)

	byteData, err := util.PostJson(c.cfg.Url+"/api/v1/bcha/unspents", fromAddrs)
	unspents := new(transfer.BchaUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 || len(unspents.Data) == 0 {
		return errors.New("bcha empty unspents")
	}
	for i, v := range unspents.Data {
		if v.Confirmations == 0 {
			continue
		}
		if i == 30 {
			break
		}
		txin := transfer.BchaTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromIndex:  uint32(v.Vout),
			FromAmount: v.Amount,
		}
		fromAmountInt64 = fromAmountInt64.Add(decimal.New(v.Amount, 0))
		txins = append(txins, txin)
	}
	toAmountInt64 = fromAmountInt64.Sub(fee.Shift(8))
	txouts = append(txouts, transfer.BchaTxOutTpl{
		ToAddr:   to,
		ToAmount: toAmountInt64.IntPart(),
	})

	tpl = &transfer.BchaTxTpl{
		MchId:    mchName,
		OrderId:  util.GetUUID(),
		CoinName: c.coinName,
		TxIns:    txins,
		TxOuts:   txouts,
	}
	createData, _ := json.Marshal(tpl)

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
		Status:     int(entity.ApplyStatus_Merge), // 因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
	}
	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     to,
		AddressFlag: "to",
		Status:      0,
		Lastmodify:  cltApply.Lastmodify,
	})
	appId, err := cltApply.TransactionAdd(applyAddresses)
	if err != nil {
		return err
	}
	orderHot := &entity.FcOrderHot{
		ApplyId:      int(appId),
		OuterOrderNo: cltApply.OutOrderid,
		OrderNo:      cltApply.OrderId,
		MchName:      mchName,
		CoinName:     c.coinName,
		FromAddress:  "",
		ToAddress:    to,
		Amount:       toAmountInt64.IntPart(), //转换整型
		Quantity:     toAmountInt64.Shift(-8).String(),
		Decimal:      8,
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}
	txid, err := c.walletServerCreate(tpl)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		return fmt.Errorf("%s 归集失败，err:%s", c.coinName, err.Error())
	}
	orderHot.TxId = txid
	orderHot.Status = int(status.BroadcastStatus)
	// 保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		// 发送给钉钉
	}
	log.Infof("归集完成，txid:%s", txid)
	return nil
}

//创建交易接口参数

//创建交易接口参数
func (srv *CollectBiwJob) walletServerCreate(orderReq *transfer.BchaTxTpl) (string, error) {
	log.Infof("bcha 发送url：%s", srv.cfg.Url+"/v1/bcha/transfer")
	log.Infof("bcha 发送结构：%+v", orderReq)
	data, err := util.PostJsonByAuth(srv.cfg.Url+"/v1/bcha/transfer", srv.cfg.User, srv.cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("bcha 发送返回：%s", string(data))
	result, err := transfer.DecodeTransferHotResp(data)
	if err != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,error:%s", orderReq.OrderId, err.Error())
	}
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OrderId)
	}
	if result.Code != 0 || result.Txid == "" {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OrderId)
	}
	txid := result.Txid
	//冻结utxo
	for _, v := range orderReq.TxIns {
		dao.FcTransPushFreezeUtxo(v.FromTxid, int(v.FromIndex), v.FromAddr)
	}
	return txid, nil
}
