package api

import (
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/group-coldwallet/blockchains-go/service/balance"
	"github.com/group-coldwallet/blockchains-go/service/coin"
	"github.com/group-coldwallet/blockchains-go/service/mch"
	"github.com/group-coldwallet/blockchains-go/service/merge"
	"github.com/group-coldwallet/blockchains-go/service/order"
	"github.com/group-coldwallet/blockchains-go/service/recycle"
	"github.com/group-coldwallet/blockchains-go/service/register"
	"github.com/group-coldwallet/blockchains-go/service/security"
	"github.com/group-coldwallet/blockchains-go/service/tx"
	"github.com/group-coldwallet/blockchains-go/service/walletorder"
)

var (
	CoinService             service.CoinService
	BalanceService          service.BalanceService
	MchService              service.MchService
	TxinfoService           service.TransactionInfoService
	TransferSecurityService service.TransferSecurityService
	OrderService            service.OrderService
	WalletOrderService      service.WalletOrderService
	RegisterService         map[string]service.RegisterService
	MergeService            map[string]service.MergeService
	RecycleService          map[string]service.RecycleService
)

func init() {
	//币种信息服务
	CoinService = coin.NewCoinBaseService()

	//商户余额服务
	BalanceService = balance.NewBalanceBaseService()

	//商户基本信息服务
	MchService = mch.NewMchBaseService()

	//区块交易信息服务
	TxinfoService = tx.NewTxInfoService()

	//交易验证服务
	TransferSecurityService = security.NewSecurityService()

	//商户订单服务
	OrderService = order.NewOrderBaseService()

	WalletOrderService = walletorder.NewWalletOrderService()

	//地址注册服务
	RegisterService = make(map[string]service.RegisterService)
	RegisterService["btc"] = register.NewBtcRegisterService()
	RegisterService["usdt"] = register.NewUsdtRegisterService()
	RegisterService["uca"] = register.NewUcaRegisterService()
	RegisterService["ltc"] = register.NewLtcRegisterService()
	RegisterService["ghost"] = register.NewGhostRegisterService()
	RegisterService["bch"] = register.NewBchRegisterService()
	RegisterService["zec"] = register.NewZecRegisterService()
	RegisterService["doge"] = register.NewDogeRegisterService()
	RegisterService["biw"] = register.NewBiwRegisterService()
	RegisterService["bcha"] = register.NewBchaRegisterService()
	RegisterService["xec"] = register.NewXecRegisterService()
	RegisterService["satcoin"] = register.NewSatRegisterService()
	RegisterService["eac"] = register.NewEacRegisterService()
	//RegisterService["btm"] = register.NewBtmRegisterService()

	//多地址，冷地址合并服务
	MergeService = make(map[string]service.MergeService)
	MergeService["bnb"] = merge.NewBnbMergeService()
	MergeService["cds"] = merge.NewCdsMergeService()
	MergeService["ar"] = merge.NewArMergeService()
	MergeService["ksm"] = merge.NewKsmMergeService()
	MergeService["crab"] = merge.NewCringMergeService()
	MergeService["hnt"] = merge.NewHntMergeService()
	MergeService["vet"] = merge.NewVetMergeService()
	MergeService["celo"] = merge.NewCeloMergeService()
	MergeService["mtr"] = merge.NewMtrMergeService()
	MergeService["fio"] = merge.NewFioMergeService()
	MergeService["dot"] = merge.NewDotMergeService()
	MergeService["sgb-sgb"] = merge.NewSgbMergeService()
	MergeService["kar"] = merge.NewKarMergeService()
	MergeService["dhx"] = merge.NewDhxMergeService()

	//目前针对utxo 零散回收
	RecycleService = make(map[string]service.RecycleService)
	RecycleService["bsv"] = recycle.NewBsvRecycleService()
	RecycleService["ghost"] = recycle.NewGhostRecycleService()
	RecycleService["ltc"] = recycle.NewLtcRecycleService()
	RecycleService["bch"] = recycle.NewBchRecycleService()
	RecycleService["zec"] = recycle.NewZecRecycleService()
	RecycleService["hc"] = recycle.NewHcRecycleService()
	RecycleService["dcr"] = recycle.NewDcrRecycleService()
	RecycleService["doge"] = recycle.NewDogeRecycleService()
	RecycleService["avax"] = recycle.NewAvaxRecycleService()
	RecycleService["dash"] = recycle.NewDashRecycleService()
	RecycleService["biw"] = recycle.NewBiwRecycleService()
	RecycleService["oneo"] = recycle.NewNeoRecycleService()
	RecycleService["bcha"] = recycle.NewBchaRecycleService()
	RecycleService["xec"] = recycle.NewXecRecycleService()
	RecycleService["zen"] = recycle.NewZenRecycleService()
	RecycleService["satcoin"] = recycle.NewSatRecycleService()
	RecycleService["eac"] = recycle.NewEacRecycleService()
	//RecycleService["btm"] = recycle.NewBtmRecycleService()
}
