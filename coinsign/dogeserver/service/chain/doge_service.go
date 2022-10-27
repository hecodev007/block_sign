package chain

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/dogeserver/model"
	"github.com/group-coldwallet/dogeserver/model/bo"
	"github.com/group-coldwallet/dogeserver/model/global"
	"github.com/group-coldwallet/dogeserver/model/vo"
	"github.com/group-coldwallet/dogeserver/pkg/dogeutil"
	"github.com/group-coldwallet/dogeserver/pkg/util"
	"github.com/group-coldwallet/dogeserver/service"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type DogeService struct {
}

func NewDogeService() service.ChainService {
	return &DogeService{}
}

func (u *DogeService) SignTx(tpl *dogeutil.DogeTxTpl) (hex string, err error) {
	var (
		fromAmount int64
		toAmount   int64
		fee        int64
	)
	//简单校验参数
	if len(tpl.TxIns) == 0 || len(tpl.TxOuts) == 0 {
		err = fmt.Errorf("empty datas")
		logrus.Error(err.Error())
		return "", err
	}
	//txin校验
	for i, v := range tpl.TxIns {
		if v.FromAddr == "" {
			err = fmt.Errorf("index:%d,empty address", i)
			logrus.Error(err.Error())
			return "", err
		}
		if v.FromIndex < 0 {
			err = fmt.Errorf("index:%d, address:%s, error vout", i, v.FromAddr)
			logrus.Errorf(err.Error())

			return "", err
		}
		if v.FromAmountInt64 < 0 {
			err = fmt.Errorf("index:%d, address:%s, error amount ", i, v.FromAddr)
			logrus.Errorf(err.Error())
			return "", err
		}

		if v.FromTxid == "" {
			err = fmt.Errorf("index:%d, address:%s, error txid", i, v.FromAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		//查询私钥
		privkey, _ := global.GetValue(v.FromAddr)
		if privkey == "" {
			err = fmt.Errorf("index:%d, address:%s, miss privkey", i, v.FromAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		//添加私钥
		tpl.TxIns[i].FromPrivkey = privkey

		fromAmount += v.FromAmountInt64
	}
	//txout校验
	for i, v := range tpl.TxOuts {
		if v.ToAddr == "" {
			err = fmt.Errorf("index:%d,empty toAddress", i)
			logrus.Error(err.Error())
			return "", err
		}
		if v.ToAmountInt64 < 546 {
			err = fmt.Errorf("index:%d, address:%s, error toAmount，min value:546", i, v.ToAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		toAmount += v.ToAmountInt64
	}

	if fromAmount < toAmount {
		err = fmt.Errorf("Wrong output total,fromAmount:%d,toAmount:%d", fromAmount, toAmount)
		logrus.Errorf(err.Error())
		return "", err
	}

	if fromAmount == toAmount {
		err = fmt.Errorf("Must pay the fee,fromAmount:%d,toAmount:%d", fromAmount, toAmount)
		logrus.Errorf(err.Error())
		return "", err
	}

	fee = fromAmount - toAmount
	//下限 0.001
	//BASE_FEE=Decimal("1")
	if fee < 100000000 {
		err = fmt.Errorf("The fee is too low, fee:%s", decimal.NewFromInt(fee).String())
		logrus.Errorf(err.Error())
		return "", err
	}

	//上限 10
	if fee > 1000000000 {
		err = fmt.Errorf("The fee is too high, fee:%s", decimal.NewFromInt(fee).String())
		logrus.Errorf(err.Error())
		return "", err
	}
	hex, err = dogeutil.SignTxTpl(tpl)
	if err != nil {
		err = fmt.Errorf("sign error:%v", err.Error())
		logrus.Error(err.Error())
		return "", err
	}
	return hex, nil

}

func (u *DogeService) CreateAddr(params *bo.CreateAddrParam, createPath string) (*vo.CreateAddrResult, error) {
	result := &vo.CreateAddrResult{
		Num: params.Num,
		MchInfo: model.MchInfo{
			OrderId:  params.OrderId,
			MchId:    params.MchId,
			CoinName: params.CoinName,
		},
	}
	//先批量生成完成，再写入文件
	addrs := make([]util.AddrInfo, 0)
	//生成地址
	for i := 0; i < params.Num; i++ {
		addrStr, privkeyStr, err := dogeutil.CreateAddr()
		if err != nil {
			logrus.Infof("createn new address error,numbers: %d, error:%v", i, err)
			return nil, fmt.Errorf("createn new address error,numbers: %d, error:%v", i, err)
		}
		addrs = append(addrs, util.AddrInfo{
			Address: addrStr,
			PrivKey: privkeyStr,
		})
	}

	returnAddrs, err := util.CreateAddrCsv(createPath, params.MchId, params.OrderId, params.OrderId, addrs)
	if err != nil {
		return nil, errors.New("write csv  error")
	}
	if len(returnAddrs) < params.Num {
		return nil, errors.New("write csv number  error")
	}
	result.Addrs = returnAddrs
	return result, nil
}
