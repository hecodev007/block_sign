package transfer

import (
	"encoding/json"
	"errors"
	_ "errors"
	"fmt"
	_ "fmt"
	"github.com/O3Labs/neo-utils/neoutils/btckey"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	_ "github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type NeoTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewNeoTransferService() service.TransferService {
	return &NeoTransferService{
		CoinName: "oneo",
		Lock:     &sync.Mutex{},
	}
}

func (s *NeoTransferService) VaildAddr(address string) error {
	ver, _, err := btckey.B58checkdecode(address)
	if err != nil {
		return fmt.Errorf("%s b58 decode valid address error: %v", s.CoinName, err)
	}
	if ver != 0x17 {
		return fmt.Errorf("%s prefix is not equal 0x17", s.CoinName)
	}
	return nil
}

func (s *NeoTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.NeoOrderRequest
		createData []byte //构造交易信息

	)

	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coinSet = global.CoinDecimal[coinType]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	toAddrs, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(ta.Id, "to")
	if err != nil {
		return "", err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return "", fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	toAddr := toAddrs[0].Address
	toAddrAmount, _ := decimal.NewFromString(toAddrs[0].ToAmount)
	orderReq, err = s.buildOrderHot(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	createData, _ = json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  "",
		ToAddress:    toAddr,
		Amount:       toAddrAmount.IntPart(), //转换整型
		Quantity:     toAddrAmount.String(),
		Decimal:      int64(global.CoinDecimal[ta.CoinName].Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}

	txid, err = s.walletServerCreateHot(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		// 写入热钱包表，创建失败
		return "", err
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

func (s *NeoTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("do not support cold transfer")
}

func (s *NeoTransferService) walletServerCreateHot(orderReq *transfer.NeoOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
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
func (s *NeoTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.NeoOrderRequest, error) {
	var (
		toAddr   string
		toAmount decimal.Decimal
	)

	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": ta.AppId,
		"coin_name":   ta.CoinName,
	})
	if err != nil {
		return nil, err
	}

	//查询出账地址和金额
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAmount, err = decimal.NewFromString(toAddrs[0].ToAmount)
	if err != nil {
		return nil, err
	}
	coinType := ta.CoinName
	//
	//if ta.Eoskey != "" {
	//	coinType = strings.ToLower(ta.Eoskey)
	//}
	coin := global.CoinDecimal[coinType]
	//accInfos, err := dao.FcAddressAmountFindTransfer(int64(ta.AppId), coinType, 5, "desc")
	//if err != nil || len(accInfos) == 0 {
	//	return nil, fmt.Errorf("find from address error: %v", err)
	//}
	//var addresses []string
	//for _, a := range accInfos {
	//	addresses = append(addresses, a.Address)
	//}
	// 查找冷地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, toAmount.String()).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址，大于 \n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr := fromAddrs[0]
	//根据地址查找utxo
	utxos, err := s.getUtxoData(fromAddr, coinType, 3)
	if err != nil {
		return nil, fmt.Errorf("get utxo error: %v", err)
	}
	txIns, changeAmount, err := s.getTxInAndTxOut(utxos, coinType, int32(coin.Decimal), toAmount)
	if err != nil {
		return nil, fmt.Errorf("get tx_ins error: %v", err)
	}
	var txOuts []transfer.NeoTxOut

	if changeAmount.GreaterThan(decimal.Zero) {
		//查找招零地址
		//查询找零地址
		changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
		if err != nil {
			return nil, fmt.Errorf("查找找零地址错误： %v", err)
		}
		changeAddress := changes[0]
		txOuts = append(txOuts, transfer.NeoTxOut{
			ToAddr:   changeAddress,
			ToAmount: changeAmount.Shift(int32(coin.Decimal)).IntPart(),
		})
	}

	txOuts = append(txOuts, transfer.NeoTxOut{
		ToAmount: toAmount.Shift(int32(coin.Decimal)).IntPart(),
		ToAddr:   toAddr,
	})

	//填充参数
	orderReq := &transfer.NeoOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.TxIns = txIns
	orderReq.TxOuts = txOuts
	return orderReq, nil
}

func (s *NeoTransferService) getUtxoData(address string, coinName string, limit int) (*transfer.NeoUtxo, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return nil, fmt.Errorf("don't find %s config", s.CoinName)
	}

	params := make(map[string]interface{})
	params["addr"] = address
	params["num"] = limit
	params["coin_name"] = strings.ToUpper(coinName)
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/get_utxos", cfg.Url, s.CoinName), cfg.User, cfg.Password, params)
	if err != nil {
		return nil, fmt.Errorf("rpc get_utxos error: %v", err)
	}
	var utxo transfer.NeoUtxo
	err = json.Unmarshal(data, &utxo)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal utxos error: %v", err)
	}
	return &utxo, nil
}

func (s *NeoTransferService) getTxInAndTxOut(utxo *transfer.NeoUtxo, coinType string, coinDecimal int32, amount decimal.Decimal) (txIns []transfer.NeoTxIn, changeAmount decimal.Decimal, err error) {
	changeAmount = decimal.Zero
	//if len(utxos) == 0 {
	//	return nil, changeAmount, errors.New("utxos is nil ptr")
	//}

	utxoAmount := decimal.Zero
	var tmpInx []transfer.NeoTxIn

	if len(utxo.Balance) == 0 {
		return nil, changeAmount, errors.New("utxo is nil ptr")
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
				//if num>=5 {
				//	break
				//}
				//if utxoAmount.GreaterThanOrEqual(amount) {
				//	break
				//}
			}
		}
	}
	if len(tmpInx) == 0 {
		return nil, changeAmount, fmt.Errorf("do not find %s txIns", s.CoinName)
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
	for _, in := range tmpInx {
		uAmount := decimal.NewFromInt(in.FromAmount).Shift(-coinDecimal)
		utxoAmount = utxoAmount.Add(uAmount)
		txIns = append(txIns, in)
		if utxoAmount.GreaterThanOrEqual(amount) {
			break
		}
	}
	if utxoAmount.LessThan(amount) {
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("utxo数量: %d,出账金额：%d,限制utxo金额不足，utxoAmount：%d", len(tmpInx), amount.IntPart(), utxoAmount.IntPart()))
		return nil, changeAmount, fmt.Errorf("amount is not enougth,utxoAmount:%d,transAmount=%d", utxoAmount.IntPart(), amount.IntPart())
	}
	//找零金额
	changeAmount = utxoAmount.Sub(amount)

	return
}
