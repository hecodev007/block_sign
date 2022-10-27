package recycle

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/txscript"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/rediskey"
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

type DcrRecycleService struct {
	CoinName string
}

func NewDcrRecycleService() service.RecycleService {
	return &DcrRecycleService{CoinName: "dcr"}
}

//params model : 0小额合并 1大额合并
func (srv *DcrRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		addrAms []entity.FcTransPush
		scanNum int
		dcrCfg  *conf.CoinServers //币种数据服务
	)

	if conf.Cfg.UtxoScan.Num <= 0 {
		scanNum = 15
	} else {
		scanNum = conf.Cfg.UtxoScan.Num
	}
	//step0 redis lock
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return "", err
	}
	//step1：to地址
	if toAddr == "" {
		return "", errors.New("缺少to地址")
	}
	//step2：判断模式，小的合并还是大的合并，查询相关地址
	if model == 0 {
		//小金额回收
		addrAms, err = dao.FcTransPushFindVaildUtxo2(int(reqHead.MchId), reqHead.CoinName, scanNum, "asc")
	} else {
		//大金额回收
		addrAms, err = dao.FcTransPushFindVaildUtxo2(int(reqHead.MchId), reqHead.CoinName, scanNum, "desc")
	}
	dcrCfg = conf.Cfg.CoinServers[reqHead.CoinName]
	if dcrCfg == nil {
		return "", errors.New("配置文件缺少dcr coinservers设置")
	}
	if len(addrAms) == 0 {
		return "", errors.New("dcr:No utxo available ")
	}
	var fromAmount decimal.Decimal //发送总金额
	var txInTpl = make([]transfer.DcrTxInTpl, 0)
	var txOutTpl = make([]transfer.DcrTxOutTpl, 0)
	var (
		totalNum int
		lockNum  int
		useNum   int
		unUseNum int
	)
	totalNum = len(addrAms)
	for _, v := range addrAms {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DCR_UTXO_LOCK, v.TransactionId, v.TrxN)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			lockNum++
			continue
		}
		if v.Confirmations == 0 {
			unUseNum++
			//暂时过滤
			continue
		}
		famount, _ := decimal.NewFromString(v.Amount)
		from_amount, _ := famount.Float64()
		oar := transfer.DcrTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.TransactionId,
			FromIndex:  uint32(v.TrxN),
			FromAmount: from_amount,
		}
		txInTpl = append(txInTpl, oar)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, reqHead.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.DCR_UTXO_LOCK_SECOND_TIME)
		fromAmount = fromAmount.Add(famount)
		useNum++
	}
	log.Infof("总共查找数目：%d,锁定数目：%d，不可用数目：%d，可用数目：%d", totalNum, lockNum, unUseNum, useNum)
	if feeFloat.LessThanOrEqual(decimal.Zero) {
		var fee int64
		fee, err = srv.getFee(len(txInTpl), 1)
		if err != nil {
			return "", err
		}
		feeFloat = decimal.New(fee, 0).Shift(-8)
	}
	log.Infof("recyle dcr fee is : %s", feeFloat.String())
	if fromAmount.LessThanOrEqual(feeFloat) {
		return "", fmt.Errorf("from amount is less or equal fee amount,From=[%s],Fee=[%s]", fromAmount.String(), feeFloat.String())
	}
	toAmount := fromAmount.Sub(feeFloat)
	to_amount, _ := toAmount.Float64()
	log.Infof("构建dcr receyle tx：from=%s,to=%f,fee=%s", fromAmount.String(), to_amount, feeFloat.String())
	//构建to
	txOutTpl = append(txOutTpl, transfer.DcrTxOutTpl{
		ToAddr:   toAddr,
		ToAmount: to_amount,
	})

	var ctReq transfer.DcrCreateTxReq
	ctReq.Vin = txInTpl
	ctReq.Vout = txOutTpl

	byteData, err := util.PostJson(dcrCfg.Url+"/api/v1/dcr/create", ctReq)
	if err != nil {
		return "", fmt.Errorf("create dcr tx error，err:%s", err.Error())
	}
	ct, errTx := transfer.DecodeCreateTxResp(byteData)
	if errTx != nil {
		return "", fmt.Errorf("decode createrawtransaction error: %v", errTx)
	}
	if ct.Code == 0 && ct.Message == "ok" && ct.Data != nil {
		var sigData transfer.DcrSignReq
		rawTx := ct.Data.(string)
		var addresses []string
		var inputs []*transfer.RawTxInput
		for _, vin := range txInTpl {
			pkScript, err := parseToScript(vin.FromAddr)
			if err != nil {
				return "", fmt.Errorf("parse %s to pkScript error: %v", vin.FromAddr, err)
			}
			addresses = append(addresses, vin.FromAddr)
			inputs = append(inputs, &transfer.RawTxInput{
				Txid:         vin.FromTxid,
				Vout:         vin.FromIndex,
				Tree:         0,
				ScriptPubKey: pkScript,
				RedeemScript: "",
			})
		}
		sigData.RawTx = rawTx
		sigData.Addresses = addresses
		sigData.Inputs = inputs
		//去做签名交易
		orderReq := &transfer.DcrOrderRequest{
			OrderRequestHead: transfer.OrderRequestHead{
				ApplyId:      int64(reqHead.ApplyId),
				OuterOrderNo: reqHead.OuterOrderNo,
				OrderNo:      reqHead.OrderNo,
				MchId:        int64(reqHead.MchId),
				MchName:      reqHead.MchName,
				CoinName:     srv.CoinName,
			},
		}
		orderReq.Data = &sigData
		var txHex string
		txHex, err = srv.walletServerCreateHot(orderReq)
		//发送交易
		if err != nil {
			return "", fmt.Errorf("dcr sign error,err=%v", err)
		}
		if strings.HasPrefix(txHex, "0x") {
			txHex = strings.TrimPrefix(txHex, "0x")
		}
		var mch *entity.FcMch
		mch, err = dao.FcMchFindById(int(reqHead.MchId))
		if err != nil {
			return "", err
		}
		//查询Dcr的coinSet
		coinSet := global.CoinDecimal[srv.CoinName]
		if coinSet == nil {
			return "", fmt.Errorf("缺少币种信息")
		}
		//写入orderHot表
		createData, _ := json.Marshal(orderReq)
		orderHot := &entity.FcOrderHot{
			ApplyId:      int(reqHead.ApplyId),
			ApplyCoinId:  coinSet.Id,
			OuterOrderNo: reqHead.OuterOrderNo,
			OrderNo:      reqHead.OrderNo,
			MchName:      mch.Platform,
			CoinName:     srv.CoinName,
			FromAddress:  "",
			ToAddress:    toAddr,
			Amount:       toAmount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
			Quantity:     toAmount.Shift(int32(coinSet.Decimal)).String(),
			Decimal:      int64(coinSet.Decimal),
			CreateData:   string(createData),
			Status:       int(status.UnknowErrorStatus),
			CreateAt:     time.Now().Unix(),
			UpdateAt:     time.Now().Unix(),
		}
		var txid string
		txid, err = srv.sendRawTransaction(orderReq.MchName, txHex)
		if err != nil {
			orderHot.Status = int(status.BroadcastErrorStatus)
			orderHot.ErrorMsg = err.Error()
			dao.FcOrderHotInsert(orderHot)
			log.Errorf("下单表订单id：%d,获取发送交易异常:%s", reqHead.ApplyId, err.Error())
			// 写入热钱包表，创建失败
			return "", err
		}
		orderHot.TxId = txid
		orderHot.Status = int(status.BroadcastStatus)
		//保存热表
		err = dao.FcOrderHotInsert(orderHot)
		if err != nil {
			//err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
			// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
			log.Error(err.Error())
			// 发送给钉钉
			dingding.ErrTransferDingBot.NotifyStr(err.Error())
		}
		for _, v := range txInTpl {
			err := dao.FcTransPushFreezeUtxo(v.FromTxid, int(v.FromIndex), v.FromAddr)
			if err != nil {
				log.Errorf("Dcr 冻结utxo失败,%+v", v)
			}
		}
		return fmt.Sprintf("dcr 零散合并成功，模式%d，outOrderId:%s,txid: %s ", model, reqHead.OuterOrderNo, txid), nil
	} else {
		return "", fmt.Errorf("create dcr tx error, %v", ct)
	}

}

