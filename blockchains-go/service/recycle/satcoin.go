package recycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"time"
)

type SatRecycleService struct {
	CoinName    string
	CoinDecimal int32
}

func NewSatRecycleService() service.RecycleService {
	return &SatRecycleService{CoinName: "satcoin", CoinDecimal: 8}
}

//params model : 0小额合并 1大额合并
func (b *SatRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		addrAms         []*entity.FcAddressAmount
		scanNum         int
		satcoinCfg      *conf.CoinServers                 //币种数据服务
		addrs           = make([]string, 0)               //utxo地址
		unspents        *transfer.SatUnspents             //utxo
		feeTmp          int64                             //临时估算手续费
		fromAmountInt64 decimal.Decimal                   //from金额
		toAmountInt64   decimal.Decimal                   //to金额
		sortUtxoDesc    transfer.SatUnspentDesc           //大额
		sortUtxoAsc     transfer.SatUnspentAsc            //小额
		txins           = make([]transfer.SatTxInTpl, 0)  //utxo模板
		txouts          = make([]transfer.SatTxOutTpl, 0) //utxo模板
		tpl             *transfer.SatTxTpl                //模板
	)

	//if conf.Cfg.UtxoScan.Num <= 0 {
	//	scanNum = 15
	//} else {
	//	scanNum = conf.Cfg.UtxoScan.Num
	//}
	scanNum = 20
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
	satcoinCfg = conf.Cfg.CoinServers[reqHead.CoinName]
	if satcoinCfg == nil {
		return "", errors.New("配置文件缺少satcoin coinservers设置")
	}
	log.Infof("发送地址：%s", satcoinCfg.Url+"/api/v1/satcoin/unspents")
	byteData, err := util.PostJson(satcoinCfg.Url+"/api/v1/satcoin/unspents", addrs)
	if err != nil {
		return "", fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.SatUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return "", fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 || len(unspents.Data) == 0 {
		return "", errors.New("satcoin empty unspents")
	}

	//排序unspent，先进行降序，找出大额的数值
	if model == 0 {
		sortUtxoAsc = append(sortUtxoAsc, unspents.Data...)
		sort.Sort(sortUtxoAsc)
		for i, v := range sortUtxoAsc {

			if i == scanNum {
				break
			}
			txin := transfer.SatTxInTpl{
				FromAddr:   v.Address,
				FromTxid:   v.Txid,
				FromIndex:  uint32(v.Vout),
				FromAmount: v.AmountInt64.IntPart(),
			}
			fromAmountInt64 = fromAmountInt64.Add(v.AmountInt64)
			txins = append(txins, txin)
		}

	} else {
		sortUtxoDesc = append(sortUtxoDesc, unspents.Data...)
		sort.Sort(sortUtxoDesc)
		for i, v := range sortUtxoDesc {
			if i == scanNum {
				break
			}
			txin := transfer.SatTxInTpl{

				FromAddr:   v.Address,
				FromAmount: v.AmountInt64.IntPart(),
				FromTxid:   v.Txid,
				FromIndex:  uint32(v.Vout),
			}
			fromAmountInt64 = fromAmountInt64.Add(v.AmountInt64)
			txins = append(txins, txin)
		}
	}
	//手续计算
	feeTmp, err = b.getSatFee(len(txins), 1)
	//step4：组装交易发送给冷签名端
	toAmountInt64 = fromAmountInt64.Sub(decimal.New(feeTmp, 0))
	txouts = append(txouts, transfer.SatTxOutTpl{
		ToAddr:   toAddr,
		ToAmount: toAmountInt64.IntPart(),
	})

	tpl = &transfer.SatTxTpl{
		MchId:    reqHead.MchName,
		OrderId:  reqHead.OuterOrderNo,
		CoinName: reqHead.CoinName,
		TxIns:    txins,
		TxOuts:   txouts,
	}

	createData, _ := json.Marshal(tpl)
	orderHot := &entity.FcOrderHot{
		ApplyId:      int(reqHead.ApplyId),
		ApplyCoinId:  int(reqHead.ApplyCoinId),
		OuterOrderNo: reqHead.OuterOrderNo,
		OrderNo:      reqHead.OrderNo,
		MchName:      reqHead.MchName,
		CoinName:     reqHead.CoinName,
		FromAddress:  "",
		ToAddress:    toAddr,
		Amount:       toAmountInt64.IntPart(), //转换整型
		Quantity:     toAmountInt64.Shift(-1 * b.CoinDecimal).String(),
		Decimal:      int64(b.CoinDecimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}

	txid, err := b.walletServerCreate(tpl)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		return "", fmt.Errorf("%s 零散回收失败，模式：%d，err:%s", b.CoinName, model, err.Error())
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
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return fmt.Sprintf("%s 零散合并成功，模式%d，txid:%s", b.CoinName, model, txid), nil
}

//创建交易接口参数
func (srv *SatRecycleService) walletServerCreate(orderReq *transfer.SatTxTpl) (string, error) {
	log.Infof("satcoin 发送url：%s", conf.Cfg.HotServers[srv.CoinName].Url+"/v1/"+strings.ToLower(srv.CoinName)+"/transfer")
	log.Infof("satcoin 发送结构：%+v", orderReq)
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[srv.CoinName].Url+"/v1/"+strings.ToLower(srv.CoinName)+"/transfer", conf.Cfg.HotServers[srv.CoinName].User, conf.Cfg.HotServers[srv.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("satcoin 发送返回：%s", string(data))
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

//手续费计算
func (srv *SatRecycleService) getSatFee(inNum, outNum int) (int64, error) {
	var (
		rate int64 = 1000
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
	byteNum := int64((inNum)*148 + 34*outNum + 10) //相差有点悬殊

	if rate == 0 {
		rate = 1000
	}

	fee := rate * byteNum
	//限定最小值
	if fee < 200000 {
		fee = 200000
	}
	//限制最大
	if fee > 150000000 {
		fee = 150000000
	}
	return fee, nil
	return fee, nil
}
