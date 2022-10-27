package walletorder

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/service"
)

type WalletOrderService struct {
}

func (w *WalletOrderService) UpdateColdOrderState(orderId string, status int) error {
	return dao.FcOrderUpdateState(orderId, status)
}

func (w *WalletOrderService) UpdateHotOrderState(orderId string, status int) error {
	return dao.FcOrderHotUpdateState(orderId, status)
}

func (w *WalletOrderService) GetHotOrder(outerOrderNo string) ([]*entity.FcOrderHot, error) {
	return dao.FcOrderHotFindByOutNo(outerOrderNo)
}

func (w *WalletOrderService) GetHotOrderByOrderId(orderId string) (*entity.FcOrderHot, error) {
	return dao.FcOrderHotFindByOrderId(orderId)
}

func (w *WalletOrderService) GetSuccessHotOrder(outerOrderNo string) (*entity.FcOrderHot, error) {
	datas, err := dao.FcOrderHotFindByNoAndStatus(outerOrderNo, status.BroadcastStatus)
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return nil, fmt.Errorf("outerOrderNo:%s,empty data", outerOrderNo)
	}
	return datas[0], nil
}

func (w *WalletOrderService) GetColdOrderByOrderId(orderId string) (*entity.FcOrder, error) {
	return dao.FcOrderFindByOrderId(orderId)
}

func (w *WalletOrderService) GetColdOrder(outerOrderNo string) ([]*entity.FcOrder, error) {
	return dao.FcOrderFindByOutNo(outerOrderNo)
}

func (w *WalletOrderService) GetSuccessColdOrder(outerOrderNo string) (*entity.FcOrder, error) {
	datas, err := dao.FcOrderFindByNoAndStatus(outerOrderNo, status.BroadcastStatus)
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return nil, fmt.Errorf("outerOrderNo:%s,empty data", outerOrderNo)
	}
	return datas[0], nil
}

func NewWalletOrderService() service.WalletOrderService {
	return &WalletOrderService{}
}