//手续费计算
func (srv *DcrRecycleService) getFee(inNum, outNum int) (int64, error) {

	var (
		rate int64 = 100
	)

	//默认费率
	if inNum <= 0 {
		return 0, errors.New(fmt.Sprintf("Error InNum"))
	}
	if outNum <= 0 {
		return 0, errors.New(fmt.Sprintf("Error OutNum"))
	}

	byteNum := int64((inNum)*148 + 34*outNum + 10) //相差有点悬殊

	if rate == 0 {
		rate = 100
	}
	fee := rate * byteNum
	//限定最小值
	if fee < 10000 {
		fee = 10000
	}
	//限制最大
	if fee > 10000000 {
		fee = 10000000
	}
	return fee, nil
}

func parseToScript(address string) (string, error) {
	addr, err := dcrutil.DecodeAddress(address)
	if err != nil {
		return "", err
	}
	pks, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(pks), nil
}

//创建交易接口参数
func (srv *DcrRecycleService) walletServerCreateHot(orderReq *transfer.DcrOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[srv.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", srv.CoinName)
	}
	data, err := util.PostJsonByAuth(cfg.Url+"/v1/dcr/sign", cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return result.Data.(string), nil
}

func (srv *DcrRecycleService) sendRawTransaction(mchId, txHex string) (string, error) {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/dcr/send"
	params := map[string]string{
		"hex":   txHex,
		"mchId": mchId,
	}
	data, err := util.PostJson(url, params)
	if err != nil {
		return "", err
	}
	//	解析data
	if len(data) == 0 {
		return "", errors.New("send tx response data is null")
	}
	var res map[string]interface{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}
	ct, errTx := transfer.DecodeCreateTxResp(data)
	if errTx != nil {
		return "", fmt.Errorf("decode sendrawtransaction error: %v", errTx)
	}
	if ct.Code == 0 && ct.Message == "ok" && ct.Data != nil {
		rawTx := ct.Data.(map[string]interface{})
		if rawTx["txid"] == nil {
			return "", errors.New("send tx response txid is null")
		} else {
			return rawTx["txid"].(string), nil
		}
	} else {
		return "", fmt.Errorf("send tx error,Err=[%s]", string(data))
	}
}
