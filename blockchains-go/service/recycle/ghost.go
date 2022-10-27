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

type GhostRecycleService struct {
	CoinName string
}

func NewGhostRecycleService() service.RecycleService {
	return &GhostRecycleService{CoinName: "ghost"}
}

//params model : 0小额合并 1大额合并
func (b *GhostRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		addrAms         []*entity.FcAddressAmount
		scanNum         int
		ghostCfg        *conf.CoinServers                  //币种数据服务
		addrs           = make([]string, 0)                //utxo地址
		unspents        *transfer.GhostUnspents            //utxo
		feeTmp          int64                              //临时估算手续费
		fromAmountInt64 decimal.Decimal                    //from金额
		toAmountInt64   decimal.Decimal                    //to金额
		sortUtxoDesc    transfer.GhostUnspentDesc          //大额
		sortUtxoAsc     transfer.GhostUnspentAsc           //小额
		txins           = make([]*transfer.GhostTxins, 0)  //utxo模板
		txouts          = make([]*transfer.GhostTxOuts, 0) //utxo模板
		tpl             *transfer.GhostOrderRequest        //模板
		changeAddr      string                             //找零地址
	)

	if conf.Cfg.UtxoScan.Num <= 0 {
		scanNum = 15
	} else {
		scanNum = conf.Cfg.UtxoScan.Num
	}
	//step1：to地址
	if toAddr == "" {
		return "", errors.New("缺少to地址")
	}
	changeAddrs, err := dao.FcGenerateAddressListFindChangeAddr(int(reqHead.MchId), reqHead.CoinName)
	if err != nil || len(changeAddrs) == 0 {
		return "", errors.New("缺少changede地址")
	}
	changeAddr = changeAddrs[0]

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
	ghostCfg = conf.Cfg.CoinServers[reqHead.CoinName]
	if ghostCfg == nil {
		return "", errors.New("配置文件缺少ghost coinservers设置")
	}
	byteData, err := util.PostJson(ghostCfg.Url+"/api/v1/ghost/unspents", addrs)
	if err != nil {
		return "", fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.GhostUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return "", fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 || len(unspents.Data) == 0 {
		return "", errors.New("ghost empty unspents")
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
			txin := &transfer.GhostTxins{
				FromAddr:   v.Address,
				FromTxId:   v.Txid,
				FromIndex:  v.Vout,
				FromAmount: v.Amount,
			}
			fromAmountInt64 = fromAmountInt64.Add(decimal.New(v.Amount, 0))
			txins = append(txins, txin)
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
			txin := &transfer.GhostTxins{

				FromAddr:   v.Address,
				FromAmount: v.Amount,
				FromTxId:   v.Txid,
				FromIndex:  v.Vout,
			}
			fromAmountInt64 = fromAmountInt64.Add(decimal.New(v.Amount, 0))
			txins = append(txins, txin)
		}
	}
	//手续计算
	feeTmp, err = getGhostFee(len(txins), 1)
	//step4：组装交易发送给冷签名端
	toAmountInt64 = fromAmountInt64.Sub(decimal.New(feeTmp, 0))
	txouts = append(txouts, &transfer.GhostTxOuts{
		ToAddr:   toAddr,
		ToAmount: toAmountInt64.IntPart(),
	})

	tpl = &transfer.GhostOrderRequest{
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
		ChangeAddr: changeAddr,
		Fee:        feeTmp,
		TxIns:      txins,
		TxOut:      txouts,
	}

	txid, err := b.walletServerCreate(tpl)
	if err != nil {
		return "", fmt.Errorf("ghost 零散回收失败，模式：%d，err:%s", model, err.Error())
	}
	return fmt.Sprintf("ghost 零散合并成功，模式%d，txid:%s", model, txid), nil
}

//创建交易接口参数
func (srv *GhostRecycleService) walletServerCreate(orderReq *transfer.GhostOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[srv.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", srv.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, srv.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 合并回收交易发送内容 :%s", srv.CoinName, string(dd))
	log.Infof("%s 合并回收交易返回内容 :%s", srv.CoinName, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order表 合并回收请求下单接口失败，outOrderId：%s,err:%s", orderReq.OuterOrderNo, err1.Error())
	}
	if thr.Code != 0 || thr.Txid == "" {
		return "", fmt.Errorf("order表 合并回收请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Txid, nil

}

//手续费计算
func getGhostFee(inNum, outNum int) (int64, error) {

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
	rate = conf.Cfg.Rate.Ghost
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
