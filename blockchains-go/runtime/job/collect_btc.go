package job

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
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/shopspring/decimal"
	"sort"
	"time"
)

const (
	coinName             = "btc"
	toAddress            = "3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw"
	DefaultBtcCollectFee = 200000 // 0.002
)

func BtcCollectProcess(mchId int64) {
	CollectManager(mchId, "collect")
}

func BtcMergeProcess(mchId int64) {
	CollectManager(mchId, "cold")
}

func CollectManager(mchId int64, runType string) {
	var (
		err       error
		addrInfos []*entity.FcAddressAmount
	)

	var utxos []transfer.BtcUtxo
	for i := 0; i < 3; i++ {
		log.Infof("开始执行第 %d 页地址", i+1)
		if runType == "cold" {
			log.Infof("本次执行所有冷地址和用户地址>>>>>>>>>>> %s", runType)
			addrInfos, err = dao.FcAddressAmountFindTransferToBtcForMerge(mchId, 100, i*100)
		} else {
			log.Infof("本次执行所有用户地址>>>>>>>>>>> %s", runType)
			addrInfos, err = dao.FcAddressAmountFindTransferToBtcForCollect(mchId, 100, i*100)
		}

		if err != nil {
			log.Errorf("从数据库获取地址失败 %v", err)
			return
		}

		result, err := BtcCollectPickUtxos(addrInfos)
		if err != nil {
			log.Errorf("[BTC自动归集] pickUtxos失败 %v", err)
			continue
		}
		log.Infof("第 %d 页 获取到可用的utxo数量 %d", i+1, len(*result))
		if len(*result) > 0 {
			utxos = append(utxos, *result...)
		}
	}

	log.Infof("本次共需处理 %d 个UTXO", len(utxos))
	if len(utxos) < 10 {
		log.Infof("可用utxo数量小于10，不执行归集操作")
		return
	}

	// 排序
	var sortUtxo transfer.BtcUnspentDesc
	sortUtxo = append(sortUtxo, utxos...)
	sort.Sort(sortUtxo)
	log.Info("UTXO排序完成")

	var collects []transfer.BtcUtxo
	collects = append(collects, sortUtxo...)

	BtcCollect(collects)
}

