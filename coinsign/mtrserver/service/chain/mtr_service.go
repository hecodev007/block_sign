package chain

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/mtrserver/model"
	"github.com/group-coldwallet/mtrserver/model/bo"
	"github.com/group-coldwallet/mtrserver/model/vo"
	"github.com/group-coldwallet/mtrserver/pkg/mtrutil"
	"github.com/group-coldwallet/mtrserver/pkg/util"
	"github.com/group-coldwallet/mtrserver/service"
	"github.com/sirupsen/logrus"
	"strings"
)

type MtrService struct {
}

func NewMtrService() service.ChainService {
	return &MtrService{}
}

func (u *MtrService) SignTx(tpl *bo.TxTpl) (hex string, err error) {

	//简单校验参数
	err = tpl.Check()
	if err != nil {
		return "", err
	}

	hex, err = tpl.SignTxTpl()
	if err != nil {
		err = fmt.Errorf("sign error:%v", err.Error())
		logrus.Error(err.Error())
		return "", err
	}
	return hex, nil

}

func (u *MtrService) CreateAddr(params *bo.CreateAddrParam, createPath string) (*vo.CreateAddrResult, error) {
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
		act, err := mtrutil.GenerateAccount()
		if err != nil {
			logrus.Infof("createn new address error,numbers: %d, error:%v", i, err)
			return nil, fmt.Errorf("createn new address error,numbers: %d, error:%v", i, err)
		}
		addrs = append(addrs, util.AddrInfo{
			Address: strings.ToLower(act.Address.String()),
			PrivKey: act.PrivateKeyStr,
		})
	}

	returnAddrs, err := util.CreateAddrCsv(createPath, params.MchId, params.OrderId, params.OrderId, addrs, 42, 66)
	if err != nil {
		return nil, errors.New("write csv  error")
	}
	if len(returnAddrs) < params.Num {
		return nil, errors.New("write csv number  error")
	}
	result.Addrs = returnAddrs
	return result, nil
}
