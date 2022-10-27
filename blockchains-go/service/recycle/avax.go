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
)

type AvaxRecycleService struct {
	CoinName   string
	DecimalBit int32
}

func NewAvaxRecycleService() service.RecycleService {
	return &AvaxRecycleService{
		CoinName:   "avax",
		DecimalBit: 9,
	}
}

//params model : 0小额合并 1大额合并
func (b *AvaxRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		addrAms         []*entity.FcAddressAmount
		scanNum         int
		avaxCfg         *conf.CoinServers   //币种数据服务
		addrs           = make([]string, 0) //utxo地址
		utxos           = make([]string, 0) //utxo
		feeTmp          int64               //临时估算手续费
		fromAmountFloat decimal.Decimal     //from金额
		toAmountFloat   decimal.Decimal     //to金额
		sortUtxoAsc     util.AvaxUnspentAsc
		sortUtxoDesc    util.AvaxUnspentDesc
		tplUtxos        = make([]string, 0)
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
	avaxCfg = conf.Cfg.CoinServers[reqHead.CoinName]
	if avaxCfg == nil {
		return "", errors.New("配置文件缺少avax coinservers设置")
	}

	//查询utxo数量
	utxos, err = util.AvaxGetUtxos(conf.Cfg.CoinServers[b.CoinName].Url, addrs...)
	if err != nil {
		return "", err
	}

	if feeFloat.IsZero() {
		//手续计算
		feeTmp, _ = util.AvaxGetTxFee(conf.Cfg.CoinServers[b.CoinName].Url)
		if feeTmp == 0 {
			feeTmp = 1000000
		}
		feeFloat = decimal.New(feeTmp, -1*b.DecimalBit)
	}

	//排序unspent，先进行降序，找出大额的数值
	if model == 0 {
		sortUtxoAsc, err = util.ParseUtxosBySortAsc(utxos)
		if err != nil {
			return "", err
		}
		for i, v := range sortUtxoAsc {
			if i == scanNum {
				break
			}
			fromAmountFloat = fromAmountFloat.Add(v.AmountFolat)
			tplUtxos = append(tplUtxos, v.UtxoStr)
		}

	} else {
		sortUtxoDesc, err = util.ParseUtxosBySortDesc(utxos)
		if err != nil {
			return "", err
		}
		for i, v := range sortUtxoDesc {
			if i == scanNum {
				break
			}
			fromAmountFloat = fromAmountFloat.Add(v.AmountFolat)
			tplUtxos = append(tplUtxos, v.UtxoStr)
		}
	}
	if len(tplUtxos) == 0 {
		return "", errors.New("avax empty uxto")
	}

	//step4：组装交易发送给冷签名端
	toAmountFloat = fromAmountFloat.Sub(feeFloat)
	tpl := &transfer.AvaxTxTpl{
		CoinName:   b.CoinName,
		OrderNo:    reqHead.OrderNo,
		MchName:    reqHead.MchName,
		FromAddr:   "",
		ToAddr:     toAddr,
		ChangeAddr: toAddr,
		Amount:     toAmountFloat.Shift(b.DecimalBit).IntPart(),
		Fee:        feeFloat.Shift(b.DecimalBit).IntPart(),
		Utxos:      tplUtxos,
	}

	txid, err := b.walletServerCreate(tpl)
	if err != nil {
		return "", fmt.Errorf("avax 零散回收失败，模式：%d，err:%s", model, err.Error())
	}
	return fmt.Sprintf("avax 零散合并成功，模式%d，txid:%s", model, txid), nil
}

func (srv *AvaxRecycleService) walletServerCreate(orderReq *transfer.AvaxTxTpl) (string, error) {
	log.Infof("avax 发送url：%s", conf.Cfg.HotServers[srv.CoinName].Url+"/v1/avax/transfer")
	dd, _ := json.Marshal(orderReq)
	log.Infof("avax 发送结构：%s", string(dd))
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[srv.CoinName].Url+"/v1/avax/transfer", conf.Cfg.HotServers[srv.CoinName].User, conf.Cfg.HotServers[srv.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("avax 发送返回：%s", string(data))
	result := transfer.DecodRespTranfer(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OrderNo)
	}
	if result.Code != 0 || result.Txid == "" {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OrderNo)
	}
	txid := result.Txid
	return txid, nil
}