func BtcCollectPickUtxos(addrInfos []*entity.FcAddressAmount) (*[]transfer.BtcUtxo, error) {
	var (
		unspents *transfer.BtcUnspents
	)
	log.Infof("查询得到可用的地址个数 %d", len(addrInfos))
	if len(addrInfos) == 0 {
		return nil, fmt.Errorf("暂无可用地址出账")
	}

	addrs := make([]string, 0)
	for _, v := range addrInfos {
		addrs = append(addrs, v.Address)
	}
	//查询utxo数量
	byteData, err := util.PostJson("http://192.170.1.176:9999/api/v1/btc/unspents", addrs)
	if err != nil {
		return nil, fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.BtcUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return nil, fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 {
		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
	}
	if len(unspents.Data) == 0 {
		return nil, errors.New("btc empty unspents")
	}

	log.Infof("%d 个地址的UTXO数量共 %d", len(addrInfos), len(unspents.Data))

	var colletUtxos []transfer.BtcUtxo
	for _, datum := range unspents.Data {
		if datum.Amount < 2500 { // 0.000025
			log.Infof("[BTC自动归集] %s [vout %d] 金额 < 0.000025,忽略", datum.Txid, datum.Vout)
			continue
		}
		if datum.Confirmations == 0 {
			continue
		}
		colletUtxos = append(colletUtxos, datum)
	}
	return &colletUtxos, nil
}

func BtcCollect(utxos []transfer.BtcUtxo) error {
	log.Infof("collect utxo数量 %d", len(utxos))
	var collectUtxos []transfer.BtcUtxo
	for i := 0; i < len(utxos); i++ {
		collectUtxos = append(collectUtxos, utxos[i])
		log.Infof("%s 金额=%s 加入UTXO列表", utxos[i].Txid, decimal.NewFromInt(utxos[i].Amount).Shift(-8).String())

		if (i+1)%100 == 0 {
			BtcCollectIndependent(collectUtxos)
			log.Infof("i=%d UTXO数量满100 开始执行归集", i)
			collectUtxos = []transfer.BtcUtxo{}
			break
		}
	}
	return nil
}

func BtcCollectIndependent(utxos []transfer.BtcUtxo) {
	coinSet := global.CoinDecimal[coinName]

	ta := &entity.FcTransfersApply{
		Username:   "api",
		OrderId:    util.GetUUID(),
		Applicant:  "hoo",
		AppId:      1,
		CallBack:   "",
		OutOrderid: fmt.Sprintf("%s-utxo-merge-%d", coinName, time.Now().Unix()),
		CoinName:   coinName,
		Type:       "hb",
		Status:     int(entity.ApplyStatus_Merge),
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
	}

	tacTo := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: coinSet.Id,
		Address:     toAddress,
		AddressFlag: "to",
		ToAmount:    "0",
		Lastmodify:  util.GetChinaTimeNow(),
	}

	orderId, err := dao.FcTransfersApplyCreate(ta, []*entity.FcTransfersApplyCoinAddress{tacTo})
	if err != nil {
		log.Errorf("[BTC自动归集]FcTransfersApplyCreate 失败 %v", err)
		return
	}

	orderReq := &transfer.BtcOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      orderId,
			ApplyCoinId:  int64(coinSet.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      util.GetUUID(),
			MchId:        1,
			MchName:      "hoo",
			CoinName:     coinName,
			Worker:       "",
		},
		Amount: 0,
		Fee:    DefaultBtcCollectFee, // 手续费 0.002
	}

	utxoTpl := make([]*transfer.BtcOrderAddrRequest, 0)
	fromAmountTotal := decimal.Zero
	for _, v := range utxos {
		oar := &transfer.BtcOrderAddrRequest{
			Dir:     transfer.DirTypeFrom,
			Address: v.Address,
			Amount:  v.Amount,
			TxID:    v.Txid,
			Vout:    v.Vout,
		}
		fromAmountTotal = fromAmountTotal.Add(decimal.NewFromInt(v.Amount))
		utxoTpl = append(utxoTpl, oar)
	}
	log.Infof("[BTC自动归集] 本次归集的总金额为 %s", fromAmountTotal.Shift(-8).String())

	transferAmount := fromAmountTotal.Sub(decimal.NewFromInt(orderReq.Fee))
	if transferAmount.Cmp(decimal.NewFromInt(100000)) == -1 {
		log.Errorf("[BTC自动归集] 本次归集的金额(%s) < 0.001，没有必要归集", transferAmount.Shift(-8).String())
		return
	}

	//组装to
	utxoTpl = append(utxoTpl, &transfer.BtcOrderAddrRequest{
		Dir:     transfer.DirTypeTo,
		Address: toAddress,
		Amount:  transferAmount.IntPart(),
	})
	orderReq.OrderAddress = utxoTpl
	orderReq.Amount = transferAmount.IntPart()

	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/btc/collect", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		log.Infof("[BTC自动归集] 调用walletserver失败 %v", err)
		log.Infof("[BTC自动归集] 调用walletserver失败 %s", string(data))
		return
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		log.Errorf("[BTC自动归集] 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
		return
	}
	if result.Code != 0 || result.Data == nil {
		log.Errorf("[BTC自动归集] 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
		return
	}
	msg := fmt.Sprintf("[BTC归集]成功 已将金额 %s 归集到地址 %s", decimal.NewFromInt(orderReq.Amount).Shift(-8).String(), toAddress)
	dingding.ReviewDingBot.NotifyStr(msg)
	log.Info(msg)
}
