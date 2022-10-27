package recycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"sort"
)

type LtcRecycleService struct {
	CoinName string
}

func NewLtcRecycleService() service.RecycleService {
	return &LtcRecycleService{CoinName: "ltc"}
}

//params model : 0小额合并 1大额合并
func (b *LtcRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		addrAms         []*entity.FcAddressAmount
		scanNum         int
		ltcCfg          *conf.CoinServers                          //币种数据服务
		addrs           = make([]string, 0)                        //utxo地址
		unspents        *transfer.LtcUnspents                      //utxo
		feeTmp          int64                                      //临时估算手续费
		fromAmountInt64 decimal.Decimal                            //from金额
		toAmountInt64   decimal.Decimal                            //to金额
		sortUtxoDesc    transfer.LtcUnspentDesc                    //大额
		sortUtxoAsc     transfer.LtcUnspentAsc                     //小额
		ltcOrderAddrs   = make([]*transfer.LtcOrderAddrRequest, 0) //输入输出模板
		tpl             *transfer.LtcOrderRequest                  //模板
	)

	if conf.Cfg.UtxoScan.Num <= 0 || conf.Cfg.UtxoScan.Num > 12 {
		scanNum = 6
	} else {
		scanNum = conf.Cfg.UtxoScan.Num
	}
	//step1：to地址
	if toAddr == "" {
		return "", errors.New("缺少to地址")
	}
	//step2：判断模式，小的合并还是大的合并，查询相关地址
	if model == 0 {
		//小金额回收
		addrAms, err = dao.FcAddressAmountFindTransfer(reqHead.MchId, reqHead.CoinName, scanNum, "asc")
	} else {
		//大金额回收
		addrAms, err = dao.FcAddressAmountFindTransfer(reqHead.MchId, reqHead.CoinName, scanNum, "desc")
	}
	for _, v := range addrAms {
		addrs = append(addrs, v.Address)
	}

	//step3：获取utxo，获取前面15个utxo
	ltcCfg = conf.Cfg.CoinServers[reqHead.CoinName]
	if ltcCfg == nil {
		return "", errors.New("配置文件缺少ltc coinservers设置")
	}
	byteData, err := util.PostJson(ltcCfg.Url+"/api/v1/ltc/unspents", addrs)
	if err != nil {
		return "", fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.LtcUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return "", fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 || len(unspents.Data) == 0 {
		return "", errors.New("ltc empty unspents")
	}

	//排序unspent，先进行降序，找出大额的数值
	if model == 0 {
		sortUtxoAsc = append(sortUtxoAsc, unspents.Data...)
		sort.Sort(sortUtxoAsc)
		for i, v := range sortUtxoAsc {
			if v.Confirmations == 0 {
				continue
			}
			if i == scanNum {
				break
			}
			txin := &transfer.LtcOrderAddrRequest{
				Dir:     transfer.DirTypeFrom,
				Address: v.Address,
				TxID:    v.Txid,
				Vout:    v.Vout,
				Amount:  v.Amount,
			}
			fromAmountInt64 = fromAmountInt64.Add(decimal.New(v.Amount, 0))
			ltcOrderAddrs = append(ltcOrderAddrs, txin)
		}

	} else {
		sortUtxoDesc = append(sortUtxoDesc, unspents.Data...)
		sort.Sort(sortUtxoDesc)
		for i, v := range sortUtxoDesc {
			if v.Confirmations == 0 {
				continue
			}
			if i == scanNum {
				break
			}
			txin := &transfer.LtcOrderAddrRequest{
				Dir:     transfer.DirTypeFrom,
				Address: v.Address,
				TxID:    v.Txid,
				Vout:    v.Vout,
				Amount:  v.Amount,
			}
			fromAmountInt64 = fromAmountInt64.Add(decimal.New(v.Amount, 0))
			ltcOrderAddrs = append(ltcOrderAddrs, txin)
		}
	}
	//手续计算
	feeTmp, err = getLtcFee(len(ltcOrderAddrs), 1)
	if !feeFloat.IsZero() {
		feeTmp = feeFloat.Shift(8).IntPart()
	}
	//step4：组装交易发送给冷签名端
	toAmountInt64 = fromAmountInt64.Sub(decimal.New(feeTmp, 0))
	ltcOrderAddrs = append(ltcOrderAddrs, &transfer.LtcOrderAddrRequest{
		Dir:     transfer.DirTypeTo,
		Address: toAddr,
		Amount:  toAmountInt64.IntPart(),
	})

	tpl = &transfer.LtcOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      reqHead.ApplyId,
			ApplyCoinId:  reqHead.ApplyCoinId,
			OuterOrderNo: reqHead.OuterOrderNo,
			OrderNo:      reqHead.OrderNo,
			MchId:        reqHead.MchId,
			MchName:      reqHead.MchName,
			CoinName:     reqHead.CoinName,
			Worker:       reqHead.Worker,
		},
		Fee:          feeTmp,
		OrderAddress: ltcOrderAddrs,
	}

	err = b.walletServerCreate(tpl)
	if err != nil {
		return "", fmt.Errorf("ltc 零散回收失败，模式：%d，err:%s", model, err.Error())
	}
	return fmt.Sprintf("ltc 零散合并成功，模式%d，outOrderId:%s", model, reqHead.OuterOrderNo), nil
}

func (srv *LtcRecycleService) walletServerCreate(orderReq *transfer.LtcOrderRequest) error {
	params, _ := json.Marshal(orderReq)
	log.Infof("发送内容：%s", string(params))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/ltc/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,data:%s", orderReq.OuterOrderNo, string(data))
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return nil

}

//手续费计算
func getLtcFee(inNum, outNum int) (int64, error) {

	var (
		rate int64
	)
	//默认费率
	if inNum <= 0 {
		return 0, errors.New(fmt.Sprintf("Error InNum"))
	}
	if outNum <= 0 {
		return 0, errors.New(fmt.Sprintf("Error OutNum"))
	}
	//近似值字节数
	//byteNum := int64(inNum*148 + 34*outNum + 10)
	//提高输出字节，加速出块
	byteNum := int64((inNum)*148 + 34*outNum + 10)
	rate = conf.Cfg.Rate.Ltc
	if rate == 0 {
		rate = 20
	}
	fee := rate * byteNum
	//限定最小值
	if fee < 1000 {
		fee = 1000
	}
	//限制最大
	if fee > 200000 {
		fee = 200000
	}
	return fee, nil
}
