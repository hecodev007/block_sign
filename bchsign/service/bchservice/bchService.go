package bchservice

import (
	"fmt"
	"github.com/group-coldwallet/bchsign/model"
	"github.com/group-coldwallet/bchsign/model/bo"
	"github.com/group-coldwallet/bchsign/model/global"
	"github.com/group-coldwallet/bchsign/model/vo"
	"github.com/group-coldwallet/bchsign/util/rylink"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
)

type BchService struct {
}

func (bch *BchService) SignTx(tpl *bo.BchTxTpl) (hex string, err error) {
	fmt.Println(fmt.Sprintf("tpl : %+v", tpl))

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
		if v.FromAmount < 0 {
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
		searchAddr := v.FromAddr
		if strings.HasPrefix(v.FromAddr, "bitcoincash:") {
			searchAddr = strings.Replace(searchAddr, "bitcoincash:", "", 1)
		} else {
			searchAddr, err = rylink.ChangeAddressToBch(searchAddr)
			if err != nil {
				return "", err
			}
		}
		privkey, _ := global.GetValue(searchAddr)
		if privkey == "" {
			err = fmt.Errorf("index:%d, address:%s, miss privkey", i, v.FromAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		//添加私钥
		tpl.TxIns[i].FromPrivkey = privkey

		fromAmount += v.FromAmount

	}

	//txout校验
	for i, v := range tpl.TxOuts {
		if v.ToAddr == "" {
			err = fmt.Errorf("index:%d,empty toAddress", i)
			logrus.Error(err.Error())
			return "", err
		}
		if v.ToAmount < 546 {
			err = fmt.Errorf("index:%d, address:%s, error toAmount，min value:546", i, v.ToAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		toAmount += v.ToAmount
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
	//下限0.00001
	if fee < 400 {
		err = fmt.Errorf("The fee is too low, fee:%d", fee)
		logrus.Errorf(err.Error())
		return "", err
	}

	//上限 1
	if fee > 100000000 {
		err = fmt.Errorf("The fee is too high, fee:%d", fee)
		logrus.Errorf(err.Error())
		return "", err
	}
	hex, err = rylink.BchSignTxTpl(tpl)
	if err != nil {
		err = fmt.Errorf("sign error:%v", err.Error())
		logrus.Error(err.Error())
		return "", err
	}
	return hex, nil
}

func (bch *BchService) CreateAddr(params *bo.CreateAddrParam, createPath, readPath string) (*vo.CreateAddrResult, error) {
	createPath = createPath + "/" + params.Mch
	readPath = readPath + "/" + params.Mch
	result := &vo.CreateAddrResult{
		Num: params.Count,
		MchInfo: model.MchInfo{
			MchId:    params.Mch,
		},
		Addrs: map[string]string{},
	}

	//临时存储地址
	resultAddrs := make([]string, 0)
	//先批量生成完成，再写入文件
	addrs := make([]*bo.AddressInfo, 0)

	//生成地址
	for i := 0; i < params.Count; i++ {
		addrStr, transformAddress, privkeyStr, err := rylink.CreateAddress()
		if err != nil {
			log.Printf("createn new address error,numbers: %d, error:%v", i, err)
			return nil, fmt.Errorf("createn new address error,numbers: %d, error:%v", i, err)
		}
		addrs = append(addrs, &bo.AddressInfo{
			Address:          addrStr,
			TransformAddress: transformAddress,
			PrivateKey:       privkeyStr,
		})
	}
	if len(addrs) != params.Count {
		log.Println(fmt.Sprintf("createn address error,len :%d", len(addrs)))
		return nil, fmt.Errorf("createn address error,len :%d", len(addrs))
	}
	for _, info := range addrs {
		resultAddrs = append(resultAddrs, info.Address)
		result.Addrs[info.TransformAddress] = info.PrivateKey
	}
	return result, nil
}
