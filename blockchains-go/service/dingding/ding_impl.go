package dingding

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/shopspring/decimal"
	"reflect"
	"strings"
)

type IDingService interface {
	TransferFee(feeAddr, toAddr string, appId int64, feeApply *entity.FcTransfersApply, fee decimal.Decimal) error
	CollectToken(name, to string, mch *entity.FcMch, fromAddresses []string, tokenCoinSet *entity.FcCoinSet) error
	FindCoinFee(mainName, address string, mch *entity.FcMch) (chainAmount string, err error)
}

type BaseDingService struct {
}

func newBDService() *BaseDingService {
	bds := new(BaseDingService)
	return bds
}

func GetIDingService(coinName string) IDingService {
	bds := newBDService()
	if strings.ToLower(coinName) == "heco" {
		coinName = "heco"
	}
	// 把coin那么的首个字符串变成大写
	coinName = strings.Replace(coinName, string(coinName[0]), strings.ToUpper(string(coinName[0])), 1)
	coinDingName := fmt.Sprintf("New%sDingService", coinName)
	return reflect.ValueOf(bds).MethodByName(coinDingName).Call(nil)[0].Interface().(IDingService)
}
