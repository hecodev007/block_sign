package service

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"strings"
	"time"
)

var HooSvr *HooService

type HooService struct {
	CallBack string
}

func NewHooService() *HooService {
	return &HooService{}
}

func (s *HooService) ApplyTransactionSubmit(request *model.TransferRequest) (int64, error) {
	request.CoinName = strings.ToLower(request.CoinName)
	coinSet, err := dao.FcCoinSetGetByName(request.CoinName, 1)
	if err != nil {
		return httpresp.UnsupportedToken, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.UnsupportedToken), request.CoinName)
	}

	mch, err := dao.FcMchFindByPlatform(request.Sfrom)
	if err != nil {
		return httpresp.SFROM_ERROR, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.SFROM_ERROR), request.Sfrom)
	}

	froms, err := dao.FcGenerateAddressListFindAddresses(1, 2, mch.Id, request.CoinName)
	if err != nil {
		return httpresp.ADR_NONE, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.ADR_NONE), request.Sfrom)
	}

	ta := &entity.FcTransfersApply{
		Username:   "api",
		Applicant:  request.Sfrom,
		AppId:      mch.Id,
		CallBack:   s.CallBack,
		OutOrderid: request.OutOrderId,
		CoinName:   request.CoinName,
		Type:       "cz",
		Memo:       request.Memo,
		Eoskey:     request.TokenName,
		Eostoken:   request.ContractAddress,
		Fee:        request.Fee.String(),
		Status:     1,
		Createtime: time.Now().Unix(),
	}
	if request.IsForce {
		ta.Isforce = 1
	}

	tacTo := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: coinSet.Id,
		Address:     request.ToAddress,
		AddressFlag: "to",
		ToAmount:    request.Amount.String(),
	}

	tacFrom := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: coinSet.Id,
		Address:     froms[0],
		AddressFlag: "from",
	}

	return dao.FcTransfersApplyCreate(ta, []*entity.FcTransfersApplyCoinAddress{tacTo, tacFrom})
}
