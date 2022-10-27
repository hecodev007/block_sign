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

type ZecRecycleService struct {
	CoinName string
}

func NewZecRecycleService() service.RecycleService {
	return &ZecRecycleService{CoinName: "zec"}
}

//params model : 0小额合并 1大额合并
func (srv *ZecRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		addrAms         []*entity.FcAddressAmount
		scanNum         int
		zecCfg          *conf.CoinServers                          //币种数据服务
		addrs           = make([]string, 0)                        //utxo地址
		unspents        *transfer.ZecUnspents                      //utxo
		feeTmp          int64                                      //临时估算手续费
		fromAmountInt64 decimal.Decimal                            //from金额
		toAmountInt64   decimal.Decimal                            //to金额
		sortUtxoDesc    transfer.ZecUnspentDesc                    //大额
		sortUtxoAsc     transfer.ZecUnspentAsc                     //小额
		zecOrderAddrs   = make([]*transfer.ZecOrderAddrRequest, 0) //输入输出模板
		tpl             *transfer.ZecOrderRequest                  //模板
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
	zecCfg = conf.Cfg.CoinServers[reqHead.CoinName]
	if zecCfg == nil {
		return "", errors.New("配置文件缺少zec coinservers设置")
	}
	byteData, err := util.PostJson(zecCfg.Url+"/api/v1/zec/unspents", addrs)
	if err != nil {
		return "", fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.ZecUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return "", fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 || len(unspents.Data) == 0 {
		return "", errors.New("zec empty unspents")
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
			txin := &transfer.ZecOrderAddrRequest{
				Dir:     transfer.DirTypeFrom,
				Address: v.Address,
				TxID:    v.Txid,
				Vout:    v.Vout,
				Amount:  v.Amount,
			}
			fromAmountInt64 = fromAmountInt64.Add(decimal.New(v.Amount, 0))
			zecOrderAddrs = append(zecOrderAddrs, txin)
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
			txin := &transfer.ZecOrderAddrRequest{
				Dir:     transfer.DirTypeFrom,
				Address: v.Address,
				TxID:    v.Txid,
				Vout:    v.Vout,
				Amount:  v.Amount,
			}
			fromAmountInt64 = fromAmountInt64.Add(decimal.New(v.Amount, 0))
			zecOrderAddrs = append(zecOrderAddrs, txin)
		}
	}
	//手续计算
	feeTmp, err = getZecFee(len(zecOrderAddrs), 1)
	if !feeFloat.IsZero() {
		feeTmp = feeFloat.Shift(8).IntPart()
	}
	//step4：组装交易发送给冷签名端
	toAmountInt64 = fromAmountInt64.Sub(decimal.New(feeTmp, 0))
	zecOrderAddrs = append(zecOrderAddrs, &transfer.ZecOrderAddrRequest{
		Dir:     transfer.DirTypeTo,
		Address: toAddr,
		Amount:  toAmountInt64.IntPart(),
	})

	tpl = &transfer.ZecOrderRequest{
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
		OrderAddress: zecOrderAddrs,
	}
	height, err := srv.findExpiryHeight()
	if err != nil {
		return "", err
	}
	tpl.ExpiryHeight = height

	err = srv.walletServerCreate(tpl)
	if err != nil {
		return "", fmt.Errorf("zec 零散回收失败，模式：%d，err:%s", model, err.Error())
	}
	return fmt.Sprintf("zec 零散合并成功，模式%d，outOrderId:%s", model, reqHead.OuterOrderNo), nil
}

func (srv *ZecRecycleService) walletServerCreate(orderReq *transfer.ZecOrderRequest) error {
	params, _ := json.Marshal(orderReq)
	log.Infof("发送内容：%s", string(params))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/zec/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
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
func getZecFee(inNum, outNum int) (int64, error) {

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
	rate = conf.Cfg.Rate.Zec
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

func (srv *ZecRecycleService) findExpiryHeight() (int64, error) {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/zec/chain"
	data, err := util.Get(url)
	if err != nil {
		err = fmt.Errorf("获取回退高度错误，%s, error:%s", srv.CoinName, err.Error())
		return 0, err
	}
	log.Infof("获取回退"+
		""+
		"高度返回结果：%s", string(data))
	zecResp := transfer.DecodeZecHeightResult(data)
	if zecResp != nil && zecResp.Data != nil {
		height := zecResp.Data.Headers
		return height + 10, nil
	}
	err = fmt.Errorf("获取回退高度错误，%s, error:resp data is null", srv.CoinName)
	return 0, err
}
