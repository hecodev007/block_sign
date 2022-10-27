package service

import "github.com/group-coldwallet/blockchains-go/entity"

type WalletOrderService interface {
	//获取指定外部订单ID的订单
	GetColdOrder(outerOrderNo string) ([]*entity.FcOrder, error)

	//获取指定内部订单ID的订单
	GetColdOrderByOrderId(orderId string) (*entity.FcOrder, error)

	//获取广播成功的订单
	GetSuccessColdOrder(outerOrderNo string) (*entity.FcOrder, error)

	//更改order状态
	UpdateColdOrderState(orderId string, status int) error

	//获取指定外部订单ID的订单
	GetHotOrder(outerOrderNo string) ([]*entity.FcOrderHot, error)

	//获取指定内部订单ID的订单
	GetHotOrderByOrderId(orderId string) (*entity.FcOrderHot, error)

	//获取广播成功的订单
	GetSuccessHotOrder(outerOrderNo string) (*entity.FcOrderHot, error)

	//更改order状态
	UpdateHotOrderState(orderId string, status int) error
}
