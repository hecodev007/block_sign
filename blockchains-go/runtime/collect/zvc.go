package collect

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
	"github.com/shopspring/decimal"
	"time"
)

//todo 需要同时适应脚本启动以及代码任务启动模式

//账户模型
type ZvcCollect struct {
}

//依赖数据库的方式自动归集
func (c *ZvcCollect) CollectByDB() error {

	//查询数据,每次查询10条
	results, err := dao.FcAddressAmountFindCollectAddr(2, "zvc", conf.Cfg.Collect.Zvc.String(), 10)
	if err != nil {
		log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}
	for _, v := range results {
		//对比数据
		amount, _ := decimal.NewFromString(v.Amount)
		forzenAmount, _ := decimal.NewFromString(v.ForzenAmount)
		fee := decimal.NewFromFloat(0.1000000)
		collectAmount := amount.Sub(forzenAmount).Sub(fee)
		if collectAmount.LessThanOrEqual(decimal.Zero) {
			//冻结金额无需归集
			log.Infof("无需归集金额，collectAmount:%s", collectAmount.String())
			continue
		}
		orderReq, err := c.buildOrder(collectAmount, fee, v)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		data, _ := json.Marshal(orderReq)

		timeNow := time.Now().Unix()
		//查询内容写入归集表
		//todo 暂无代币币种测试 主链币名用MainCoinName
		collectData := &entity.FcCollect{
			MainCoinName:    v.CoinType,
			AppId:           v.AppId,
			CreateTime:      timeNow,
			UpdateTime:      timeNow,
			CoinName:        v.CoinType,
			ContractOrToken: "",
			FeeCoinName:     v.CoinType,
			Txid:            "",
			ToAddr:          orderReq.ToAddress,
			FromAddr:        orderReq.FromAddress,
			ChangeAddr:      "",
			ToAmount:        collectAmount.String(),
			ToFee:           fee.String(),
			Status:          0,
			SendData:        string(data),
			ErrData:         "",
			OutOrderNo:      orderReq.OuterOrderNo,
			Memo:            orderReq.Memo,
		}
		cid, err := dao.FcCollectInsert(collectData)
		if err != nil {
			log.Errorf("归集插入数据失败，%s", err.Error())
			continue
		}

		//更新增加冻结金额
		lockAmount := collectAmount.Add(fee)

		lockAmount = decimal.Zero //暂时不冻结

		err = dao.FcAddressAmountUpdateAddForzenAmount(v.Address, v.CoinType, v.AppId, lockAmount)
		if err != nil {
			log.Errorf("address :%s,归集冻结金额失败，%s", v.Address, err.Error())
			dao.FcCollectUpdateFailOrder(cid, err.Error())
			continue
		}

		//直接发起交易
		txid, err := c.transfer(collectData)
		if err != nil {
			log.Errorf("zvc归集交易失败，%s", err.Error())
			//更新减少冻结金额
			err2 := dao.FcAddressAmountUpdateSubForzenAmount(v.Address, v.CoinType, v.AppId, lockAmount)
			errStr := err.Error()
			if err2 != nil {
				errStr = errStr + " | " + err2.Error()
				log.Errorf("zvc归集交易失败,减少冻结金额异常，%s", errStr)
			}
			err3 := dao.FcCollectUpdateFailOrder(cid, errStr)
			if err3 != nil {
				errStr = errStr + " | " + err3.Error()
				log.Errorf("zvc归集交易失败,设置失败的归集记录异常，%s", errStr)
			}
			continue
		}
		err2 := dao.FcCollectUpdateStateSuccess(v.Id, txid)
		if err2 != nil {
			log.Errorf("交易成功，但是数据库更新失败，id:%d,txid:%s", v.Id, txid)
		}
		log.Infof("address：%s,归集交易成功，txid:%s", v.Address, txid)
	}
	return nil
}

func (c *ZvcCollect) transfer(fc *entity.FcCollect) (txid string, err error) {
	orderReq := new(transfer.ZvcOrderRequest)
	err = json.Unmarshal([]byte(fc.SendData), orderReq)
	if err != nil {
		return "", err
	}
	txid, err = c.walletServerCreate(orderReq)
	if err != nil {
		//修改为失败
		return "", err
	}
	if txid == "" {
		return "", errors.New("empty txid")
	}
	return txid, nil
}

//=======================私有方法==========================
func (c *ZvcCollect) buildOrder(toAmount, fee decimal.Decimal, fa *entity.FcAddressAmount) (*transfer.ZvcOrderRequest, error) {
	//账户模型没有找零
	//私有方法 构建cocos订单
	var (
		fromAddr string
		toAddr   string
		//changeAddr string
	)
	// 查找from地址和金额
	fromAddr = fa.Address

	////查询这个币种的找零地址
	//changeAddrs, err := dao.FcGenerateAddressListFindChangeAddr(int(fa.AppId), fa.CoinType)
	//if err != nil {
	//	return nil, fmt.Errorf("商户归集异常,无法查询找零地址，商户：%d,coinName:%s,err:%s", fa.AppId, fa.CoinType, err.Error())
	//}
	//随机选一个
	//index := util.RandInt64(0, int64(len(changeAddrs)))
	//changeAddr = changeAddrs[index]

	//查询这个币种的归集地址
	toAddrs, err := dao.FcGenerateAddressListFindAddresses(1, 2, int(fa.AppId), fa.CoinType)
	if err != nil {
		return nil, fmt.Errorf("商户归集异常,无法查询归集地址，商户：%d,coinName:%s,err:%s", fa.AppId, fa.CoinType, err.Error())

	}
	//随机选一个
	index := util.RandInt64(0, int64(len(toAddrs)))
	toAddr = toAddrs[index]

	outerOrderNo := util.GetUUID()
	//填充参数
	orderReq := &transfer.ZvcOrderRequest{}
	orderReq.ApplyId = -1
	orderReq.OuterOrderNo = outerOrderNo
	orderReq.OrderNo = outerOrderNo
	orderReq.MchName = "robot"
	orderReq.FromAddress = fromAddr
	orderReq.CoinName = fa.CoinType
	orderReq.ToAddress = toAddr
	orderReq.Memo = outerOrderNo
	orderReq.ToAmount = toAmount
	return orderReq, nil
}

//创建交易接口参数
func (c *ZvcCollect) walletServerCreate(orderReq *transfer.ZvcOrderRequest) (txid string, err error) {
	dd, _ := json.Marshal(orderReq)
	log.Infof("zvc 交易发送内容 :%s", string(dd))
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[orderReq.CoinName].Url+"/v1/zvc/Transfer", conf.Cfg.HotServers[orderReq.CoinName].User, conf.Cfg.HotServers[orderReq.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("zvc 交易返回内容 :%s", string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败:%s，outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常,%s，outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	txid = fmt.Sprintf("%v", result.Data)
	return txid, nil
}
