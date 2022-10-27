package recycle

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type BsvRecycleService struct {
	CoinName string
}

func NewBsvRecycleService() service.RecycleService {
	return &BsvRecycleService{CoinName: "bsv"}
}

func (b *BsvRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {

	//开始组装bsv发送结构
	orderReq := &transfer.BsvOrderRequest{
		OrderRequestHead: *reqHead,
		OrderAddress: []*transfer.BsvOrderAddressRequest{
			&transfer.BsvOrderAddressRequest{
				Address: toAddr,
				Amount:  0,
			},
		},
	}
	//默认小金额回收
	url := conf.Cfg.Walletserver.Url + "/bsv/collect/little"
	if model == 1 {
		//大金额回收
		url = conf.Cfg.Walletserver.Url + "/bsv/collect/big"
	}
	log.Infof("零散回收url:%s", url)
	data, err := util.PostJsonByAuth(url, conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("返回data:%s", string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("bsv 回收请求下单接口失败，模式：%d，outOrderId：%s,data:%s", model, orderReq.OuterOrderNo, string(data))
	}
	if result.Code != 0 || result.Data == nil {
		return "", fmt.Errorf("order表 回收请求下单接口返回值失败,m欧式:%d,服务器返回异常，data:%s,outOrderId：%s", model, string(data), orderReq.OuterOrderNo)
	}
	return "bsv零散归集发送成功", nil

}
