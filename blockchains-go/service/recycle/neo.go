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
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

type NeoRecycleService struct {
	CoinName string
}

func NewNeoRecycleService() service.RecycleService {
	return &NeoRecycleService{CoinName: "oneo"}
}

//params model : 0小额合并 1大额合并
func (b *NeoRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		//addrAms []*entity.FcAddressAmount
		//scanNum int
		txid string //币种数据服务
		//addrs   = make([]string, 0) //utxo地址
		utxos []transfer.NeoUtxo //utxo
		//临时估算手续费
		toAmount decimal.Decimal                //to金额
		txIns    = make([]transfer.NeoTxIn, 0)  //utxo模板
		txOuts   = make([]transfer.NeoTxOut, 0) //utxo模板
	)

	//if conf.Cfg.UtxoScan.Num <= 0 {
	//	scanNum = 5
	//} else {
	//	scanNum = conf.Cfg.UtxoScan.Num
	//}
	//step1：to地址
	if toAddr == "" {
		return "", errors.New("缺少to地址")
	}
	//step2：判断模式，小的合并还是大的合并，查询相关地址、只是处理同地址合并即可，from即为utxo发送方
	if model == 0 {
		//小金额回收
		//addrAms, err = dao.FcAddressAmountFindTransfer(reqHead.MchId, reqHead.CoinName, scanNum, "asc")
	} else {
		//大金额回收
		//addrAms, err = dao.FcAddressAmountFindTransfer(reqHead.MchId, reqHead.CoinName, scanNum, "desc")
	}
	//for _, v := range addrAms {
	//	addrs = append(addrs, v.Address)
	//}
	//var addresses []string
	//for _, addr := range addrAms {
	//	addresses = append(addresses, addr.Address)
	//}
	var addresses []string
	if reqHead.RecycleAddress != "" {
		addresses = append(addresses, reqHead.RecycleAddress)
	} else {
		addresses = append(addresses, toAddr)
	}
	utxos, err = b.getUtxoData(addresses, reqHead.CoinName, 5)
	if err != nil {
		return "", fmt.Errorf("%s get utxo error: %v", reqHead.CoinName, err)
	}
	coin := global.CoinDecimal[strings.ToLower(reqHead.CoinName)]
	txIns, toAmount, err = b.getTxInAndTxOut(utxos, reqHead.CoinName, int32(coin.Decimal))
	if err != nil {
		return "", fmt.Errorf("%s get txIns error: %v", reqHead.CoinName, err)
	}
	if len(txIns) == 0 {
		return "", fmt.Errorf("%s txIns is null", toAddr)
	}
	//构建txOut
	txOuts = append(txOuts, transfer.NeoTxOut{
		ToAddr:   toAddr,
		ToAmount: toAmount.Shift(int32(coin.Decimal)).IntPart(),
	})
	neoReq := new(transfer.NeoOrderRequest)
	neoReq.OrderRequestHead = *reqHead
	neoReq.TxIns = txIns
	neoReq.TxOuts = txOuts
	createData, _ := json.Marshal(neoReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      int(reqHead.ApplyId),
		ApplyCoinId:  int(reqHead.ApplyCoinId),
		OuterOrderNo: reqHead.OuterOrderNo,
		OrderNo:      reqHead.OrderNo,
		MchName:      reqHead.MchName,
		CoinName:     reqHead.CoinName,
		FromAddress:  "",
		ToAddress:    toAddr,
		Amount:       toAmount.Shift(int32(coin.Decimal)).IntPart(), //转换整型
		Quantity:     toAmount.String(),
		Decimal:      int64(coin.Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}

	txid, err = b.walletServerCreateHot(neoReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", reqHead.ApplyId, err.Error())
		return "", fmt.Errorf("%s send error: %v", reqHead.CoinName, err)
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
	return txid, nil
}

func (b *NeoRecycleService) walletServerCreateHot(orderReq *transfer.NeoOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[b.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", b.CoinName)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", b.CoinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, b.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("%s 交易返回内容 :%s", b.CoinName, string(data))
	result, err := transfer.DecodeTransferHotResp(data)

	if err != nil || result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,err: %v", orderReq.OuterOrderNo, err)
	}
	if result.Code != 0 || result.Txid == "" {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result.Txid, nil
}

func (b *NeoRecycleService) getUtxoData(addresses []string, coinName string, limit int) ([]transfer.NeoUtxo, error) {
	cfg, ok := conf.Cfg.HotServers[b.CoinName]
	if !ok {
		return nil, fmt.Errorf("don't find %s config", b.CoinName)
	}
	var utxos []transfer.NeoUtxo
	for _, addr := range addresses {
		params := make(map[string]interface{})
		params["addr"] = addr
		params["num"] = limit
		params["coin_name"] = strings.ToUpper(coinName)
		data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/get_utxos", cfg.Url, b.CoinName), cfg.User, cfg.Password, params)
		if err != nil {
			return nil, fmt.Errorf("rpc get_utxos error: %v", err)
		}
		var utxo transfer.NeoUtxo
		err = json.Unmarshal(data, &utxo)
		if err != nil {
			return nil, fmt.Errorf("json unmarshal utxos error: %v", err)
		}
		utxos = append(utxos, utxo)
	}

	return utxos, nil
}

func (b *NeoRecycleService) getTxInAndTxOut(utxos []transfer.NeoUtxo, coinType string, coinDecimal int32) (txIns []transfer.NeoTxIn, utxoAmount decimal.Decimal, err error) {
	utxoAmount = decimal.Zero
	if len(utxos) == 0 {
		return nil, utxoAmount, errors.New("utxos is nil ptr")
	}
	var tmpInx []transfer.NeoTxIn
	for _, utxo := range utxos {
		if len(utxo.Balance) == 0 {
			continue
		}
		for _, u := range utxo.Balance {
			if strings.ToUpper(coinType) == strings.ToUpper(u.AssetSymbol) {
				for _, unspent := range u.Unspent {
					uAmount, _ := decimal.NewFromString(unspent.Value)
					//utxoAmount = utxoAmount.Add(uAmount)
					tmpInx = append(tmpInx, transfer.NeoTxIn{
						FromAddr:   utxo.Address,
						FromIndex:  unspent.N,
						FromAmount: uAmount.Shift(coinDecimal).IntPart(),
						FromTxid:   unspent.Txid,
					})
					//utxoAmount = utxoAmount.Add(uAmount)
					//num++
					//if num==5 {
					//	break
					//}
				}
			}
		}
	}
	if len(tmpInx) == 0 {
		return nil, utxoAmount, fmt.Errorf("do not find %s txIns", b.CoinName)
	}
	//排序
	for i := 0; i < len(tmpInx)-1; i++ {
		for j := i + 1; j < len(tmpInx); j++ {
			if tmpInx[i].FromAmount < tmpInx[j].FromAmount {
				tmpInx[i], tmpInx[j] = tmpInx[j], tmpInx[i]
			}
		}
	}
	//塞选
	for i, in := range tmpInx {
		if i >= 3 {
			break
		}
		uAmount := decimal.NewFromInt(in.FromAmount).Shift(-coinDecimal)
		utxoAmount = utxoAmount.Add(uAmount)
		txIns = append(txIns, in)
	}
	return
}
